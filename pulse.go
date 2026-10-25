package main

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// sample is one reading of the Go runtime, taken once per second.
type sample struct {
	At         int64   `json:"t"` // unix milliseconds
	Goroutines int     `json:"g"`
	HeapMB     float64 `json:"h"`
	NumGC      uint32  `json:"gc"`
	PauseUs    float64 `json:"p"` // duration of the last GC pause, microseconds
}

const historySize = 90 // 90 s of history: the chart is full as soon as the page opens

// pulse keeps a ring of the last historySize samples. ReadMemStats stops the
// world, so it runs on this single ticker whatever the number of viewers.
var pulse struct {
	mu   sync.RWMutex
	ring []sample
}

// PauseText formats the last GC pause for the page: "61 µs" or "2.3 ms".
func (s sample) PauseText() string {
	if s.PauseUs >= 1000 {
		return fmt.Sprintf("%.1f ms", s.PauseUs/1000)
	}
	return fmt.Sprintf("%.0f µs", s.PauseUs)
}

func readSample() sample {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	s := sample{
		At:         time.Now().UnixMilli(),
		Goroutines: runtime.NumGoroutine(),
		HeapMB:     float64(int(float64(ms.HeapAlloc)/1024/1024*100)) / 100,
		NumGC:      ms.NumGC,
	}
	if ms.NumGC > 0 {
		s.PauseUs = float64(ms.PauseNs[(ms.NumGC+255)%256]) / 1000
	}
	return s
}

func record(s sample) {
	pulse.mu.Lock()
	defer pulse.mu.Unlock()
	pulse.ring = append(pulse.ring, s)
	if len(pulse.ring) > historySize {
		pulse.ring = pulse.ring[len(pulse.ring)-historySize:]
	}
}

// history returns a copy of the ring, oldest first.
func history() []sample {
	pulse.mu.RLock()
	defer pulse.mu.RUnlock()
	return append([]sample(nil), pulse.ring...)
}

func latest() sample {
	pulse.mu.RLock()
	defer pulse.mu.RUnlock()
	if len(pulse.ring) == 0 {
		return readSample()
	}
	return pulse.ring[len(pulse.ring)-1]
}

// startSampler fills the ring every second until stop is closed.
func startSampler(stop <-chan struct{}) {
	record(readSample())
	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				record(readSample())
			}
		}
	}()
}

// Load generators behind the three buttons of the page. Each one is bounded:
// at most one goroutine burst and one heap burst at a time, a fixed heap size,
// and one forced GC per gcCooldown. Concurrent clicks get 429, never more load.
const (
	burstMinLife = 4 * time.Second
	burstMaxLife = 12 * time.Second
	heapBurstMB  = 64
	heapHold     = 8 * time.Second
	gcCooldown   = 2 * time.Second
)

var (
	burstUntil atomic.Int64 // unix ms until which a goroutine burst runs, 0 when idle
	heapUntil  atomic.Int64 // unix ms until which the heap burst is held
	lastGC     atomic.Int64 // unix ms of the last forced GC
)

// startGoroutineBurst starts n goroutines that each sleep 4 to 12 s, so the
// curve climbs at once and comes down in a slope. False when a burst runs.
func startGoroutineBurst(n int) bool {
	until := time.Now().Add(burstMaxLife).UnixMilli()
	if !burstUntil.CompareAndSwap(0, until) {
		return false
	}
	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		life := burstMinLife + rand.N(burstMaxLife-burstMinLife)
		go func() {
			defer wg.Done()
			time.Sleep(life)
		}()
	}
	go func() {
		wg.Wait()
		burstUntil.Store(0)
	}()
	return true
}

// startHeapBurst allocates heapBurstMB in 1 MB slices, holds them for
// heapHold, releases them and runs a GC so the drop shows on the chart.
func startHeapBurst() bool {
	until := time.Now().Add(heapHold).UnixMilli()
	if !heapUntil.CompareAndSwap(0, until) {
		return false
	}
	go func() {
		held := make([][]byte, heapBurstMB)
		for i := range held {
			b := make([]byte, 1<<20)
			for j := 0; j < len(b); j += 4096 {
				b[j] = 1 // touch every page: real resident memory, not just a reservation
			}
			held[i] = b
		}
		time.Sleep(heapHold)
		runtime.KeepAlive(held) // held stays reachable until here, then becomes garbage
		runtime.GC()
		heapUntil.Store(0)
	}()
	return true
}

// forceGC runs one garbage collection, at most once per gcCooldown.
func forceGC() bool {
	now := time.Now().UnixMilli()
	prev := lastGC.Load()
	if now-prev < gcCooldown.Milliseconds() || !lastGC.CompareAndSwap(prev, now) {
		return false
	}
	go runtime.GC()
	return true
}
