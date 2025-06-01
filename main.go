package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/demarijm/inboxproxy/config"
	"github.com/demarijm/inboxproxy/logging"
	"github.com/demarijm/inboxproxy/metrics"
)

func main() {
	cfgPath := flag.String("config", "", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	log := logging.New("inboxproxy")
	log.Info("starting server", map[string]interface{}{"addr": cfg.Addr})

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		metrics.Inc("health_requests_total")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprint(w, metrics.Dump())
	})

	server := &http.Server{Addr: cfg.Addr, Handler: loggingMiddleware(log, mux)}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for sig := range sigCh {
			switch sig {
			case syscall.SIGHUP:
				if err := cfg.Reload(); err != nil {
					log.Error("reload config", map[string]interface{}{"error": err.Error()})
				} else {
					log.Info("config reloaded", nil)
					server.Addr = cfg.Addr
				}
			default:
				log.Info("shutting down", map[string]interface{}{"signal": sig.String()})
				server.Close()
				return
			}
		}
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server error", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
}

func loggingMiddleware(log *logging.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggingResponseWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(lrw, r)
		metrics.Inc("http_requests_total")
		log.Info("request", map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
			"status": lrw.status,
		})
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
