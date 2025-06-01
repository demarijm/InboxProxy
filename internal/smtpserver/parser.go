package parser

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/emersion/go-message/mail"
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

func ParseEmail(data []byte, dir string) (*ParsedEmail, error) {
	mr, err := mail.CreateReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	header := mr.Header
	from, _ := header.AddressList("From")
	to, _ := header.AddressList("To")
	subject, _ := header.Subject()

	email := &ParsedEmail{
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
			return nil, err
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
				return nil, err
			}
			email.Attachments = append(email.Attachments, Attachment{Filename: filename, ContentType: ctype})
		}
	}

	return email, nil
}

func addressesToString(addrs []*mail.Address) string {
	var parts []string
	for _, a := range addrs {
		parts = append(parts, a.String())
	}
	return join(parts, ", ")
}

func join(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	res := strs[0]
	for _, s := range strs[1:] {
		res += sep + s
	}
	return res
}
