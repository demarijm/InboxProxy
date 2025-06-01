package main

import (
    "io"
    "log"

    smtp "github.com/emersion/go-smtp"
)

type Backend struct{}

func (b *Backend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
    return nil, smtp.ErrAuthUnsupported
}

func (b *Backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
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

func (s *Session) Data(r io.Reader) error {
    raw, err := io.ReadAll(r)
    if err != nil {
        return err
    }
    cfg := LoadConfig()
    email, err := ParseEmail(raw, cfg)
    if err != nil {
        return err
    }
    log.Printf("received mail from %s to %v subject=%s", email.From, email.To, email.Subject)
    return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error { return nil }

func NewSMTPServer(cfg Config) *smtp.Server {
    backend := &Backend{}
    s := smtp.NewServer(backend)
    s.Addr = cfg.ListenAddr
    s.Domain = "localhost"
    s.MaxMessageBytes = cfg.MaxMessageBytes
    s.AllowInsecureAuth = true
    return s
}
