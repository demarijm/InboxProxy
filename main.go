package main

import (
	"bytes"
	"encoding/json"
	"expvar"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-smtp"
)

var (
	emailsReceived = expvar.NewInt("emails_received")
)

type backend struct{}

func (b *backend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	return nil, smtp.ErrAuthUnsupported
}

func (b *backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	return &session{start: time.Now()}, nil
}

type session struct {
	from  string
	rcpts []string
	start time.Time
}

func (s *session) Mail(from string, _ *smtp.MailOptions) error {
	s.from = from
	return nil
}

func (s *session) Rcpt(to string) error {
	s.rcpts = append(s.rcpts, to)
	return nil
}

func (s *session) Data(r io.Reader) error {
	maxSize := getMaxSize()
	data, err := io.ReadAll(io.LimitReader(r, int64(maxSize)))
	if err != nil {
		return err
	}

	id := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(1000))
	dir := filepath.Join(getStorageDir(), id)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dir, "raw.eml"), data, 0644); err != nil {
		return err
	}

	meta, err := parseMessage(bytes.NewReader(data))
	if err != nil {
		return err
	}
	meta.ID = id
	meta.From = s.from
	meta.To = s.rcpts

	b, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), b, 0644); err != nil {
		return err
	}

	emailsReceived.Add(1)
	log.Printf("stored email %s from %s", id, s.from)
	return nil
}

func (s *session) Reset() {}

func (s *session) Logout() error { return nil }

type metadata struct {
	ID          string       `json:"id"`
	From        string       `json:"from"`
	To          []string     `json:"to"`
	Subject     string       `json:"subject"`
	TextBody    string       `json:"text_body,omitempty"`
	HTMLBody    string       `json:"html_body,omitempty"`
	Attachments []attachment `json:"attachments,omitempty"`
}

type attachment struct {
	FileName string `json:"file_name"`
	Size     int64  `json:"size"`
	Path     string `json:"path"`
}

func parseMessage(r io.Reader) (*metadata, error) {
	msg, err := mail.ReadMessage(r)
	if err != nil {
		return nil, err
	}

	meta := &metadata{Subject: msg.Header.Get("Subject")}

	mr, err := mail.CreateReader(msg)
	if err != nil {
		return nil, err
	}

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			ct, _, _ := h.ContentType()
			b, _ := io.ReadAll(p.Body)
			if strings.HasPrefix(ct, "text/plain") {
				meta.TextBody = string(b)
			} else if strings.HasPrefix(ct, "text/html") {
				meta.HTMLBody = string(b)
			}
		case *mail.AttachmentHeader:
			name, _ := h.Filename()
			outPath := filepath.Join(getStorageDir(), meta.ID, name)
			f, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, err
			}
			n, err := io.Copy(f, p.Body)
			f.Close()
			if err != nil {
				return nil, err
			}
			meta.Attachments = append(meta.Attachments, attachment{FileName: name, Size: n, Path: outPath})
		}
	}

	return meta, nil
}

func getStorageDir() string {
	dir := os.Getenv("STORAGE_DIR")
	if dir == "" {
		dir = "./data"
	}
	return dir
}

func getMaxSize() int {
	v := os.Getenv("MAX_MESSAGE_SIZE")
	if v == "" {
		return 10 << 20 // 10MB
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 10 << 20
	}
	return n
}

func main() {
	rand.Seed(time.Now().UnixNano())

	if err := os.MkdirAll(getStorageDir(), 0755); err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	b := &backend{}
	s := smtp.NewServer(b)
	s.Addr = ":" + getEnv("SMTP_PORT", "2525")
	s.Domain = "localhost"
	s.ReadTimeout = 10 * time.Minute
	s.WriteTimeout = 10 * time.Minute
	s.MaxMessageBytes = getMaxSize()
	s.AllowInsecureAuth = true

	go func() {
		httpAddr := ":" + getEnv("METRICS_PORT", "8080")
		log.Printf("metrics listening on %s", httpAddr)
		if err := http.ListenAndServe(httpAddr, nil); err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("SMTP server listening on %s", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
