package logging

import (
	"encoding/json"
	"log"
	"time"
)

// Logger provides structured JSON logging.
type Logger struct {
	Service string
}

// New creates a new Logger for a service.
func New(service string) *Logger {
	return &Logger{Service: service}
}

func (l *Logger) log(level, msg string, fields map[string]interface{}) {
	entry := map[string]interface{}{
		"time":    time.Now().Format(time.RFC3339),
		"level":   level,
		"service": l.Service,
		"msg":     msg,
	}
	for k, v := range fields {
		entry[k] = v
	}
	b, _ := json.Marshal(entry)
	log.Print(string(b))
}

// Info logs a message at info level.
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.log("info", msg, fields)
}

// Error logs a message at error level.
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.log("error", msg, fields)
}
