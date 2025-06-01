package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	smtp "github.com/emersion/go-smtp"
)

type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
}

type ParsedEmail struct {
	From        string       `json:"from"`
	To          string       `json:"to"`
	Subject     string       `json:"subject"`
	TextBody    string       `json:"text_body,omitempty"`
	HTMLBody    string       `json:"html_body,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

var totalEmails uint64

type Backend struct {
	storageDir  string
	maxFileSize int64
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
	atomic.AddUint64(&totalEmails, 1)

	id := time.Now().UnixNano()
	dir := filepath.Join(s.backend.storageDir, fmt.Sprintf("%d", id))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := io.ReadAll(io.LimitReader(r, s.backend.maxFileSize))
	if err != nil {
		return err
	}

	rawPath := filepath.Join(dir, "raw.eml")
	if err := os.WriteFile(rawPath, data, 0644); err != nil {
		return err
	}

	if err := parseAndStore(bytes.NewReader(data), dir); err != nil {
		return err
	}

	log.Printf("stored email %d from %s", id, s.from)
	return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error { return nil }

func parseAndStore(r io.Reader, dir string) error {
	mr, err := mail.CreateReader(r)
	if err != nil {
		return err
	}

	header := mr.Header
	from, _ := header.AddressList("From")
	to, _ := header.AddressList("To")
	subject, _ := header.Subject()

	email := ParsedEmail{
		From:    addressesToString(from),
		To:      addressesToString(to),
		Subject: subject,
	}

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			ctype, _, _ := h.ContentType()
			b, _ := io.ReadAll(p.Body)
			if ctype == "text/plain" {
				email.TextBody = string(b)
			} else if ctype == "text/html" {
				email.HTMLBody = string(b)
			}
		case *mail.AttachmentHeader:
			filename, _ := h.Filename()
			ctype, _, _ := h.ContentType()
			b, _ := io.ReadAll(p.Body)
			path := filepath.Join(dir, filename)
			if err := os.WriteFile(path, b, 0644); err != nil {
				return err
			}
			email.Attachments = append(email.Attachments, Attachment{Filename: filename, ContentType: ctype})
		}
	}

	metaBytes, err := json.MarshalIndent(email, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "meta.json"), metaBytes, 0644)
}

func addressesToString(addrs []*mail.Address) string {
	var parts []string
	for _, a := range addrs {
		parts = append(parts, a.String())
	}
	return strings.Join(parts, ", ")
}

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

	backend := &Backend{storageDir: storageDir, maxFileSize: maxSize}

	s := smtp.NewServer(backend)
	s.Addr = addr
	s.Domain = "localhost"
	s.AllowInsecureAuth = true

	log.Printf("listening on %s", addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
