package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"inboxproxy/internal/smtpserver"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	flag.Parse()
	addr := getenv("SMTP_ADDR", ":2525")
	storageDir := getenv("STORAGE_DIR", "data")
	maxSizeStr := getenv("MAX_FILE_SIZE", "10485760")
	maxSize, err := strconv.ParseInt(maxSizeStr, 10, 64)
	if err != nil {
		maxSize = 10 << 20
	}

	backend := smtpserver.NewBackend(storageDir, maxSize)

	s := smtpserver.NewServer(backend, addr)

	log.Printf("listening on %s", addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
