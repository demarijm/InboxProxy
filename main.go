package main

import (
    "log"
)

func main() {
    cfg := LoadConfig()
    server := NewSMTPServer(cfg)
    log.Printf("listening on %s", cfg.ListenAddr)
    if err := server.ListenAndServe(); err != nil {
        log.Fatalf("server failed: %v", err)
    }
}
