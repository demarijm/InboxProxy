package hooks

import (
	"encoding/json"
	"os"
)

// Action describes what to do when a hook matches.
type Action struct {
	Type    string `json:"type"` // "http_post" or "command"
	URL     string `json:"url,omitempty"`
	Command string `json:"command,omitempty"`
}

// Hook defines matching rules and the action to execute.
type Hook struct {
	ID              string `json:"id"`
	Sender          string `json:"sender,omitempty"`
	SubjectContains string `json:"subject_contains,omitempty"`
	FileType        string `json:"file_type,omitempty"`
	Action          Action `json:"action"`
}

// EmailEvent represents minimal information about an email.
type EmailEvent struct {
	Sender   string
	Subject  string
	FileType string
}

// LoadHooks reads hook definitions from a JSON file.
func LoadHooks(path string) ([]Hook, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var hooks []Hook
	if err := json.Unmarshal(data, &hooks); err != nil {
		return nil, err
	}
	return hooks, nil
}
