package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"inboxproxy/internal/smtpserver/parser"
)

func SaveMetadata(email *parser.ParsedEmail, dir string) error {
	metaBytes, err := json.MarshalIndent(email, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "meta.json"), metaBytes, 0644)
}
