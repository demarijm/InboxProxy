package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
)

// HookRunner executes hooks in response to email events.
type HookRunner struct {
	hooks  []Hook
	events chan EmailEvent
	wg     sync.WaitGroup
}

// NewHookRunner creates a runner with the given hooks.
func NewHookRunner(hooks []Hook) *HookRunner {
	return &HookRunner{
		hooks:  hooks,
		events: make(chan EmailEvent, 10),
	}
}

// Start begins the event processing goroutine.
func (hr *HookRunner) Start() {
	hr.wg.Add(1)
	go hr.loop()
}

// Stop stops the runner and waits for completion.
func (hr *HookRunner) Stop() {
	close(hr.events)
	hr.wg.Wait()
}

// Process submits an email event for processing.
func (hr *HookRunner) Process(ev EmailEvent) {
	hr.events <- ev
}

// AddHook registers a new hook at runtime.
func (hr *HookRunner) AddHook(h Hook) {
	hr.hooks = append(hr.hooks, h)
}

func (hr *HookRunner) loop() {
	defer hr.wg.Done()
	for ev := range hr.events {
		for _, h := range hr.hooks {
			if !hr.match(h, ev) {
				continue
			}
			if err := hr.execute(h, ev); err != nil {
				log.Printf("hook %s failed: %v", h.ID, err)
			} else {
				log.Printf("hook %s executed successfully", h.ID)
			}
		}
	}
}

func (hr *HookRunner) match(h Hook, ev EmailEvent) bool {
	if h.Sender != "" && h.Sender != ev.Sender {
		return false
	}
	if h.SubjectContains != "" && !strings.Contains(ev.Subject, h.SubjectContains) {
		return false
	}
	if h.FileType != "" && h.FileType != ev.FileType {
		return false
	}
	return true
}

func (hr *HookRunner) execute(h Hook, ev EmailEvent) error {
	switch h.Action.Type {
	case "command":
		cmd := exec.Command("sh", "-c", h.Action.Command)
		return cmd.Run()
	case "http_post":
		payload, _ := json.Marshal(ev)
		resp, err := http.Post(h.Action.URL, "application/json", bytes.NewReader(payload))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			return fmt.Errorf("http %d", resp.StatusCode)
		}
		return nil
	default:
		return fmt.Errorf("unknown action type %s", h.Action.Type)
	}
}
