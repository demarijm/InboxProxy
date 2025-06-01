package main

import (
    "log"
    "os"
    "strconv"
)

type Config struct {
    ListenAddr      string
    MaxMessageBytes int
    StorageDir      string
}

func LoadConfig() Config {
    cfg := Config{
        ListenAddr:      getEnv("SMTP_ADDR", ":2525"),
        MaxMessageBytes: getEnvInt("MAX_MESSAGE_BYTES", 10<<20),
        StorageDir:      getEnv("STORAGE_DIR", "./data"),
    }
    if err := os.MkdirAll(cfg.StorageDir, 0o755); err != nil {
        log.Fatalf("unable to create storage dir: %v", err)
    }
    return cfg
}

func getEnv(key, def string) string {
    v := os.Getenv(key)
    if v == "" {
        return def
    }
    return v
}

func getEnvInt(key string, def int) int {
    v := os.Getenv(key)
    if v == "" {
        return def
    }
    i, err := strconv.Atoi(v)
    if err != nil {
        log.Printf("invalid int for %s: %v", key, err)
        return def
    }
    return i
}
