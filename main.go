package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Config holds service configuration.
type Config struct {
	Port string `json:"port"`
}

var (
	cfg           Config
	cfgMu         sync.RWMutex
	requestsTotal int64
)

// loadConfig reads configuration from file specified by CONFIG_FILE or config.json.
func loadConfig() (Config, error) {
	file := os.Getenv("CONFIG_FILE")
	if file == "" {
		file = "config.json"
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return Config{}, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return Config{}, err
	}
	if c.Port == "" {
		c.Port = "8080"
	}
	return c, nil
}

// reloadConfig loads configuration and updates the global cfg.
func reloadConfig() {
	c, err := loadConfig()
	if err != nil {
		logJSON("error", "failed to reload config", map[string]interface{}{"error": err.Error()})
		return
	}
	cfgMu.Lock()
	cfg = c
	cfgMu.Unlock()
	logJSON("info", "config reloaded", nil)
}

// logJSON writes structured logs to stdout.
func logJSON(level, msg string, fields map[string]interface{}) {
	entry := map[string]interface{}{
		"level": level,
		"msg":   msg,
		"time":  time.Now().Format(time.RFC3339),
	}
	for k, v := range fields {
		entry[k] = v
	}
	b, _ := json.Marshal(entry)
	fmt.Println(string(b))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&requestsTotal, 1)
	logJSON("info", "health check", map[string]interface{}{"path": r.URL.Path})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&requestsTotal, 1)
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# HELP http_requests_total The total number of HTTP requests.\n")
	fmt.Fprintf(w, "# TYPE http_requests_total counter\n")
	fmt.Fprintf(w, "http_requests_total %d\n", atomic.LoadInt64(&requestsTotal))
}

func main() {
	reloadConfig()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP)
	go func() {
		for range signalChan {
			reloadConfig()
		}
	}()

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/metrics", metricsHandler)

	cfgMu.RLock()
	port := cfg.Port
	cfgMu.RUnlock()
	logJSON("info", "starting server", map[string]interface{}{"port": port})
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logJSON("error", "server stopped", map[string]interface{}{"error": err.Error()})
	}
}
