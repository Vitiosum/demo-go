package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/", indexPage)
	http.HandleFunc("/health", healthCheck)

	http.ListenAndServe("0.0.0.0:8080", nil)
}

func indexPage(w http.ResponseWriter, r *http.Request) {

	hostname, _ := os.Hostname()
	port := os.Getenv("PORT")
	instance := os.Getenv("INSTANCE_NUMBER")
	now := time.Now().Format(time.RFC1123)

	fmt.Fprintf(
		w, `<h1>Hello, Youtube</h1>
<p>Go is running on Clever Cloud 💡☁️</p>

<p><b>Hostname:</b> %s</p>
<p><b>Instance:</b> %s</p>
<p><b>Port:</b> %s</p>
<p><b>Server time:</b> %s</p>

<p>Try it yourself on <a href="https://www.clever.cloud">Clever Cloud</a> 🚀</p>`,
		hostname, instance, port, now,
	)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
