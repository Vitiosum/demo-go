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
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

//go:embed static
var staticFS embed.FS

var (
	startTime    = time.Now()
	requestCount int64
	indexTmpl    = template.Must(template.ParseFS(staticFS, "static/index.html"))
)

// platform is what Clever Cloud injects into the environment of a running instance.
// Every field falls back to "—" outside the platform (Live == false).
type platform struct {
	Live         bool
	AppName      string
	AppID        string
	Instance     string
	InstanceType string
	Commit       string
	Deployment   string
}

type pageData struct {
	Hostname string
	Port     string
	GoVer    string
	Arch     string
	CC       platform
}

type statsResponse struct {
	Goroutines int    `json:"goroutines"`
	HeapMB     string `json:"heap_mb"`
	GCCycles   uint32 `json:"gc_cycles"`
	UptimeSec  int64  `json:"uptime_sec"`
	Requests   int64  `json:"requests"`
	GoVersion  string `json:"go_version"`
}

func main() {
	// Clever Cloud collects stdout and stderr alike; one timestamped stream is enough.
	log.SetOutput(os.Stdout)

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

// listenPort returns the port injected by Clever Cloud, or 8080 locally.
func listenPort() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}

func goVersion() string {
	return runtime.Version()[2:] // "go1.24.0" → "1.24.0"
}

func env(k string) string { return os.Getenv(k) }

// cut returns s truncated to n runes (never in the middle of a UTF-8 sequence),
// or "—" when s is empty.
func cut(s string, n int) string {
	if s == "" {
		return "—"
	}
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}

// platformInfo reads the variables Clever Cloud sets on each instance.
// Reference: https://www.clever.cloud/developers/doc/reference/reference-environment-variables/
func platformInfo() platform {
	inst := "—"
	if n := env("INSTANCE_NUMBER"); n != "" {
		inst = "#" + n
		if p := env("CC_PRETTY_INSTANCE_NAME"); p != "" {
			inst += " · " + cut(p, 40)
		}
	}
	return platform{
		Live:         env("APP_ID") != "",
		AppName:      cut(env("CC_APP_NAME"), 40),
		AppID:        cut(env("APP_ID"), 40),
		Instance:     inst,
		InstanceType: cut(env("INSTANCE_TYPE"), 20),
		Commit:       cut(env("CC_COMMIT_ID"), 7),
		Deployment:   cut(env("CC_DEPLOYMENT_ID"), 16),
	}
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&requestCount, 1)
	hostname, _ := os.Hostname()
	// Render into a buffer first: a template error must yield a clean 500,
	// not a 200 with a half-written page.
	var buf bytes.Buffer
	err := indexTmpl.Execute(&buf, pageData{
		Hostname: hostname,
		Port:     listenPort(),
		GoVer:    goVersion(),
		Arch:     runtime.GOOS + "/" + runtime.GOARCH,
		CC:       platformInfo(),
	})
	if err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("write error (/): %v", err)
	}
}

// memSnap caches the last runtime.MemStats reading: ReadMemStats stops the world,
// so it is refreshed at most once per second whatever the polling rate.
var memSnap struct {
	mu    sync.Mutex
	at    time.Time
	heap  uint64
	numGC uint32
}

func memStats() (heap uint64, numGC uint32) {
	memSnap.mu.Lock()
	defer memSnap.mu.Unlock()
	if time.Since(memSnap.at) >= time.Second {
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		memSnap.heap, memSnap.numGC, memSnap.at = ms.HeapAlloc, ms.NumGC, time.Now()
	}
	return memSnap.heap, memSnap.numGC
}

func statsPage(w http.ResponseWriter, r *http.Request) {
	heap, numGC := memStats()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	err := json.NewEncoder(w).Encode(statsResponse{
		Goroutines: runtime.NumGoroutine(),
		HeapMB:     fmt.Sprintf("%.2f", float64(heap)/1024/1024),
		GCCycles:   numGC,
		UptimeSec:  int64(time.Since(startTime).Seconds()),
		Requests:   atomic.LoadInt64(&requestCount),
		GoVersion:  goVersion(),
	})
	if err != nil {
		log.Printf("encode error (/stats): %v", err)
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
