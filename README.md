# Go Runtime Dashboard — Live metrics on Clever Cloud

> The Go process that answers you, measured every second: goroutines, heap, GC cycles and pauses, drawn live on an oscilloscope. Three bounded buttons make the runtime react on camera. Standard library only, deployed on Clever Cloud, dressed in the **Clever Brand Kit**, with the Clever Cloud Academy certification front and center.

---

## Deploy on Clever Cloud

1. Create a **Go** application with clever-tools (`clever create --type go demo-go`). There is no GitHub link (`clever create --github` is refused by the API: "wrong oauth application data"; the account's GitHub integration must be reconnected in the Console)
2. No add-on needed
3. Recommended variables (all checked against the [Go runtime doc](https://www.clever.cloud/developers/doc/deploy/applications/golang/)):

```bash
clever env set CC_GO_BUILD_TOOL gomod          # build the module from go.mod, no detection
clever env set CC_HEALTH_CHECK_PATH /health    # deployment validated on /health
```

4. Optional demo variables (see below), then `git push` (GitHub backup) and `clever deploy`: Clever Cloud runs `go install` on the module and starts the binary. A plain `git push` does not redeploy.

Live app from 2026-10-25 to 2026-11-05 (deletion planned after the 2026-11-04 video shoot): `app_61a7ff90-36fe-49eb-b42b-2077d0e9db58`, https://app-61a7ff90-36fe-49eb-b42b-2077d0e9db58.cleverapps.io/. After that date, recreate it with the procedure above.

**Why the build is short:** no external dependency (no `go.sum`, nothing to download), and `go.mod` declares `go 1.27`, which is not newer than the Go of the image (1.27.3 since the 2026-10-21 image update). The doc warns: *"If the Go version declared in your go.mod file is newer than the one installed on the instance, the Go toolchain downloads the required version automatically"* — an extra download on every build that this setting avoids. `CC_CACHE_DEPENDENCIES` brings nothing here (no module to cache) and is not needed. No dedicated build instance needed.

`clevercloud/go.json` is a legacy file; the doc says that using it *"is a deprecated method and should no longer be used"*. It is kept until a deployment confirms it can be removed.

---

## Demo variables (environment variables tab of the console)

| Variable | Default | Visible effect |
|---|---|---|
| `APP_ENV` | `local` | Environment pill on the control desk (Clever gradient when `production`) |
| `ANNOUNCEMENT` | empty | Banner on top of the page when set (160 characters max) |
| `BURST_GOROUTINES` | `500` (1..2000) | Size of the goroutine burst: button label and height of the spike |
| `HEAP_ALERT_MB` | `32` (1..1024) | Dashed threshold on the chart; the Heap figure turns red and « Heap au-dessus de N MB » appears above it |

Invalid or out-of-range values fall back to the default. Scenario: change `HEAP_ALERT_MB` to `128`, `clever restart`, reload → the 64 MB allocation no longer crosses the threshold.

Platform variables shown on the desk (read-only, « injecté au déploiement » locally): `CC_APP_NAME`, `APP_ID`, `INSTANCE_NUMBER` · `CC_PRETTY_INSTANCE_NAME`, `INSTANCE_TYPE`, `COMMIT_ID`/`CC_COMMIT_ID`, `CC_DEPLOYMENT_ID`/`DEPLOYMENT_ID`. The commit comes from the platform variable, which stays set after a `clever restart` (checked on another demo) and is also set by `clever deploy`; when the variable is empty, the desk shows the fallback label « aucun (clever restart) ».

---

## Bounded load generators

| Route | Effect | Bound |
|---|---|---|
| `POST /load/goroutines` | `BURST_GOROUTINES` goroutines, each sleeping 4 to 12 s | one burst at a time (max 2000), 429 otherwise |
| `POST /load/heap` | 64 MB allocated (pages touched), held 8 s, released, then `runtime.GC()` | one allocation at a time, fixed size |
| `POST /load/gc` | `runtime.GC()` | one per 2 s |

Worst case: 2000 sleeping goroutines plus 64 MB. No CPU loop, nothing persistent.

---

## Routes

`GET /` (page), `GET /stats` (latest sample, uptime, page count, burst state; `?history=1` adds the last 90 samples), `GET /health`, `GET /cc-brand.css`, `GET /app.js`, the three `POST /load/*`. Other verbs → 405, other paths → 404. Security headers on every response (strict CSP: no inline script).

The runtime is sampled once per second by a single goroutine (`runtime.ReadMemStats` stops the world, so it never runs per request).

---

## Local development

```bash
go run .                                   # http://localhost:8080
PORT=8104 APP_ENV=production ANNOUNCEMENT="Maintenance ce soir à 22 h" \
  BURST_GOROUTINES=800 HEAP_ALERT_MB=40 go run .
gofmt -l . ; go vet ./... && go test -race ./...
```

Requires Go 1.27+.

---

## Project structure

```
main.go             → HTTP server (timeouts, SIGTERM graceful shutdown), routes, CSP middleware, handlers
pulse.go            → runtime sampler (ring of 90 samples) and bounded load generators
config.go           → demo variables (APP_ENV, ANNOUNCEMENT, BURST_GOROUTINES, HEAP_ALERT_MB) and CC_* platform variables
main_test.go        → httptest tests (routes, headers, demo variables, bounds)
static/index.html   → page template (html/template), logo and badge SVG
static/app.js       → oscilloscope drawing, polling, buttons (separate file: strict CSP)
static/cc-brand.css → Clever Brand Kit, copied as-is (do not edit)
```

`static/` is embedded with `//go:embed static`: the binary is self-contained.
