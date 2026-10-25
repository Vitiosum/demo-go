// Command demo-go serves a live Go runtime dashboard, built as a Clever Cloud demo.
// Standard library only: the page template and the Clever Brand Kit stylesheet are
// compiled into the binary with embed, so nothing has to ship next to it.
package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

//go:embed static
var staticFS embed.FS

var (
	startTime    = time.Now()
	requestCount atomic.Int64
	indexTmpl    = template.Must(template.ParseFS(staticFS, "static/index.html"))
)

type pageData struct {
	Hostname    string
	GoVer       string
	Arch        string
	MaxProcs    int
	Uptime      string
	Now         sample
	HeapBurstMB int
	CC          platform
	Demo        demoConfig
}

// statsResponse is what the page polls every second on /stats.
type statsResponse struct {
	Now         sample   `json:"now"`
	UptimeSec   int64    `json:"uptime_sec"`
	Requests    int64    `json:"requests"`
	GoVersion   string   `json:"go_version"`
	BurstLeftMs int64    `json:"burst_left_ms"` // > 0 while a goroutine burst runs
	HeapLeftMs  int64    `json:"heap_left_ms"`  // > 0 while the heap burst is held
	History     []sample `json:"history,omitempty"`
}

func main() {
	// Clever Cloud collects stdout and stderr alike; one timestamped stream is enough.
	log.SetOutput(os.Stdout)

	stopSampler := make(chan struct{})
	defer close(stopSampler)
	startSampler(stopSampler)

	port := listenPort()
	srv := &http.Server{
		Addr:              "0.0.0.0:" + port,
		Handler:           newHandler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// SIGTERM is what Clever Cloud sends on redeploy or scale-down: finish the
	// in-flight requests (10 s max) instead of cutting them.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("Go runtime dashboard on :%s", port)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server error: %v", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		stop()
		log.Printf("signal received, shutting down (10 s max)")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
			os.Exit(1)
		}
		log.Printf("server stopped")
	}
}

// newHandler wires the routes. Method patterns (Go 1.22+) answer 405 to other verbs,
// and "/{$}" matches the root path only, so any other path gets a 404.
func newHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", indexPage)
	mux.HandleFunc("GET /health", healthCheck)
	mux.HandleFunc("GET /stats", statsPage)
	mux.HandleFunc("POST /load/goroutines", loadGoroutines)
	mux.HandleFunc("POST /load/heap", loadHeap)
	mux.HandleFunc("POST /load/gc", loadGC)
	mux.HandleFunc("GET /cc-brand.css", staticFile("cc-brand.css", "text/css; charset=utf-8", "public, max-age=86400"))
	mux.HandleFunc("GET /app.js", staticFile("app.js", "text/javascript; charset=utf-8", "public, max-age=3600"))
	return secure(mux)
}

// contentSecurityPolicy allows only what the page needs: same-origin script and
// data, Google Fonts (stylesheet + font files), inline <style> and style="--i:n"
// attributes from the brand kit, data: URIs for the favicon and CSS masks.
const contentSecurityPolicy = "default-src 'self'; script-src 'self'; " +
	"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src https://fonts.gstatic.com; " +
	"img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'"

// secure adds the security headers to every response, 404/405 included.
func secure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Content-Security-Policy", contentSecurityPolicy)
		next.ServeHTTP(w, r)
	})
}

func goVersion() string {
	return strings.TrimPrefix(runtime.Version(), "go") // "go1.27.2" → "1.27.2"
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	requestCount.Add(1)
	hostname, _ := os.Hostname()
	// Render into a buffer first: a template error must yield a clean 500,
	// not a 200 with a half-written page.
	var buf bytes.Buffer
	err := indexTmpl.Execute(&buf, pageData{
		Hostname:    cut(hostname, 40),
		GoVer:       goVersion(),
		Arch:        runtime.GOOS + "/" + runtime.GOARCH,
		MaxProcs:    runtime.GOMAXPROCS(0),
		Uptime:      fmtUptime(time.Since(startTime)),
		Now:         latest(),
		HeapBurstMB: heapBurstMB,
		CC:          loadPlatform(os.Getenv),
		Demo:        loadDemoConfig(os.Getenv),
	})
	if err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("write error (/): %v", err)
	}
}

// fmtUptime matches the format of app.js: "42s", "3m 07s", "2h 05m".
func fmtUptime(d time.Duration) string {
	sec := int(d.Seconds())
	switch {
	case sec < 60:
		return fmt.Sprintf("%ds", sec)
	case sec < 3600:
		return fmt.Sprintf("%dm %02ds", sec/60, sec%60)
	default:
		return fmt.Sprintf("%dh %02dm", sec/3600, sec%3600/60)
	}
}

func leftMs(until int64) int64 {
	if until == 0 {
		return 0
	}
	return max(until-time.Now().UnixMilli(), 1)
}

func statsPage(w http.ResponseWriter, r *http.Request) {
	resp := statsResponse{
		Now:         latest(),
		UptimeSec:   int64(time.Since(startTime).Seconds()),
		Requests:    requestCount.Load(),
		GoVersion:   goVersion(),
		BurstLeftMs: leftMs(burstUntil.Load()),
		HeapLeftMs:  leftMs(heapUntil.Load()),
	}
	if r.URL.Query().Get("history") == "1" {
		resp.History = history()
	}
	writeJSON(w, http.StatusOK, resp)
}

type loadResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// answerLoad turns the result of a bounded load generator into 202 or 429.
func answerLoad(w http.ResponseWriter, started bool, ok, busy string) {
	if !started {
		w.Header().Set("Retry-After", "2")
		writeJSON(w, http.StatusTooManyRequests, loadResponse{false, busy})
		return
	}
	writeJSON(w, http.StatusAccepted, loadResponse{true, ok})
}

func loadGoroutines(w http.ResponseWriter, r *http.Request) {
	n := loadDemoConfig(os.Getenv).Burst
	answerLoad(w, startGoroutineBurst(n),
		fmt.Sprintf("%d goroutines lancées", n), "Une rafale tourne déjà")
}

func loadHeap(w http.ResponseWriter, r *http.Request) {
	answerLoad(w, startHeapBurst(),
		fmt.Sprintf("%d MB alloués pour %d s", heapBurstMB, int(heapHold.Seconds())), "Une allocation est déjà en cours")
}

func loadGC(w http.ResponseWriter, r *http.Request) {
	answerLoad(w, forceGC(), "GC déclenché", "Un GC vient de passer")
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode error: %v", err)
	}
}

// staticFile serves one file of the embedded static/ directory (brand kit
// stylesheet, polling script) with its MIME type and cache policy.
func staticFile(name, contentType, cacheControl string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := staticFS.ReadFile("static/" + name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", cacheControl)
		if _, err := w.Write(b); err != nil {
			log.Printf("write error (/%s): %v", name, err)
		}
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Printf("write error (/health): %v", err)
	}
}
