# The Runtime Oscilloscope — Go on Clever Cloud

> The Go process that answers you, measured every second: goroutines, heap, GC cycles and pauses, drawn live on an oscilloscope over the last 90 seconds. Three bounded buttons make the runtime react on camera: 500 goroutines that sleep and fade away, 64 MB held for 8 seconds then collected, a forced GC. Standard library only, no add-on: the **Clever Cloud Go runtime** builds the module from `go.mod`, with no Dockerfile, and starts the binary. Dressed in the Clever Brand Kit.

![The dashboard run locally after two rounds of the buttons: 508 goroutines, 65 MB of heap above the HEAP_ALERT_MB line, GC ticks on the oscilloscope](docs/screenshot-desktop.png)

<img src="docs/screenshot-phone.png" alt="The headline and the oscilloscope on a phone, run locally, with the heap above its threshold" width="300">

**Run it yourself.** The hosted demo was taken down after the presales review (November 2026). The steps in [Deploy on Clever Cloud](#deploy-on-clever-cloud) recreate it from scratch in about 3 minutes.

## What the Go runtime does here

| Where | Feature | What you see |
|---|---|---|
| Build | Go runtime, `CC_GO_BUILD_TOOL=gomod`: `go install` of the module declared in `go.mod` | The section *Du git push au binaire* lists the four real steps. No dependency, so the build downloads nothing |
| Start | The binary `$GOPATH/bin/demo-go` listens on `0.0.0.0:8080`; on SIGTERM (redeploy, scale-down) it finishes in-flight requests, 10 s max | `/health` answers 200, then traffic switches over |
| Every second | One sampler goroutine: `runtime.NumGoroutine`, `runtime.ReadMemStats` (it stops the world, so it never runs per request), ring of 90 samples | The headline *N goroutines en ce moment.*, the oscilloscope (full as soon as the page opens), heap, GC cycles, last GC pause, `GOMAXPROCS` |
| The control desk | Three bounded load generators (below) | A goroutine spike and its slope; a heap plateau over the dashed `HEAP_ALERT_MB` line, then the drop with a GC tick; the button counts down while a burst runs |
| The control desk | Demo variables and platform variables | `APP_ENV` pill, `ANNOUNCEMENT` banner, burst size, threshold, instance, commit and deployment |

| Route | Effect | Bound |
|---|---|---|
| `POST /load/goroutines` | `BURST_GOROUTINES` goroutines, each sleeping 4 to 12 s | one burst at a time (max 2000), 429 otherwise |
| `POST /load/heap` | 64 MB allocated (pages touched), held 8 s, released, then `runtime.GC()` | one allocation at a time, fixed size |
| `POST /load/gc` | `runtime.GC()` | one per 2 s |

Worst case: 2000 sleeping goroutines plus 64 MB. No CPU loop, nothing persistent.

## Deploy on Clever Cloud

You need [clever-tools](https://www.clever.cloud/developers/doc/cli/) (`clever login` done) and the ID of the organisation that will own the resources (`orga_…`, shown in the Console). Run the commands from the repo root: `clever create` links the folder to the new app.

```bash
# 1. The app and its size (no dedicated build instance: the build compiles the standard library only)
clever create --type go demo-go --org <org_id> --region par
clever scale --min-flavor nano --max-flavor XS --min-instances 1 --max-instances 2

# 2. Variables (checked against the Go runtime doc)
clever env set CC_GO_BUILD_TOOL gomod          # build the module from go.mod, no detection
clever env set CC_HEALTH_CHECK_PATH /health    # deployment validated on /health

# 3. Deploy
clever deploy
clever open
```

No add-on, no setup script. Clever Cloud runs `go install` on the module and starts the binary; `clever deploy` is what deploys (a `git push` to another remote does not).

**Why the build is short:** no external dependency (no `go.sum`, nothing to download), and `go.mod` declares `go 1.27`, which is not newer than the Go of the image (1.27.3 since the 2026-10-21 image update). The [Go runtime doc](https://www.clever.cloud/developers/doc/deploy/applications/golang/) warns: *"If the Go version declared in your go.mod file is newer than the one installed on the instance, the Go toolchain downloads the required version automatically"*, an extra download on every build that this setting avoids. `CC_CACHE_DEPENDENCIES` brings nothing here (no module to cache).

`clevercloud/go.json` is a legacy file; the doc says that using it *"is a deprecated method and should no longer be used"*. It is kept until a deployment confirms it can be removed.

## Variables

| Variable | Set by | Role |
|---|---|---|
| `CC_GO_BUILD_TOOL` | you | `gomod`: build the module from `go.mod` |
| `CC_HEALTH_CHECK_PATH` | you | `/health` |
| `APP_ENV` | optional, default `local` | Environment pill on the control desk (Clever gradient when `production`) |
| `ANNOUNCEMENT` | optional, empty | Banner on top of the page when set (160 characters max) |
| `BURST_GOROUTINES` | optional, default `500` (1 to 2000) | Size of the goroutine burst: button label and height of the spike |
| `HEAP_ALERT_MB` | optional, default `32` (1 to 1024) | Dashed threshold on the chart; above it the Heap figure turns red and *Heap au-dessus de N MB* appears |
| `PORT` | Clever Cloud | Listen port, `8080` when unset |

Invalid or out-of-range values fall back to the default. To show it live: `clever env set HEAP_ALERT_MB 128`, `clever restart`, reload: the 64 MB allocation no longer crosses the threshold.

The desk shows, read-only, `CC_APP_NAME`, `APP_ID` (switches the *Live on Clever Cloud* pill), `INSTANCE_NUMBER`, `CC_PRETTY_INSTANCE_NAME`, `INSTANCE_TYPE`, `COMMIT_ID`/`CC_COMMIT_ID` and `CC_DEPLOYMENT_ID`/`DEPLOYMENT_ID`; locally each one reads « injecté au déploiement ». The commit variable is set by `clever deploy` and stays set after a `clever restart` (checked on another demo); when it is empty on Clever Cloud, the desk says « aucun (clever restart) ».

## Routes

`GET /` (page) · `GET /stats` (latest sample, uptime, page count, burst state; `?history=1` adds the last 90 samples; the page polls it every second) · `GET /health` · `GET /cc-brand.css` · `GET /app.js` · the three `POST /load/*`. Other verbs → 405, other paths → 404. Security headers on every response, 404 and 405 included (strict CSP: no inline script).

## Code tour

```
main.go             HTTP server (timeouts, SIGTERM graceful shutdown), routes, CSP middleware, handlers
pulse.go            runtime sampler (ring of 90 samples) and the bounded load generators
config.go           demo variables (APP_ENV, ANNOUNCEMENT, BURST_GOROUTINES, HEAP_ALERT_MB) and CC_* platform variables
main_test.go        httptest tests (routes, headers, demo variables, bounds)
static/index.html   page template (html/template), logo and badge SVG
static/app.js       oscilloscope drawing, polling, buttons (separate file: strict CSP)
static/cc-brand.css Clever Brand Kit, copied as-is (do not edit)
clevercloud/go.json legacy file, deprecated by the doc (see above)
```

`static/` is embedded with `//go:embed static`: the binary is self-contained. Go standard library only.

## Run locally

```bash
go run .                                   # http://localhost:8080
PORT=8104 APP_ENV=production ANNOUNCEMENT="Maintenance ce soir à 22 h" \
  BURST_GOROUTINES=800 HEAP_ALERT_MB=40 go run .
gofmt -l . ; go vet ./... && go test -race ./...
```

Requires Go 1.27+. The screenshots above come from `go run .` on a laptop (hence « Local · hors Clever Cloud » and `darwin/arm64`).

## License

[MIT](LICENSE) © 2026 [Vitiosum](https://github.com/Vitiosum)
