// Command demo-go serves a live Go runtime dashboard, built as a Clever Cloud demo.
// Standard library only: the page template and the Clever Brand Kit stylesheet are
// compiled into the binary with embed, so nothing has to ship next to it.
package main

import (
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
	mux.HandleFunc("GET /cc-brand.css", brandCSS)
	return mux
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

// cut returns s truncated to n bytes, or "—" when s is empty.
func cut(s string, n int) string {
	if s == "" {
		return "—"
	}
	if len(s) > n {
		return s[:n]
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
			inst += " · " + p
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := indexTmpl.Execute(w, pageData{
		Hostname: hostname,
		Port:     listenPort(),
		GoVer:    goVersion(),
		Arch:     runtime.GOOS + "/" + runtime.GOARCH,
		CC:       platformInfo(),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "template error: %v\n", err)
	}
}

func statsPage(w http.ResponseWriter, r *http.Request) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(statsResponse{
		Goroutines: runtime.NumGoroutine(),
		HeapMB:     fmt.Sprintf("%.2f", float64(ms.HeapAlloc)/1024/1024),
		GCCycles:   ms.NumGC,
		UptimeSec:  int64(time.Since(startTime).Seconds()),
		Requests:   atomic.LoadInt64(&requestCount),
		GoVersion:  goVersion(),
	})
}

// brandCSS serves the Clever Brand Kit stylesheet straight from the embedded FS.
func brandCSS(w http.ResponseWriter, r *http.Request) {
	b, err := staticFS.ReadFile("static/cc-brand.css")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(b)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
