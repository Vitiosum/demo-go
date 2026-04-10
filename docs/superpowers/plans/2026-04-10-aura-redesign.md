# Aura Full Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remplacer le design system "Track Night" par "Aura Full" dans la constante `indexHTML` de `main.go`.

**Architecture:** Un seul fichier à modifier — `main.go`. Toute la logique Go (serveur HTTP, endpoint `/stats`, `/health`) et tout le JS (fetch, flash, tickUptime) restent inchangés. Seule la constante `indexHTML` est remplacée : fonts, CSS complet, et structure HTML du body.

**Tech Stack:** Go stdlib, Inter + Newsreader + DM Mono (Google Fonts CDN), CSS vanilla, JS vanilla.

---

## Fichiers touchés

| Fichier | Action | Contenu |
|---|---|---|
| `main.go` | Modifier | Remplacer la constante `indexHTML` (lignes 91–334) |

Aucun autre fichier Go n'est touché. Aucune dépendance Go ajoutée.

---

## Task 1 : Remplacer la constante `indexHTML`

**Fichiers :**
- Modifier : `main.go` — constante `indexHTML` (lignes 91–334)

- [ ] **Step 1 : Remplacer `indexHTML` avec le nouveau template Aura Full**

Dans `main.go`, remplacer tout le bloc entre `const indexHTML =` et le backtick fermant (ligne 334) par ce qui suit. **Le JS (fetch, flash, tickUptime) en bas du template est copié tel quel depuis l'original.**

```go
const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Go · Clever Cloud</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:ital,wght@0,300;0,400;0,600;0,700;1,300&family=Newsreader:ital,wght@1,300&family=DM+Mono:wght@300;400&display=swap" rel="stylesheet">
<style>
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
html{font-family:'Inter',sans-serif;background:hsl(0,0%,9%);color:hsl(0,0%,98%);min-height:100vh}
body{min-height:100vh;overflow-x:hidden;position:relative}

/* Orbs */
.orb{position:fixed;border-radius:50%;pointer-events:none;filter:blur(80px);opacity:0.2}
.orb-1{width:500px;height:500px;background:radial-gradient(circle,#3b82f6 0%,transparent 70%);top:-150px;right:-100px}
.orb-2{width:350px;height:350px;background:radial-gradient(circle,#8b5cf6 0%,transparent 70%);bottom:-100px;left:-80px}
.orb-3{width:250px;height:250px;background:radial-gradient(circle,#06b6d4 0%,transparent 70%);top:45%;left:40%}

/* Nav glass blur */
.nav{display:flex;align-items:center;justify-content:space-between;padding:12px 28px;background:rgba(23,23,23,0.75);backdrop-filter:blur(16px);-webkit-backdrop-filter:blur(16px);border-bottom:1px solid hsl(0,0%,20%);position:sticky;top:0;z-index:10}
.nav-logo{font-size:14px;font-weight:700;color:hsl(0,0%,98%);letter-spacing:-0.02em}
.nav-logo span{color:#3b82f6}
.nav-pill{display:flex;align-items:center;gap:6px;background:rgba(59,130,246,0.1);border:1px solid rgba(59,130,246,0.2);border-radius:99px;padding:5px 12px;font-size:10px;color:#60a5fa;font-weight:600;letter-spacing:0.06em;text-transform:uppercase}
.nav-dot{width:5px;height:5px;background:#3b82f6;border-radius:50%;animation:pulse 1.5s ease-in-out infinite}

/* Container */
.container{max-width:900px;margin:0 auto;padding:40px 20px 0;position:relative;z-index:1}

/* Hero */
.hero{text-align:center;margin-bottom:40px}
.live-badge{display:inline-flex;align-items:center;gap:8px;background:rgba(59,130,246,0.08);border:1px solid rgba(59,130,246,0.2);border-radius:99px;color:#60a5fa;font-size:11px;font-weight:600;letter-spacing:0.06em;text-transform:uppercase;padding:5px 16px;margin-bottom:18px}
.live-dot{width:6px;height:6px;background:#3b82f6;border-radius:50%;animation:pulse 1.5s ease-in-out infinite}
@keyframes pulse{0%,100%{opacity:1;transform:scale(1)}50%{opacity:0.4;transform:scale(0.8)}}
h1{font-size:clamp(2.5rem,7vw,4rem);font-weight:700;color:hsl(0,0%,98%);letter-spacing:-0.05em;line-height:1.05;margin-bottom:6px}
.hero-serif{display:block;font-family:'Newsreader',serif;font-style:italic;font-weight:300;font-size:clamp(1.4rem,4vw,2.2rem);color:hsl(0,0%,65%);letter-spacing:-0.02em;margin-bottom:12px}
.hero-sub{color:hsl(0,0%,55%);font-size:14px;letter-spacing:-0.01em}

/* Metric cards */
.metrics{display:grid;grid-template-columns:repeat(3,1fr);gap:12px;margin-bottom:24px}
@media(max-width:620px){.metrics{grid-template-columns:repeat(2,1fr)}}
@media(max-width:380px){.metrics{grid-template-columns:1fr}}
.card{background:hsl(0,0%,11%);border:1px solid hsl(0,0%,20%);border-radius:12px;padding:18px 16px;position:relative;overflow:hidden;transition:border-color 0.3s}
.card::before{content:'';position:absolute;top:0;left:0;right:0;height:2px;border-radius:12px 12px 0 0;background:var(--accent)}
.card::after{content:'';position:absolute;top:0;left:0;right:0;bottom:0;background:radial-gradient(ellipse at top,var(--glow) 0%,transparent 70%);opacity:0;transition:opacity 0.3s;pointer-events:none}
.card.flash::after{opacity:1;animation:flash-fade 0.5s ease-out forwards}
@keyframes flash-fade{0%{opacity:0.4}100%{opacity:0}}
.card-goroutines{--accent:linear-gradient(90deg,#3b82f6,#2563eb);--glow:rgba(59,130,246,0.12)}
.card-heap{--accent:linear-gradient(90deg,#8b5cf6,#7c3aed);--glow:rgba(139,92,246,0.10)}
.card-uptime{--accent:linear-gradient(90deg,#22c55e,#16a34a);--glow:rgba(34,197,94,0.10)}
.card-gc{--accent:linear-gradient(90deg,#06b6d4,#0891b2);--glow:rgba(6,182,212,0.10)}
.card-requests{--accent:linear-gradient(90deg,#fbbf24,#f59e0b);--glow:rgba(245,158,11,0.10)}
.card-version{--accent:linear-gradient(90deg,hsl(0,0%,40%),hsl(0,0%,28%));--glow:rgba(100,116,139,0.08)}
.card-label{font-size:9px;font-weight:600;letter-spacing:0.1em;text-transform:uppercase;color:hsl(0,0%,55%);margin-bottom:10px}
.card-value{font-family:'DM Mono',monospace;font-size:2rem;font-weight:300;color:hsl(0,0%,98%);line-height:1;margin-bottom:4px;font-variant-numeric:tabular-nums;letter-spacing:-0.02em}
.card-unit{font-size:10px;color:hsl(0,0%,40%)}

/* Status bar */
.status-bar{display:flex;align-items:center;justify-content:center;gap:6px;font-size:11px;color:hsl(0,0%,45%);margin-bottom:24px;letter-spacing:0.01em}
.status-dot{width:5px;height:5px;background:#22c55e;border-radius:50%;animation:pulse 2s ease-in-out infinite}

/* Info card */
.info-card{background:hsl(0,0%,11%);border:1px solid hsl(0,0%,20%);border-radius:12px;padding:16px 20px;margin-bottom:24px}
.info-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:10px}
@media(max-width:500px){.info-grid{grid-template-columns:1fr 1fr}}
.info-item{background:hsl(0,0%,9%);border:1px solid hsl(0,0%,20%);border-radius:8px;padding:10px}
.info-label{font-size:8px;font-weight:600;letter-spacing:0.1em;text-transform:uppercase;color:hsl(0,0%,40%);margin-bottom:4px}
.info-value{font-family:'DM Mono',monospace;font-size:12px;color:hsl(0,0%,60%);overflow:hidden;text-overflow:ellipsis;white-space:nowrap}

/* Marquee */
.marquee-wrap{overflow:hidden;position:relative;margin-bottom:24px}
.marquee-wrap::before,.marquee-wrap::after{content:'';position:absolute;top:0;bottom:0;width:80px;z-index:2;pointer-events:none}
.marquee-wrap::before{left:0;background:linear-gradient(90deg,hsl(0,0%,9%) 0%,transparent 100%)}
.marquee-wrap::after{right:0;background:linear-gradient(-90deg,hsl(0,0%,9%) 0%,transparent 100%)}
.marquee-track{display:flex;gap:10px;width:max-content;animation:marquee 30s linear infinite}
@keyframes marquee{0%{transform:translateX(0)}100%{transform:translateX(-50%)}}
.marquee-item{flex-shrink:0;display:flex;align-items:center;gap:6px;background:hsl(0,0%,11%);border:1px solid hsl(0,0%,20%);border-radius:6px;padding:6px 14px;font-size:11px;color:hsl(0,0%,50%);font-weight:500;white-space:nowrap}
.marquee-dot{width:4px;height:4px;border-radius:50%;background:hsl(0,0%,35%);flex-shrink:0}

/* Shiny CTA */
.shiny-wrap{display:flex;justify-content:center;margin-bottom:24px}
.shiny-cta{position:relative;display:inline-flex;align-items:center;gap:10px;padding:14px 32px;border-radius:99px;font-family:'Inter',sans-serif;font-size:14px;font-weight:600;color:hsl(0,0%,98%);letter-spacing:-0.01em;background:hsl(0,0%,9%);border:none;cursor:pointer;text-decoration:none;outline:none}
.shiny-cta::before{content:'';position:absolute;inset:-1.5px;border-radius:99px;background:conic-gradient(from 0deg,#3b82f6,#8b5cf6,#06b6d4,#fbbf24,#3b82f6);z-index:-1;animation:border-spin 4s linear infinite}
.shiny-cta::after{content:'';position:absolute;inset:1px;border-radius:99px;background:hsl(0,0%,9%);background-image:radial-gradient(circle,rgba(255,255,255,0.07) 1px,transparent 1px);background-size:16px 16px;z-index:-1}
@keyframes border-spin{0%{transform:rotate(0deg)}100%{transform:rotate(360deg)}}
.shiny-icon{font-size:18px}

/* Footer */
footer{border-top:1px solid hsl(0,0%,18%);padding:20px 0 12px;margin-top:0}
.footer-links{display:flex;flex-wrap:wrap;gap:8px;justify-content:center;margin-bottom:12px}
.fl{display:inline-flex;align-items:center;gap:6px;background:rgba(255,255,255,0.03);border:1px solid hsl(0,0%,18%);border-radius:8px;color:hsl(0,0%,40%);font-size:11px;padding:6px 12px;text-decoration:none;transition:all 0.2s}
.fl:hover{background:rgba(255,255,255,0.06);color:hsl(0,0%,65%);border-color:hsl(0,0%,25%)}
.fl-cc{color:#86efac!important;border-color:rgba(34,197,94,0.2)!important;background:rgba(34,197,94,0.05)!important}
.fl-cc:hover{background:rgba(34,197,94,0.1)!important}
.fl-li{color:#93c5fd!important;border-color:rgba(96,165,250,0.2)!important;background:rgba(96,165,250,0.05)!important}
.fl-li:hover{background:rgba(96,165,250,0.1)!important}
.fl-cert{color:#fbbf24!important;border-color:rgba(251,191,36,0.25)!important;background:rgba(251,191,36,0.06)!important}
.fl-cert:hover{background:rgba(251,191,36,0.12)!important}
.footer-copy{text-align:center;color:hsl(0,0%,30%);font-size:11px;padding-bottom:8px}
</style>
</head>
<body>

<div class="orb orb-1"></div>
<div class="orb orb-2"></div>
<div class="orb orb-3"></div>

<nav class="nav">
  <div class="nav-logo">Go<span>.</span>Runtime</div>
  <div class="nav-pill"><div class="nav-dot"></div>LIVE</div>
</nav>

<div class="container">
  <div class="hero">
    <div class="live-badge"><span class="live-dot"></span>Live runtime</div>
    <h1>Go Runtime<span class="hero-serif">dashboard</span></h1>
    <p class="hero-sub">Real-time metrics — refreshed every 2 seconds</p>
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
      <div class="card-value" style="font-size:1.4rem" id="v-version">{{.GoVer}}</div>
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

  <div class="marquee-wrap">
    <div class="marquee-track">
      <div class="marquee-item"><div class="marquee-dot"></div>Go 1.24</div>
      <div class="marquee-item"><div class="marquee-dot"></div>Clever Cloud</div>
      <div class="marquee-item"><div class="marquee-dot"></div>Goroutines</div>
      <div class="marquee-item"><div class="marquee-dot"></div>Runtime Metrics</div>
      <div class="marquee-item"><div class="marquee-dot"></div>GC Cycles</div>
      <div class="marquee-item"><div class="marquee-dot"></div>stdlib only</div>
      <div class="marquee-item"><div class="marquee-dot"></div>No dependencies</div>
      <div class="marquee-item"><div class="marquee-dot"></div>Open source</div>
      <div class="marquee-item"><div class="marquee-dot"></div>Go 1.24</div>
      <div class="marquee-item"><div class="marquee-dot"></div>Clever Cloud</div>
      <div class="marquee-item"><div class="marquee-dot"></div>Goroutines</div>
      <div class="marquee-item"><div class="marquee-dot"></div>Runtime Metrics</div>
      <div class="marquee-item"><div class="marquee-dot"></div>GC Cycles</div>
      <div class="marquee-item"><div class="marquee-dot"></div>stdlib only</div>
      <div class="marquee-item"><div class="marquee-dot"></div>No dependencies</div>
      <div class="marquee-item"><div class="marquee-dot"></div>Open source</div>
    </div>
  </div>

  <div class="shiny-wrap">
    <a class="shiny-cta" href="https://academy.clever.cloud/" target="_blank" rel="noopener noreferrer">
      <span class="shiny-icon">🎓</span>
      Obtenir la certification Clever Cloud →
    </a>
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
      <a class="fl" href="https://www.clever.cloud/developers/doc/deploy/application/golang/go/" target="_blank" rel="noopener noreferrer">
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
```

- [ ] **Step 2 : Vérifier que le code Go compile**

```bash
cd /Users/vitio/ClaudeByVitio/demo-go
go build -o /tmp/demo-go-test main.go
```

Résultat attendu : aucune erreur, binaire généré dans `/tmp/`.

- [ ] **Step 3 : Tester l'app localement**

```bash
go run main.go
```

Ouvrir http://localhost:8080 et vérifier :
- [ ] Fond sombre `hsl(0,0%,9%)` — plus de #07080C
- [ ] Nav sticky glass blur avec "Go.Runtime" et badge LIVE bleu
- [ ] 3 orbs bleu/violet/cyan visibles en arrière-plan
- [ ] Titre "Go Runtime" Inter 700 + "dashboard" Newsreader italic en dessous
- [ ] 6 cartes avec top accent 2px coloré (bleu/violet/vert/cyan/or/slate)
- [ ] Marquee qui défile horizontalement
- [ ] Bouton Shiny CTA avec border animé conic-gradient
- [ ] Footer avec liens colorés
- [ ] Les métriques se mettent à jour toutes les 2s (flash sur changement)

Stopper avec `Ctrl+C`.

- [ ] **Step 4 : Commit**

```bash
cd /Users/vitio/ClaudeByVitio/demo-go
git add main.go
git commit -m "feat: Aura Full redesign — Inter/Newsreader, glass nav, orbs, marquee, shiny CTA"
```

---

## Task 2 : Deploy sur Clever Cloud

**Fichiers :** aucun — push du commit existant

- [ ] **Step 1 : Push**

```bash
cd /Users/vitio/ClaudeByVitio/demo-go
git push
```

Résultat attendu : Clever Cloud déclenche automatiquement un rebuild Go. Le build doit compléter en < 60s (pas de dépendances, stdlib only).

- [ ] **Step 2 : Mettre à jour CLAUDE.md**

Dans `main.go` → `CLAUDE.md` du projet, mettre à jour la section design :

```markdown
## 🎨 Design

**Design system :** Aura Full (Inter 300–700 + Newsreader italic 300 + DM Mono 300)
**Couleurs :** fond `hsl(0,0%,9%)`, cartes `hsl(0,0%,11%)`, accent bleu `#3b82f6`, or `#fbbf24`
**Composants :** nav glass blur, orbs iridescents (bleu/violet/cyan), hero Inter/Newsreader, cartes top-accent, marquee, shiny CTA (conic-gradient animé), footer
**Commit :** feat: Aura Full redesign
```

- [ ] **Step 3 : Commit CLAUDE.md**

```bash
cd /Users/vitio/ClaudeByVitio/demo-go
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md with Aura Full design system"
git push
```
