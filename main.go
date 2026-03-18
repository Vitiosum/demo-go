package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/", indexPage)
	http.HandleFunc("/health", healthCheck)

	http.ListenAndServe("0.0.0.0:8080", nil)
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	port := os.Getenv("PORT")
	instance := os.Getenv("INSTANCE_NUMBER")
	now := time.Now().Format(time.RFC1123)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
  <title>Go on Clever Cloud</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      min-height: 100vh;
      background: linear-gradient(135deg, #0a0f1e 0%%, #0d1a2e 50%%, #0a0f1e 100%%);
      color: #e2e8f0;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
      padding: 40px 16px 64px;
    }
    .container { max-width: 600px; margin: 0 auto; }
    .header { text-align: center; margin-bottom: 40px; }
    .badge {
      display: inline-flex; align-items: center; gap: 8px;
      background: rgba(56,189,248,0.1); border: 1px solid rgba(56,189,248,0.2);
      border-radius: 99px; color: #7dd3fc; font-size: 13px; font-weight: 500;
      padding: 6px 16px; margin-bottom: 16px;
    }
    .dot { width: 7px; height: 7px; background: #38bdf8; border-radius: 50%; animation: pulse 2s ease infinite; }
    @keyframes pulse { 0%%,100%% { opacity:1; } 50%% { opacity:0.3; } }
    h1 { font-size: 38px; font-weight: 700; color: white; margin-bottom: 8px; }
    h1 .go { color: #00ADD8; }
    h1 .cc { color: #22c55e; }
    .subtitle { color: #64748b; font-size: 15px; }
    .card {
      background: rgba(15,23,42,0.85); backdrop-filter: blur(8px);
      border: 1px solid rgba(255,255,255,0.07); border-radius: 16px;
      padding: 24px; margin-bottom: 14px;
    }
    .card-title { font-size: 11px; font-weight: 600; letter-spacing: 0.08em; text-transform: uppercase; color: #475569; margin-bottom: 16px; }
    .info-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
    .info-item { background: rgba(2,6,23,0.6); border: 1px solid rgba(255,255,255,0.05); border-radius: 10px; padding: 14px; }
    .info-label { font-size: 10px; font-weight: 600; letter-spacing: 0.08em; text-transform: uppercase; color: #475569; margin-bottom: 4px; }
    .info-value { font-size: 14px; color: #e2e8f0; word-break: break-all; font-family: monospace; }
    .info-value.empty { color: #334155; font-style: italic; }
    .cta {
      display: inline-flex; align-items: center; gap: 8px;
      background: #22c55e; border-radius: 10px; color: #052e16;
      font-size: 14px; font-weight: 600; padding: 12px 20px;
      text-decoration: none; transition: all 0.2s;
    }
    .cta:hover { background: #16a34a; transform: translateY(-1px); box-shadow: 0 4px 20px rgba(34,197,94,0.3); }
    @media (max-width: 480px) { .info-grid { grid-template-columns: 1fr; } }
  </style>
</head>
<body>
<div class="container">
  <div class="header">
    <div class="badge"><span class="dot"></span>Live on Clever Cloud</div>
    <h1><span class="go">Go</span> + <span class="cc">Clever Cloud</span></h1>
    <p class="subtitle">A minimal Go web server running in production</p>
  </div>

  <div class="card">
    <div class="card-title">Server Info</div>
    <div class="info-grid">
      <div class="info-item">
        <div class="info-label">Hostname</div>
        <div class="info-value">%s</div>
      </div>
      <div class="info-item">
        <div class="info-label">Instance</div>
        <div class="info-value %s">%s</div>
      </div>
      <div class="info-item">
        <div class="info-label">Port</div>
        <div class="info-value %s">%s</div>
      </div>
      <div class="info-item">
        <div class="info-label">Server Time</div>
        <div class="info-value">%s</div>
      </div>
    </div>
  </div>

  <div class="card">
    <div class="card-title">Deploy your own</div>
    <a class="cta" href="https://www.clever-cloud.com" target="_blank" rel="noopener noreferrer">
      <svg width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6M15 3h6v6M10 14L21 3"/></svg>
      Try Clever Cloud
    </a>
  </div>
</div>
</body>
</html>`,
		hostname,
		emptyClass(instance), emptyVal(instance, "—"),
		emptyClass(port), emptyVal(port, "—"),
		now,
	)
}

func emptyClass(s string) string {
	if s == "" {
		return "empty"
	}
	return ""
}

func emptyVal(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
