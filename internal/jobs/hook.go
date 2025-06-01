package jobs

import (
	"crypto/rand"
	"encoding/hex"
)

// HookFunc represents a custom action to be executed by the job.
type HookFunc func() error

// HookJob executes an arbitrary HookFunc.
type HookJob struct {
	IDVal string
	Hook  HookFunc
}

func NewHookJob(fn HookFunc) *HookJob {
	id := make([]byte, 16)
	rand.Read(id)
	return &HookJob{IDVal: hex.EncodeToString(id), Hook: fn}
}

func (h *HookJob) ID() string { return h.IDVal }

func (h *HookJob) Run() error {
	if h.Hook != nil {
		return h.Hook()
	}
	return nil
}
