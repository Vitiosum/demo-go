package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
	if s.GoVersion == "" || s.Goroutines == 0 {
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

func TestCutRunes(t *testing.T) {
	if got := cut("", 5); got != "—" {
		t.Errorf(`cut("") = %q, want "—"`, got)
	}
	if got := cut("héllo wörld", 5); got != "héllo" {
		t.Errorf("cut = %q, want héllo (rune-based)", got)
	}
}
