package metrics

import (
	"fmt"
	"sync"
	"time"
)

var (
	mu       sync.Mutex
	counters = make(map[string]int)
	start    = time.Now()
)

// Inc increments the named counter.
func Inc(name string) {
	mu.Lock()
	counters[name]++
	mu.Unlock()
}

// Dump returns metrics in Prometheus format.
func Dump() string {
	mu.Lock()
	defer mu.Unlock()
	out := fmt.Sprintf("uptime_seconds %d\n", int(time.Since(start).Seconds()))
	for k, v := range counters {
		out += fmt.Sprintf("%s %d\n", k, v)
	}
	return out
}
