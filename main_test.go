package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// do runs one request through the full handler chain (routes + secure middleware).
func do(t *testing.T, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	newHandler().ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func TestIndex(t *testing.T) {
	rec := do(t, http.MethodGet, "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	if !strings.Contains(rec.Body.String(), `<script src="/app.js">`) {
		t.Error("page must load /app.js (no inline script under the CSP)")
	}
}

func TestUnknownPathIs404(t *testing.T) {
	if rec := do(t, http.MethodGet, "/foo"); rec.Code != http.StatusNotFound {
		t.Fatalf("GET /foo = %d, want 404", rec.Code)
	}
}

func TestPostIs405(t *testing.T) {
	if rec := do(t, http.MethodPost, "/"); rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST / = %d, want 405", rec.Code)
	}
}

func TestHealth(t *testing.T) {
	rec := do(t, http.MethodGet, "/health")
	if rec.Code != http.StatusOK || rec.Body.String() != "OK" {
		t.Fatalf("GET /health = %d %q, want 200 OK", rec.Code, rec.Body.String())
	}
}

func TestStats(t *testing.T) {
	rec := do(t, http.MethodGet, "/stats")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /stats = %d, want 200", rec.Code)
	}
	var s statsResponse
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("/stats is not valid JSON: %v", err)
	}
	if s.GoVersion == "" || s.Now.Goroutines == 0 || s.History != nil {
		t.Errorf("unexpected stats payload: %+v", s)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("/stats must not send Access-Control-Allow-Origin")
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", rec.Header().Get("Cache-Control"))
	}
}

func TestStaticFiles(t *testing.T) {
	for path, want := range map[string]string{"/cc-brand.css": "text/css", "/app.js": "text/javascript"} {
		rec := do(t, http.MethodGet, path)
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", path, rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, want) {
			t.Errorf("GET %s Content-Type = %q, want %s", path, ct, want)
		}
	}
}

func TestSecurityHeaders(t *testing.T) {
	for _, path := range []string{"/", "/stats", "/nope"} {
		h := do(t, http.MethodGet, path).Header()
		for name, want := range map[string]string{
			"X-Content-Type-Options": "nosniff",
			"Referrer-Policy":        "strict-origin-when-cross-origin",
			"X-Frame-Options":        "DENY",
		} {
			if got := h.Get(name); got != want {
				t.Errorf("GET %s %s = %q, want %q", path, name, got, want)
			}
		}
		if csp := h.Get("Content-Security-Policy"); !strings.Contains(csp, "script-src 'self'") {
			t.Errorf("GET %s CSP = %q, want script-src 'self'", path, csp)
		}
	}
}

func TestStatsHistory(t *testing.T) {
	record(readSample())
	var s statsResponse
	if err := json.NewDecoder(do(t, http.MethodGet, "/stats?history=1").Body).Decode(&s); err != nil {
		t.Fatal(err)
	}
	if len(s.History) == 0 || len(s.History) > historySize {
		t.Errorf("history has %d samples, want 1..%d", len(s.History), historySize)
	}
}

func TestHistoryRingIsBounded(t *testing.T) {
	for range historySize + 20 {
		record(sample{At: 1})
	}
	if n := len(history()); n != historySize {
		t.Fatalf("ring holds %d samples, want %d", n, historySize)
	}
}

func TestCutRunes(t *testing.T) {
	if got := cut("", 5); got != "" {
		t.Errorf(`cut("") = %q, want ""`, got)
	}
	if got := cut("héllo wörld", 5); got != "héllo" {
		t.Errorf("cut = %q, want héllo (rune-based)", got)
	}
}

// env builds a getenv function from a map, for the config readers.
func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestDemoConfigDefaults(t *testing.T) {
	c := loadDemoConfig(env(nil))
	want := demoConfig{Env: "local", Announcement: "", Burst: 500, HeapAlertMB: 32}
	if c != want {
		t.Fatalf("defaults = %+v, want %+v", c, want)
	}
	if c.IsProduction() {
		t.Error("local must not be production")
	}
}

func TestDemoConfigValues(t *testing.T) {
	c := loadDemoConfig(env(map[string]string{
		"APP_ENV":          " Production ",
		"ANNOUNCEMENT":     "  Maintenance ce soir à 22 h  ",
		"BURST_GOROUTINES": " 1500 ",
		"HEAP_ALERT_MB":    "128",
	}))
	if c.Env != "Production" || !c.IsProduction() {
		t.Errorf("APP_ENV = %q (production %v)", c.Env, c.IsProduction())
	}
	if c.Announcement != "Maintenance ce soir à 22 h" {
		t.Errorf("ANNOUNCEMENT = %q", c.Announcement)
	}
	if c.Burst != 1500 || c.HeapAlertMB != 128 {
		t.Errorf("Burst = %d, HeapAlertMB = %d", c.Burst, c.HeapAlertMB)
	}
}

func TestDemoConfigInvalidFallsBack(t *testing.T) {
	for _, v := range []string{"abc", "-5", "0", "1e3", "12.5", "999999", "  "} {
		c := loadDemoConfig(env(map[string]string{"BURST_GOROUTINES": v, "HEAP_ALERT_MB": v, "APP_ENV": "   "}))
		if c.Burst != defaultBurst || c.HeapAlertMB != defaultHeapAlertMB || c.Env != "local" {
			t.Errorf("value %q: got %+v, want defaults", v, c)
		}
	}
	if c := loadDemoConfig(env(map[string]string{"BURST_GOROUTINES": "2001", "HEAP_ALERT_MB": "1025"})); c.Burst != defaultBurst || c.HeapAlertMB != defaultHeapAlertMB {
		t.Errorf("above max: got %+v, want defaults", c)
	}
	if c := loadDemoConfig(env(map[string]string{"BURST_GOROUTINES": "2000", "HEAP_ALERT_MB": "1"})); c.Burst != 2000 || c.HeapAlertMB != 1 {
		t.Errorf("bounds must be accepted: got %+v", c)
	}
	long := strings.Repeat("é", 300)
	if c := loadDemoConfig(env(map[string]string{"ANNOUNCEMENT": long})); len([]rune(c.Announcement)) != maxAnnouncement {
		t.Errorf("ANNOUNCEMENT not capped: %d runes", len([]rune(c.Announcement)))
	}
}

func TestPlatform(t *testing.T) {
	if p := loadPlatform(env(nil)); p.Live || p.AppName != "" || p.Instance != "" || p.Commit != "" {
		t.Errorf("local platform = %+v, want empty", p)
	}
	p := loadPlatform(env(map[string]string{
		"APP_ID": "app_123", "CC_APP_NAME": "demo-go", "INSTANCE_NUMBER": "0",
		"CC_PRETTY_INSTANCE_NAME": "Brave Gopher", "CC_COMMIT_ID": "0123456789abcdef",
		"CC_DEPLOYMENT_ID": "deployment_0123456789abcdef",
	}))
	if !p.Live || p.Instance != "#0 · Brave Gopher" || p.Commit != "0123456" || p.Deployment != "deployment_01234" {
		t.Errorf("platform = %+v", p)
	}
}

func TestIndexShowsDemoVariables(t *testing.T) {
	t.Setenv("ANNOUNCEMENT", "Maintenance <b>ce soir</b>")
	t.Setenv("APP_ENV", "production")
	t.Setenv("BURST_GOROUTINES", "750")
	t.Setenv("HEAP_ALERT_MB", "48")
	body := do(t, http.MethodGet, "/").Body.String()
	for _, want := range []string{
		`class="go-announce"`, "Maintenance &lt;b&gt;ce soir&lt;/b&gt;", // escaped, never raw HTML
		`go-env is-production`, "Lancer 750 goroutines", `data-alert-mb="48"`, "Heap au-dessus de 48 MB",
		`href="#cert-title"`, `id="cert-title"`, "https://academy.clever.cloud/course/section.php?id=18",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("page is missing %q", want)
		}
	}
}

func TestIndexDefaults(t *testing.T) {
	for _, k := range []string{"ANNOUNCEMENT", "APP_ENV", "BURST_GOROUTINES", "HEAP_ALERT_MB", "APP_ID"} {
		t.Setenv(k, "")
	}
	body := do(t, http.MethodGet, "/").Body.String()
	if strings.Contains(body, `class="go-announce"`) {
		t.Error("no banner expected without ANNOUNCEMENT")
	}
	for _, want := range []string{">local<", "Lancer 500 goroutines", "injecté au déploiement", "Local · hors Clever Cloud"} {
		if !strings.Contains(body, want) {
			t.Errorf("page is missing %q", want)
		}
	}
	if strings.Count(body, `id="ccBadgeGrad"`) != 1 {
		t.Error("the badge SVG (with ids) must appear exactly once")
	}
}

func TestLoadEndpointsAreBounded(t *testing.T) {
	t.Setenv("BURST_GOROUTINES", "50")
	for _, kind := range []string{"goroutines", "heap", "gc"} {
		first := do(t, http.MethodPost, "/load/"+kind)
		if first.Code != http.StatusAccepted {
			t.Fatalf("POST /load/%s = %d, want 202", kind, first.Code)
		}
		second := do(t, http.MethodPost, "/load/"+kind)
		if second.Code != http.StatusTooManyRequests || second.Header().Get("Retry-After") == "" {
			t.Errorf("second POST /load/%s = %d, want 429 with Retry-After", kind, second.Code)
		}
		if rec := do(t, http.MethodGet, "/load/"+kind); rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("GET /load/%s = %d, want 405", kind, rec.Code)
		}
	}
	var s statsResponse
	if err := json.NewDecoder(do(t, http.MethodGet, "/stats").Body).Decode(&s); err != nil {
		t.Fatal(err)
	}
	if s.BurstLeftMs <= 0 || s.HeapLeftMs <= 0 {
		t.Errorf("bursts not reported in /stats: %+v", s)
	}
}

func TestFormatting(t *testing.T) {
	for d, want := range map[time.Duration]string{42 * time.Second: "42s", 187 * time.Second: "3m 07s", 7500 * time.Second: "2h 05m"} {
		if got := fmtUptime(d); got != want {
			t.Errorf("fmtUptime(%v) = %q, want %q", d, got, want)
		}
	}
	if got := (sample{PauseUs: 61.4}).PauseText(); got != "61 µs" {
		t.Errorf("PauseText = %q", got)
	}
	if got := (sample{}).PauseText(); got != "—" {
		t.Errorf("PauseText(0) = %q", got)
	}
	if got := (sample{}).GCText(); got != "aucun" {
		t.Errorf("GCText(0) = %q", got)
	}
	if got := (sample{PauseUs: 2300}).PauseText(); got != "2.3 ms" {
		t.Errorf("PauseText = %q", got)
	}
}
