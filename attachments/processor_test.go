package attachments

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"testing"
)

func buildTestMessage() []byte {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	h1 := textproto.MIMEHeader{}
	h1.Set("Content-Type", "text/plain")
	h1.Set("Content-Disposition", "attachment; filename=\"hello.txt\"")
	p1, _ := w.CreatePart(h1)
	p1.Write([]byte("hello"))

	h2 := textproto.MIMEHeader{}
	h2.Set("Content-Type", "application/octet-stream")
	h2.Set("Content-Disposition", "attachment; filename=\"big.bin\"")
	p2, _ := w.CreatePart(h2)
	p2.Write(bytes.Repeat([]byte("a"), 26*1024*1024)) // 26MB

	w.Close()

	var msg bytes.Buffer
	msg.WriteString("Content-Type: multipart/mixed; boundary=" + w.Boundary() + "\r\n\r\n")
	msg.Write(body.Bytes())
	return msg.Bytes()
}

func TestProcess(t *testing.T) {
	msg := buildTestMessage()
	store := &LocalStorage{Dir: t.TempDir()}
	metas, err := Process(bytes.NewReader(msg), store)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}
	if len(metas) != 2 {
		t.Fatalf("expected 2 attachments, got %d", len(metas))
	}
	small := metas[0]
	big := metas[1]
	if len(small.Chunks) != 1 {
		t.Errorf("small attachment should have 1 chunk, got %d", len(small.Chunks))
	}
	if big.Size <= MaxSize || len(big.Chunks) <= 1 {
		t.Errorf("big attachment should be chunked; size=%d chunks=%d", big.Size, len(big.Chunks))
	}
}
