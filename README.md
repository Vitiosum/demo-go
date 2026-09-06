# Go Runtime Dashboard — Live metrics on Clever Cloud

> Real-time Go runtime metrics (goroutines, heap, GC, uptime, requests) served by a dependency-free Go app deployed on Clever Cloud — dressed in the **Clever Brand Kit**, with the Clever Cloud certification front and center.

---

## Deploy on Clever Cloud

1. Fork this repository
2. In the Clever Cloud console, create a new **Go** application — connect your forked repo
3. No add-on needed
4. No environment variables to set manually — Clever Cloud injects `PORT`, `INSTANCE_NUMBER` and the `CC_*` variables automatically
5. Push → Clever Cloud builds and deploys automatically

**Configuration file:** `clevercloud/go.json`

```json
{ "deploy": { "appIsToBeBuilt": true } }
```

---

## Stack

| Layer      | Technology                                                              |
|------------|-------------------------------------------------------------------------|
| Language   | Go 1.24 (`go.mod`)                                                      |
| Deps       | None (stdlib + `embed` only)                                            |
| Frontend   | `static/index.html` (Go `html/template`), compiled into the binary      |
| Fonts      | Plus Jakarta Sans, JetBrains Mono (Google Fonts, system fallback)       |
| Design     | Clever Brand Kit (Plus Jakarta Sans, navy #13172e, dégradé Clever)      |

---

## Features

- Live dashboard refreshed every 2 seconds via JS polling on `/stats`, with a 300 ms orange flash on every value change
- Metrics: goroutines, heap (MB), uptime, GC cycles, request count, Go version
- **Certification Clever Cloud** block right under the hero (two official tracks + CTA to the Academy)
- **« Vu depuis Clever Cloud »** panel: application, App ID, instance (`INSTANCE_NUMBER` · `CC_PRETTY_INSTANCE_NAME`), instance type, deployed commit, deployment ID, host:port, runtime — shows « Local · hors Clever Cloud » when the app runs outside the platform
- `/stats` JSON endpoint, `/health` endpoint (200 OK), `/cc-brand.css` (embedded stylesheet)

---

## Certification Clever Cloud

The demo puts the [Clever Cloud Academy](https://academy.clever.cloud/) certification at the heart of the page: a badge, the two official tracks (**Cloud Computing Fundamentals** and **Advanced Deployment**) and a call to action. Digital badges are issued automatically once a track is validated.

---

## Local Development

### Prerequisites

- Go 1.24+

### Run

```bash
git clone https://github.com/Vitiosum/demo-go
cd demo-go
go run .
# → http://localhost:8080  (PORT=8084 go run . to pick another port)
```

### Check

```bash
go vet ./... && go build -o /dev/null .
curl localhost:8080/health   # OK
curl localhost:8080/stats    # JSON
```

---

## Project structure

```
main.go               → HTTP server: /, /stats, /health, /cc-brand.css, platform info (CC_* env vars)
static/index.html     → page template (html/template): Clever Brand Kit markup + polling JS
static/cc-brand.css   → Clever Brand Kit stylesheet, copied as-is from the shared kit
clevercloud/go.json   → Clever Cloud build configuration
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
| `CC_COMMIT_ID`, `CC_DEPLOYMENT_ID` | auto | Deployed commit (7 chars) and deployment ID (16 chars)       |

No variables need to be set manually. Reference: [Clever Cloud environment variables](https://www.clever.cloud/developers/doc/reference/reference-environment-variables/).

---

## Deployment Notes

- The app listens on `0.0.0.0:$PORT` as required by Clever Cloud
- HTML, CSS and JS live in `static/` and are embedded at build time — no static file serving from disk
- Only `/` counts as a request; `/stats` polling does not inflate the counter
- Build time is fast — no external dependencies
