package attachments

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
)

const (
	MaxSize   = 25 * 1024 * 1024 // 25MB threshold for chunking
	ChunkSize = 5 * 1024 * 1024  // 5MB chunks when size exceeds MaxSize
)

// Metadata describes a stored attachment.
type Metadata struct {
	ID          string // Temporary reference token
	FileName    string
	Size        int64
	Checksum    string
	ContentType string
	Chunks      []string // Keys of stored chunks in order
}

// ObjectStorage defines minimal storage behavior required by the processor.
type ObjectStorage interface {
	Save(ctx context.Context, key string, r io.Reader) error
}

// LocalStorage is a simple filesystem-backed implementation of ObjectStorage.
type LocalStorage struct {
	Dir string
}

// Save writes the contents of r to a file relative to the storage directory.
func (l *LocalStorage) Save(ctx context.Context, key string, r io.Reader) error {
	path := filepath.Join(l.Dir, key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

// Process reads MIME data from r, storing any attachments using the provided
// storage backend. It returns metadata for each stored attachment.
func Process(r io.Reader, store ObjectStorage) ([]Metadata, error) {
	msg, err := mail.ReadMessage(r)
	if err != nil {
		return nil, err
	}
	ctype := msg.Header.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(ctype)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(mediaType, "multipart/") {
		return nil, fmt.Errorf("unsupported content type: %s", mediaType)
	}

	mr := multipart.NewReader(msg.Body, params["boundary"])
	var metas []Metadata
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if part.FileName() == "" {
			continue // skip non attachment
		}
		meta, err := processPart(part, store)
		if err != nil {
			return nil, err
		}
		metas = append(metas, meta)
	}
	return metas, nil
}

func processPart(p *multipart.Part, store ObjectStorage) (Metadata, error) {
	var meta Metadata
	id, err := newID()
	if err != nil {
		return meta, err
	}
	meta.ID = id
	meta.FileName = p.FileName()
	meta.ContentType = p.Header.Get("Content-Type")

	data, err := io.ReadAll(p)
	if err != nil {
		return meta, err
	}
	meta.Size = int64(len(data))
	sum := sha256.Sum256(data)
	meta.Checksum = hex.EncodeToString(sum[:])

	ctx := context.Background()
	if meta.Size <= MaxSize {
		if err := store.Save(ctx, id, bytes.NewReader(data)); err != nil {
			return meta, err
		}
		meta.Chunks = []string{id}
		return meta, nil
	}

	// Chunking
	for off, idx := 0, 0; off < len(data); off += ChunkSize {
		end := off + ChunkSize
		if end > len(data) {
			end = len(data)
		}
		chunkKey := fmt.Sprintf("%s-%d", id, idx)
		if err := store.Save(ctx, chunkKey, bytes.NewReader(data[off:end])); err != nil {
			return meta, err
		}
		meta.Chunks = append(meta.Chunks, chunkKey)
		idx++
	}
	return meta, nil
}

func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
