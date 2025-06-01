package main

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
)

// StoreAttachment saves the attachment reader to disk and returns the path
func StoreAttachment(cfg Config, filename string, r io.Reader) (string, error) {
    path := filepath.Join(cfg.StorageDir, filename)
    f, err := os.Create(path)
    if err != nil {
        return "", fmt.Errorf("create file: %w", err)
    }
    defer f.Close()
    if _, err := io.Copy(f, r); err != nil {
        return "", fmt.Errorf("write file: %w", err)
    }
    return path, nil
}

// TODO: Implement more robust storage (e.g., S3)
