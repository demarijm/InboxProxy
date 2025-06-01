package smtpserver

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	smtp "github.com/emersion/go-smtp"

	"inboxproxy/internal/smtpserver/parser"
	"inboxproxy/internal/storage"
)

type Backend struct {
	StorageDir  string
	MaxFileSize int64
}

func NewBackend(storageDir string, maxFileSize int64) *Backend {
	return &Backend{StorageDir: storageDir, MaxFileSize: maxFileSize}
}

func (b *Backend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	return &Session{backend: b}, nil
}

func (b *Backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	return &Session{backend: b}, nil
}

type Session struct {
	backend *Backend
	from    string
	to      []string
}

func (s *Session) Mail(from string, _ smtp.MailOptions) error {
	s.from = from
	return nil
}

func (s *Session) Rcpt(to string) error {
	s.to = append(s.to, to)
	return nil
}

func (s *Session) Data(r io.Reader) error {
	id := time.Now().UnixNano()
	dir := filepath.Join(s.backend.StorageDir, fmt.Sprintf("%d", id))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := io.ReadAll(io.LimitReader(r, s.backend.MaxFileSize))
	if err != nil {
		return err
	}

	rawPath := filepath.Join(dir, "raw.eml")
	if err := os.WriteFile(rawPath, data, 0644); err != nil {
		return err
	}

	parsed, err := parser.ParseEmail(data, dir)
	if err != nil {
		return err
	}
	if err := storage.SaveMetadata(parsed, dir); err != nil {
		return err
	}

	log.Printf("stored email %d from %s", id, s.from)
	return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error { return nil }

func NewServer(backend *Backend, addr string) *smtp.Server {
	s := smtp.NewServer(backend)
	s.Addr = addr
	s.Domain = "localhost"
	s.AllowInsecureAuth = true
	return s
}
