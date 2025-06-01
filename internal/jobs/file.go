package jobs

import (
	"crypto/rand"
	"encoding/hex"
	"os"
)

// SaveFileJob writes data to a specified path.
// It implements the queue.Job interface.
type SaveFileJob struct {
	IDVal string
	Path  string
	Data  []byte
}

func NewSaveFileJob(path string, data []byte) *SaveFileJob {
	id := make([]byte, 16)
	rand.Read(id)
	return &SaveFileJob{
		IDVal: hex.EncodeToString(id),
		Path:  path,
		Data:  data,
	}
}

func (j *SaveFileJob) ID() string { return j.IDVal }

func (j *SaveFileJob) Run() error {
	return os.WriteFile(j.Path, j.Data, 0644)
}
