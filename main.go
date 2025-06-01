package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-smtp"
)

var (
	processed  uint64
	failed     uint64
	maxSize    int64
	storageDir string
)

type Backend struct{}

func (bkd *Backend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	return &Session{}, nil
}

func (bkd *Backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	return &Session{}, nil
}

type Session struct {
	from string
	to   []string
}

func (s *Session) Mail(from string, opts smtp.MailOptions) error {
	s.from = from
	return nil
}

func (s *Session) Rcpt(to string) error {
	s.to = append(s.to, to)
	return nil
}

func (s *Session) Reset()        {}
func (s *Session) Logout() error { return nil }

func (s *Session) Data(r io.Reader) error {
	limited := io.LimitReader(r, maxSize)
	raw, err := io.ReadAll(limited)
	if err != nil {
		atomic.AddUint64(&failed, 1)
		return err
	}

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	dir := filepath.Join(storageDir, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		atomic.AddUint64(&failed, 1)
		return err
	}

	if err := os.WriteFile(filepath.Join(dir, "raw.eml"), raw, 0o644); err != nil {
		atomic.AddUint64(&failed, 1)
		return err
	}

	meta, err := parseMessage(raw, dir)
	if err != nil {
		atomic.AddUint64(&failed, 1)
		return err
	}

	meta.From = s.from
	meta.To = append(meta.To, s.to...)
	metaBytes, _ := json.MarshalIndent(meta, "", "  ")
	os.WriteFile(filepath.Join(dir, "metadata.json"), metaBytes, 0o644)

	atomic.AddUint64(&processed, 1)
	return nil
}

// Attachment describes stored attachment metadata
type Attachment struct {
	FileName    string `json:"file_name"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

// Metadata represents parsed message metadata
type Metadata struct {
	From        string       `json:"from"`
	To          []string     `json:"to"`
	Subject     string       `json:"subject"`
	Text        string       `json:"text_body,omitempty"`
	HTML        string       `json:"html_body,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

func parseMessage(raw []byte, dir string) (*Metadata, error) {
	mr, err := message.Read(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}

	hdr := mail.Header{Header: mr.Header}
	subject, _ := hdr.Subject()

	meta := &Metadata{Subject: subject}

	contentType := mr.Header.Get("Content-Type")
	mediaType, _, _ := mime.ParseMediaType(contentType)

	if strings.HasPrefix(mediaType, "multipart/") {
		mpReader := mr.MultipartReader()
		for {
			p, err := mpReader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return meta, err
			}

			disp, _, _ := mime.ParseMediaType(p.Header.Get("Content-Disposition"))
			ctype := p.Header.Get("Content-Type")
			if disp == "attachment" {
				name := p.FileName()
				fpath := filepath.Join(dir, name)
				f, err := os.Create(fpath)
				if err != nil {
					continue
				}
				n, _ := io.Copy(f, p.Body)
				f.Close()
				meta.Attachments = append(meta.Attachments, Attachment{FileName: name, Size: n, ContentType: ctype})
			} else {
				b, _ := io.ReadAll(p.Body)
				if strings.HasPrefix(ctype, "text/html") {
					meta.HTML = string(b)
				} else if strings.HasPrefix(ctype, "text/plain") {
					meta.Text = string(b)
				}
			}
		}
	} else {
		b, _ := io.ReadAll(mr.Body)
		if strings.HasPrefix(mediaType, "text/html") {
			meta.HTML = string(b)
		} else {
			meta.Text = string(b)
		}
	}

	return meta, nil
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "processed_messages %d\nfailed_messages %d\n", atomic.LoadUint64(&processed), atomic.LoadUint64(&failed))
}

func main() {
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "2525"
	}

	sd := os.Getenv("STORAGE_DIR")
	if sd == "" {
		sd = "./data"
	}
	storageDir = sd

	sizeStr := os.Getenv("MAX_MESSAGE_SIZE")
	if sizeStr == "" {
		sizeStr = "10485760" // 10MB
	}
	sz, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil || sz <= 0 {
		sz = 10 << 20
	}
	maxSize = sz

	be := &Backend{}
	srv := smtp.NewServer(be)
	srv.Addr = ":" + port
	srv.Domain = "localhost"
	srv.MaxMessageBytes = int(maxSize)
	srv.ReadTimeout = 10 * time.Minute
	srv.WriteTimeout = 10 * time.Minute

	go func() {
		if err := http.ListenAndServe(":8080", http.HandlerFunc(metricsHandler)); err != nil {
			log.Printf("metrics server error: %v", err)
		}
	}()

	log.Printf("SMTP server listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("smtp server error: %v", err)
	}
}
