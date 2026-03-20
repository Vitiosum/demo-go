package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync/atomic"
	"text/template"
	"time"
)

var (
	startTime    = time.Now()
	requestCount int64
	indexTmpl    = template.Must(template.New("index").Parse(indexHTML))
)

type pageData struct {
	Hostname string
	Instance string
	Port     string
	GoVer    string
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/", indexPage)
	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/stats", statsPage)
	fmt.Printf("Go runtime dashboard on :%s\n", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
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
		GoVersion:  runtime.Version()[2:],
	})
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&requestCount, 1)
	hostname, _ := os.Hostname()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	instance := os.Getenv("INSTANCE_NUMBER")
	if instance == "" {
		instance = "—"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	indexTmpl.Execute(w, pageData{
		Hostname: hostname,
		Instance: instance,
		Port:     port,
		GoVer:    runtime.Version()[2:],
	})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Go · Clever Cloud</title>
<style>
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
html{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;background:#050c1a;color:#e2e8f0;min-height:100vh}
body{min-height:100vh;overflow-x:hidden;position:relative}

.orb{position:fixed;border-radius:50%;filter:blur(80px);opacity:0.18;pointer-events:none;animation:orb-drift 12s ease-in-out infinite alternate}
.orb-1{width:600px;height:600px;background:#00ADD8;top:-200px;left:-200px;animation-delay:0s}
.orb-2{width:500px;height:500px;background:#22c55e;bottom:-150px;right:-150px;animation-delay:-6s}
@keyframes orb-drift{0%{transform:translate(0,0) scale(1)}100%{transform:translate(40px,30px) scale(1.08)}}

.container{max-width:900px;margin:0 auto;padding:32px 20px 0;position:relative;z-index:1}

.hero{text-align:center;margin-bottom:40px}
.live-badge{display:inline-flex;align-items:center;gap:8px;background:rgba(0,173,216,0.1);border:1px solid rgba(0,173,216,0.25);border-radius:99px;color:#00ADD8;font-size:12px;font-weight:600;letter-spacing:0.06em;text-transform:uppercase;padding:6px 16px;margin-bottom:20px}
.live-dot{width:7px;height:7px;background:#00ADD8;border-radius:50%;animation:pulse 1.5s ease-in-out infinite}
@keyframes pulse{0%,100%{opacity:1;transform:scale(1)}50%{opacity:0.4;transform:scale(0.8)}}

h1{font-size:clamp(2rem,6vw,3.5rem);font-weight:800;line-height:1.1;margin-bottom:12px;letter-spacing:-0.02em}
.go-word{color:#00ADD8;text-shadow:0 0 40px rgba(0,173,216,0.5)}
.cc-word{color:#22c55e;text-shadow:0 0 40px rgba(34,197,94,0.4)}
.subtitle{color:#64748b;font-size:15px}

.metrics{display:grid;grid-template-columns:repeat(3,1fr);gap:14px;margin-bottom:28px}
@media(max-width:620px){.metrics{grid-template-columns:repeat(2,1fr)}}
@media(max-width:380px){.metrics{grid-template-columns:1fr}}

.card{background:rgba(15,23,42,0.85);backdrop-filter:blur(8px);border:1px solid rgba(255,255,255,0.06);border-radius:16px;padding:20px 18px;position:relative;overflow:hidden;transition:border-color 0.3s}
.card::before{content:'';position:absolute;top:0;left:0;right:0;height:3px;border-radius:16px 16px 0 0;background:var(--accent)}
.card::after{content:'';position:absolute;top:0;left:0;right:0;bottom:0;background:radial-gradient(ellipse at top,var(--glow) 0%,transparent 70%);opacity:0;transition:opacity 0.3s;pointer-events:none}
.card.flash::after{opacity:1;animation:flash-fade 0.5s ease-out forwards}
@keyframes flash-fade{0%{opacity:0.4}100%{opacity:0}}

.card-goroutines{--accent:linear-gradient(90deg,#00ADD8,#0ea5e9);--glow:rgba(0,173,216,0.12)}
.card-heap{--accent:linear-gradient(90deg,#a855f7,#8b5cf6);--glow:rgba(168,85,247,0.10)}
.card-uptime{--accent:linear-gradient(90deg,#22c55e,#16a34a);--glow:rgba(34,197,94,0.10)}
.card-gc{--accent:linear-gradient(90deg,#06b6d4,#0891b2);--glow:rgba(6,182,212,0.10)}
.card-requests{--accent:linear-gradient(90deg,#f59e0b,#d97706);--glow:rgba(245,158,11,0.10)}
.card-version{--accent:linear-gradient(90deg,#475569,#334155);--glow:rgba(71,85,105,0.08)}

.card-label{font-size:10px;font-weight:600;letter-spacing:0.1em;text-transform:uppercase;color:#475569;margin-bottom:12px}
.card-value{font-size:2.2rem;font-weight:300;color:#f8fafc;line-height:1;margin-bottom:4px;font-variant-numeric:tabular-nums}
.card-unit{font-size:11px;color:#475569}

.status-bar{display:flex;align-items:center;justify-content:center;gap:6px;font-size:12px;color:#334155;margin-bottom:28px}
.status-dot{width:6px;height:6px;background:#22c55e;border-radius:50%;animation:pulse 2s ease-in-out infinite}

.info-card{background:rgba(15,23,42,0.7);border:1px solid rgba(255,255,255,0.06);border-radius:16px;padding:20px 24px;margin-bottom:28px}
.info-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:12px}
@media(max-width:500px){.info-grid{grid-template-columns:1fr 1fr}}
.info-item{background:rgba(2,6,23,0.5);border:1px solid rgba(255,255,255,0.05);border-radius:10px;padding:12px}
.info-label{font-size:9px;font-weight:600;letter-spacing:0.1em;text-transform:uppercase;color:#334155;margin-bottom:4px}
.info-value{font-size:13px;color:#94a3b8;font-family:monospace;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}

.cert-banner{display:flex;align-items:center;gap:16px;background:rgba(251,191,36,0.07);border:1px solid rgba(251,191,36,0.25);border-radius:16px;padding:20px 24px;margin-bottom:14px}
.cert-icon{font-size:28px;flex-shrink:0}
.cert-content{flex:1}
.cert-title{font-size:15px;font-weight:600;color:#fbbf24;margin-bottom:4px}
.cert-sub{font-size:13px;color:#78716c;line-height:1.5}
.cert-btn{display:inline-flex;align-items:center;gap:6px;background:#fbbf24;border-radius:10px;color:#1c1917;font-size:13px;font-weight:700;padding:10px 18px;text-decoration:none;transition:all 0.2s;white-space:nowrap;flex-shrink:0}
.cert-btn:hover{background:#f59e0b;transform:translateY(-1px);box-shadow:0 4px 16px rgba(251,191,36,0.3)}
@media(max-width:540px){.cert-banner{flex-direction:column;text-align:center}}

footer{border-top:1px solid rgba(255,255,255,0.06);padding:24px 0 12px;margin-top:0}
.footer-links{display:flex;flex-wrap:wrap;gap:8px;justify-content:center;margin-bottom:14px}
.fl{display:inline-flex;align-items:center;gap:6px;background:rgba(255,255,255,0.04);border:1px solid rgba(255,255,255,0.07);border-radius:8px;color:#64748b;font-size:12px;padding:6px 12px;text-decoration:none;transition:all 0.2s}
.fl:hover{background:rgba(255,255,255,0.08);color:#94a3b8}
.fl-cc{color:#86efac!important;border-color:rgba(34,197,94,0.2)!important;background:rgba(34,197,94,0.06)!important}
.fl-cc:hover{background:rgba(34,197,94,0.12)!important}
.fl-li{color:#93c5fd!important;border-color:rgba(96,165,250,0.2)!important;background:rgba(96,165,250,0.06)!important}
.fl-li:hover{background:rgba(96,165,250,0.12)!important}
.fl-cert{color:#fbbf24!important;border-color:rgba(251,191,36,0.25)!important;background:rgba(251,191,36,0.07)!important}
.fl-cert:hover{background:rgba(251,191,36,0.14)!important}
.footer-copy{text-align:center;color:#1e293b;font-size:11px;padding-bottom:8px}
</style>
</head>
<body>
<div class="orb orb-1"></div>
<div class="orb orb-2"></div>

<div class="container">
  <div class="hero">
    <div class="live-badge"><span class="live-dot"></span>Live runtime</div>
    <h1><span class="go-word">Go</span> + <span class="cc-word">Clever Cloud</span></h1>
    <p class="subtitle">Real-time runtime metrics — refreshed every 2 seconds</p>
  </div>

  <div class="metrics">
    <div class="card card-goroutines" id="c-goroutines">
      <div class="card-label">Goroutines</div>
      <div class="card-value" id="v-goroutines">—</div>
      <div class="card-unit">active</div>
    </div>
    <div class="card card-heap" id="c-heap">
      <div class="card-label">Heap</div>
      <div class="card-value" id="v-heap">—</div>
      <div class="card-unit">MB allocated</div>
    </div>
    <div class="card card-uptime" id="c-uptime">
      <div class="card-label">Uptime</div>
      <div class="card-value" id="v-uptime">—</div>
      <div class="card-unit">seconds</div>
    </div>
    <div class="card card-gc" id="c-gc">
      <div class="card-label">GC Cycles</div>
      <div class="card-value" id="v-gc">—</div>
      <div class="card-unit">collections</div>
    </div>
    <div class="card card-requests" id="c-requests">
      <div class="card-label">Requests</div>
      <div class="card-value" id="v-requests">—</div>
      <div class="card-unit">served</div>
    </div>
    <div class="card card-version" id="c-version">
      <div class="card-label">Go Version</div>
      <div class="card-value" style="font-size:1.6rem" id="v-version">{{.GoVer}}</div>
      <div class="card-unit">runtime</div>
    </div>
  </div>

  <div class="status-bar">
    <span class="status-dot"></span>
    <span>LIVE · updates every 2s · Last: <span id="last-update">—</span></span>
  </div>

  <div class="info-card">
    <div class="info-grid">
      <div class="info-item">
        <div class="info-label">Host</div>
        <div class="info-value" title="{{.Hostname}}">{{.Hostname}}</div>
      </div>
      <div class="info-item">
        <div class="info-label">Instance</div>
        <div class="info-value">{{.Instance}}</div>
      </div>
      <div class="info-item">
        <div class="info-label">Port</div>
        <div class="info-value">{{.Port}}</div>
      </div>
    </div>
  </div>

  <div class="cert-banner">
    <div class="cert-icon">🎓</div>
    <div class="cert-content">
      <div class="cert-title">Envie de maîtriser Clever Cloud ?</div>
      <div class="cert-sub">Validez vos compétences avec la certification officielle — et devenez expert de la plateforme.</div>
    </div>
    <a class="cert-btn" href="https://academy.clever.cloud/" target="_blank" rel="noopener noreferrer">Obtenir la certification →</a>
  </div>

  <footer>
    <div class="footer-links">
      <a class="fl fl-cc" href="https://www.clever-cloud.com" target="_blank" rel="noopener noreferrer">
        <svg width="12" height="12" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path d="M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z"/></svg>
        clever-cloud.com
      </a>
      <a class="fl fl-li" href="https://www.linkedin.com/company/clever-cloud/" target="_blank" rel="noopener noreferrer">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><path d="M16 8a6 6 0 0 1 6 6v7h-4v-7a2 2 0 0 0-2-2 2 2 0 0 0-2 2v7h-4v-7a6 6 0 0 1 6-6zM2 9h4v12H2z"/><circle cx="4" cy="4" r="2"/></svg>
        LinkedIn
      </a>
      <a class="fl" href="https://developers.clever-cloud.com/doc/applications/go/" target="_blank" rel="noopener noreferrer">
        <svg width="12" height="12" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>
        Go on CC
      </a>
      <a class="fl" href="https://go.dev/doc/" target="_blank" rel="noopener noreferrer">
        <svg width="12" height="12" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
        Go Docs
      </a>
      <a class="fl fl-cert" href="https://academy.clever.cloud/" target="_blank" rel="noopener noreferrer">
        <svg width="12" height="12" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path d="M22 10v6M2 10l10-5 10 5-10 5z"/><path d="M6 12v5c3 3 9 3 12 0v-5"/></svg>
        Certification Clever Cloud
      </a>
    </div>
    <p class="footer-copy">Open source demo &middot; Deployed on Clever Cloud</p>
  </footer>
</div>

<script>
const prev = {};

function flash(cardId) {
  const el = document.getElementById(cardId);
  if (!el) return;
  el.classList.remove('flash');
  void el.offsetWidth;
  el.classList.add('flash');
  setTimeout(() => el.classList.remove('flash'), 600);
}

function set(valId, cardId, val) {
  const el = document.getElementById(valId);
  if (!el) return;
  const s = String(val);
  if (prev[valId] !== undefined && prev[valId] !== s) flash(cardId);
  prev[valId] = s;
  el.textContent = s;
}

let baseUptime = 0;
let lastFetch = Date.now();

function fmtUptime(sec) {
  if (sec < 60) return sec + 's';
  if (sec < 3600) return Math.floor(sec / 60) + 'm ' + (sec % 60) + 's';
  const h = Math.floor(sec / 3600), m = Math.floor((sec % 3600) / 60), s = sec % 60;
  return h + 'h ' + m + 'm ' + s + 's';
}

async function fetchStats() {
  try {
    const r = await fetch('/stats');
    const d = await r.json();
    baseUptime = d.uptime_sec;
    lastFetch = Date.now();
    set('v-goroutines', 'c-goroutines', d.goroutines);
    set('v-heap', 'c-heap', d.heap_mb);
    set('v-gc', 'c-gc', d.gc_cycles);
    set('v-requests', 'c-requests', d.requests);
    set('v-version', 'c-version', d.go_version);
    const now = new Date();
    document.getElementById('last-update').textContent = now.toTimeString().slice(0, 8);
  } catch (e) {}
}

function tickUptime() {
  const elapsed = Math.floor((Date.now() - lastFetch) / 1000);
  set('v-uptime', 'c-uptime', fmtUptime(baseUptime + elapsed));
}

fetchStats();
setInterval(fetchStats, 2000);
setInterval(tickUptime, 1000);
</script>
</body>
</html>`
