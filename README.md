# Go Runtime Dashboard — Live metrics on Clever Cloud

> Real-time Go runtime metrics (goroutines, heap, GC, uptime, requests) served by a dependency-free Go app deployed on Clever Cloud — dressed in the **Clever Brand Kit**, with the Clever Cloud certification front and center.

---

## Deploy on Clever Cloud

1. Fork this repository
2. In the Clever Cloud console, create a new **Go** application — connect your forked repo
3. No add-on needed
4. No environment variables to set manually — Clever Cloud injects `PORT`, `INSTANCE_NUMBER` and the `CC_*` variables automatically
5. Optional but recommended: wire the health check — `clever env set CC_HEALTH_CHECK_PATH /health`
6. Push → Clever Cloud builds and deploys automatically

**Build configuration:** none. The build is driven by `go.mod` (module `github.com/Vitiosum/demo-go`, `go 1.26`): Clever Cloud detects the Go runtime, compiles the module and runs the resulting binary — no Dockerfile, no image to maintain, exactly what the page claims. The instance needs Go ≥ 1.26 (the toolchain is downloaded automatically when `go.mod` asks for a newer one; otherwise set `CC_GO_VERSION`).

`clevercloud/go.json` is still in the repository but is a **deprecated** mechanism according to Clever Cloud's Go documentation; it is not needed and can be deleted.

---

## Stack

| Layer      | Technology                                                              |
|------------|-------------------------------------------------------------------------|
| Language   | Go 1.26 (`go.mod`)                                                      |
| Deps       | None (stdlib + `embed` only)                                            |
| Frontend   | `static/index.html` (Go `html/template`), compiled into the binary      |
| Fonts      | Plus Jakarta Sans, JetBrains Mono (Google Fonts, system fallback)       |
| Design     | Clever Brand Kit (Plus Jakarta Sans, navy #13172e, dégradé Clever)      |

---

## Features

- Live dashboard refreshed every 2 seconds via JS polling on `/stats`, with a 300 ms orange flash when a metric changes (uptime ticks every second, without flashing)
- Metrics: goroutines, heap (MB), uptime, GC cycles, request count, Go version
- **Certification Clever Cloud** block right under the hero (two official tracks + CTA to the Academy)
- **« Vu depuis Clever Cloud »** panel: application, App ID, instance (`INSTANCE_NUMBER` · `CC_PRETTY_INSTANCE_NAME`), instance type, deployed commit, deployment ID, host:port, runtime — shows « Local · hors Clever Cloud » when the app runs outside the platform
- `/stats` JSON endpoint, `/health` endpoint (200 OK), `/cc-brand.css` and `/app.js` (embedded assets) — GET only (405 otherwise), unknown paths → 404
- Security headers on every response: strict Content-Security-Policy (no inline script, Google Fonts allowed), `X-Content-Type-Options`, `Referrer-Policy`, `X-Frame-Options`
- Hardened server: read/write/idle timeouts, graceful shutdown on SIGTERM (in-flight requests finish before a redeploy kills the instance)

---

## Certification Clever Cloud

The demo puts the [Clever Cloud Academy](https://academy.clever.cloud/) certification at the heart of the page: a badge, the two official tracks (**Cloud Computing Fundamentals** and **Advanced Deployment**) and a call to action. Digital badges are issued automatically once a track is validated.

---

## Local Development

### Prerequisites

- Go 1.26+

### Run

```bash
git clone https://github.com/Vitiosum/demo-go
cd demo-go
go run .
# → http://localhost:8080  (PORT=8084 go run . to pick another port)
```

### Check

```bash
gofmt -l . ; go vet ./... && go build -o /dev/null . && go test ./...
curl localhost:8080/health   # OK
curl localhost:8080/stats    # JSON
```

### Known vulnerabilities (stdlib included)

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

---

## Project structure

```
main.go               → HTTP server (timeouts, SIGTERM shutdown), routes, secure() middleware (CSP), platform info (CC_* env vars)
main_test.go          → httptest smoke tests (status codes, JSON, headers)
static/index.html     → page template (html/template): Clever Brand Kit markup
static/app.js         → polling script for /stats (separate file: the CSP forbids inline scripts)
static/cc-brand.css   → Clever Brand Kit stylesheet, copied as-is from the shared kit
clevercloud/go.json   → legacy build config, deprecated by Clever Cloud (the build is driven by go.mod)
docs/superpowers/     → design specs and implementation plans
```

`static/` is embedded with `//go:embed static` — the binary is self-contained, nothing to copy next to it.

---

## Environment Variables

| Variable                  | Required | Description                                                      |
|---------------------------|----------|------------------------------------------------------------------|
| `PORT`                    | auto     | Injected by Clever Cloud (default: 8080)                         |
| `INSTANCE_NUMBER`         | auto     | Injected by Clever Cloud, shown in the platform panel            |
| `CC_APP_NAME`, `APP_ID`   | auto     | Application name and ID (panel; `APP_ID` toggles « Production ») |
| `INSTANCE_TYPE`, `CC_PRETTY_INSTANCE_NAME` | auto | Instance type and friendly name                       |
| `COMMIT_ID` / `CC_COMMIT_ID`, `CC_DEPLOYMENT_ID` / `DEPLOYMENT_ID` | auto | Deployed commit (7 chars) and deployment ID (16 chars) — both names are tried |
| `CC_HEALTH_CHECK_PATH`    | optional | Set to `/health` so Clever Cloud checks the app before switching traffic |

No variables need to be set manually. Reference: [Clever Cloud environment variables](https://www.clever.cloud/developers/doc/reference/reference-environment-variables/).

**Empty deployed commit?** Not a bug: the commit context is only attached to a deployment coming from the GitHub integration (`git push`). A deployment triggered by `clever restart` carries none — `clever activity` shows `N/A` in the commit column — and the panel shows a dash.

---

## Deployment Notes

- The app listens on `0.0.0.0:$PORT` as required by Clever Cloud
- HTML, CSS and JS live in `static/` and are embedded at build time — no static file serving from disk
- Only `/` counts as a request; `/stats` polling does not inflate the counter
- Build time is fast — no external dependencies
