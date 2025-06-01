package main

import (
    "bytes"
    "io"

    mailmsg "github.com/emersion/go-message/mail"
)

// Email represents a parsed email message.
type Email struct {
    From        string
    To          []string
    Subject     string
    Text        string
    HTML        string
    Attachments []Attachment
    Raw         []byte
}

// Attachment metadata
type Attachment struct {
    Filename    string
    ContentType string
    Size        int64
    Path        string
}

// ParseEmail reads the raw message bytes and extracts common fields.
func ParseEmail(raw []byte, cfg Config) (*Email, error) {
    var email Email
    email.Raw = raw

    mr, err := mailmsg.NewReader(bytes.NewReader(raw))
    if err != nil {
        return nil, err
    }
    hdr := mr.Header
    if from, err := hdr.AddressList("From"); err == nil && len(from) > 0 {
        email.From = from[0].String()
    }
    if to, err := hdr.AddressList("To"); err == nil {
        for _, a := range to {
            email.To = append(email.To, a.String())
        }
    }
    if subj, err := hdr.Subject(); err == nil {
        email.Subject = subj
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
        case *mailmsg.InlineHeader:
            ct, _, _ := h.ContentType()
            b, _ := io.ReadAll(p.Body)
            if ct == "text/plain" {
                email.Text = string(b)
            } else if ct == "text/html" {
                email.HTML = string(b)
            }
        case *mailmsg.AttachmentHeader:
            ct, _, _ := h.ContentType()
            filename, _ := h.Filename()
            path, err := StoreAttachment(cfg, filename, p.Body)
            if err != nil {
                return nil, err
            }
            size := int64(-1)
            if lr, ok := p.Body.(interface{ Len() int }); ok {
                size = int64(lr.Len())
            }
            email.Attachments = append(email.Attachments, Attachment{
                Filename:    filename,
                ContentType: ct,
                Size:        size,
                Path:        path,
            })
        }
    }

    return &email, nil
}
