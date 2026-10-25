package main

import (
	"os"
	"strconv"
	"strings"
)

// platform is what Clever Cloud injects into the environment of a running
// instance. A field is empty outside the platform; the page then shows
// « injecté au déploiement ».
// Reference: https://www.clever.cloud/developers/doc/develop/common-configuration/environment-variables/reference/
type platform struct {
	Live         bool
	AppName      string
	AppID        string
	Instance     string
	InstanceType string
	Commit       string
	Deployment   string
}

// demoConfig holds the demo settings, read from the environment variables tab
// of the Clever Cloud console. Each one has a visible effect on the page.
type demoConfig struct {
	Env          string // APP_ENV: environment pill, Clever gradient when "production"
	Announcement string // ANNOUNCEMENT: banner on top of the page when not empty
	Burst        int    // BURST_GOROUTINES: goroutines started by the « goroutines » button
	HeapAlertMB  int    // HEAP_ALERT_MB: heap threshold drawn on the chart, card turns red above it
}

const (
	defaultEnv         = "local"
	defaultBurst       = 500
	maxBurst           = 2000
	defaultHeapAlertMB = 32
	maxHeapAlertMB     = 1024
	maxAnnouncement    = 160
)

// loadDemoConfig reads the demo variables through getenv (os.Getenv in
// production, a map in tests). Any invalid or out-of-range value falls back to
// its default: a typo in the console never breaks the page.
func loadDemoConfig(getenv func(string) string) demoConfig {
	c := demoConfig{
		Env:          cut(strings.TrimSpace(getenv("APP_ENV")), 24),
		Announcement: cut(strings.TrimSpace(getenv("ANNOUNCEMENT")), maxAnnouncement),
		Burst:        intInRange(getenv("BURST_GOROUTINES"), 1, maxBurst, defaultBurst),
		HeapAlertMB:  intInRange(getenv("HEAP_ALERT_MB"), 1, maxHeapAlertMB, defaultHeapAlertMB),
	}
	if c.Env == "" {
		c.Env = defaultEnv
	}
	return c
}

// IsProduction drives the gradient environment pill.
func (c demoConfig) IsProduction() bool { return strings.EqualFold(c.Env, "production") }

// intInRange parses s as a base-10 integer within [lo, hi], or returns def.
func intInRange(s string, lo, hi, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < lo || n > hi {
		return def
	}
	return n
}

// cut returns s truncated to n runes, never in the middle of a UTF-8 sequence.
func cut(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}

// firstEnv returns the first non-empty value among the given variable names.
// The deployed-commit variable is only populated when the deployment carries a
// commit (git push); a `clever restart` has none, and the page says so.
func firstEnv(getenv func(string) string, keys ...string) string {
	for _, k := range keys {
		if v := getenv(k); v != "" {
			return v
		}
	}
	return ""
}

// loadPlatform reads the variables Clever Cloud sets on each instance.
func loadPlatform(getenv func(string) string) platform {
	inst := ""
	if n := getenv("INSTANCE_NUMBER"); n != "" {
		inst = "#" + cut(n, 4)
		if p := getenv("CC_PRETTY_INSTANCE_NAME"); p != "" {
			inst += " · " + cut(p, 40)
		}
	}
	return platform{
		Live:         getenv("APP_ID") != "",
		AppName:      cut(getenv("CC_APP_NAME"), 40),
		AppID:        cut(getenv("APP_ID"), 40),
		Instance:     inst,
		InstanceType: cut(getenv("INSTANCE_TYPE"), 20),
		Commit:       cut(firstEnv(getenv, "COMMIT_ID", "CC_COMMIT_ID"), 7),
		Deployment:   cut(firstEnv(getenv, "CC_DEPLOYMENT_ID", "DEPLOYMENT_ID"), 16),
	}
}

// listenPort returns the port injected by Clever Cloud, or 8080 locally.
func listenPort() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}
