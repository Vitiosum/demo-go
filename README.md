# Go Runtime Dashboard — Live metrics on Clever Cloud

> Real-time Go runtime metrics (goroutines, heap, GC, uptime, requests) served from a single-file Go app deployed on Clever Cloud.

---

## Deploy on Clever Cloud

1. Fork this repository
2. In the Clever Cloud console, create a new **Go** application — connect your forked repo
3. No add-on needed
4. No environment variables to set manually — Clever Cloud injects `PORT` and `INSTANCE_NUMBER` automatically
5. Push → Clever Cloud builds and deploys automatically

**Configuration file:** `clevercloud/go.json`

```json
{ "deploy": { "appIsToBeBuilt": true } }
```

---

## Stack

| Layer      | Technology                              |
|------------|-----------------------------------------|
| Language   | Go 1.24                                 |
| Deps       | None (stdlib only)                      |
| Frontend   | HTML/CSS/JS embedded in `main.go`       |
| Fonts      | Inter, Newsreader, DM Mono (Google CDN) |
| Design     | Aura Full (dark, blue accent)           |

---

## Features

- Live dashboard updated every 2 seconds via JS polling
- Metrics: goroutines, heap (MB), uptime, GC cycles, request count, Go version
- Flash animation on metric change
- Sticky glass-blur nav, iridescent orbs background
- Marquee strip, animated shiny CTA button
- `/stats` JSON endpoint
- `/health` endpoint (200 OK)

---

## Local Development

### Prerequisites

- Go 1.24+

### Run

```bash
git clone https://github.com/Vitiosum/demo-go
cd demo-go
go run main.go
# → http://localhost:8080
```

---

## Environment Variables

| Variable          | Required | Description                                      |
|-------------------|----------|--------------------------------------------------|
| `PORT`            | auto     | Injected by Clever Cloud (default: 8080)         |
| `INSTANCE_NUMBER` | auto     | Injected by Clever Cloud for multi-instance setups |

No variables need to be set manually.

---

## Deployment Notes

- All HTML, CSS, and JS live inside the `indexHTML` constant in `main.go` — no static file serving
- The app listens on `0.0.0.0:$PORT` as required by Clever Cloud
- `INSTANCE_NUMBER` is injected automatically and displayed in the info panel
- Build time is fast — no external dependencies
