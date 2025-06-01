package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// Email represents a captured email
type Email struct {
	ID      int      `json:"id"`
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	Files   []File   `json:"files"`
}

// File represents an attachment
type File struct {
	ID       int    `json:"id"`
	Filename string `json:"filename"`
	Data     []byte `json:"-"`
}

// Hook represents a registered hook
type Hook struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

var (
	emailStore = struct {
		sync.RWMutex
		emails []Email
	}{}

	fileStore = struct {
		sync.RWMutex
		files map[int]File
	}{files: make(map[int]File)}

	hookStore = struct {
		sync.RWMutex
		hooks []Hook
	}{}

	metrics = struct {
		sync.Mutex
		counters map[string]*uint64
	}{counters: make(map[string]*uint64)}
)

func metricInc(name string) {
	metrics.Lock()
	c, ok := metrics.counters[name]
	if !ok {
		var v uint64
		metrics.counters[name] = &v
		c = &v
	}
	metrics.Unlock()
	atomic.AddUint64(c, 1)
}

func main() {
	// Populate with a sample email and file
	fileStore.Lock()
	fileStore.files[1] = File{ID: 1, Filename: "sample.txt", Data: []byte("sample content")}
	fileStore.Unlock()

	emailStore.Lock()
	emailStore.emails = append(emailStore.emails, Email{
		ID:      1,
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Hello",
		Body:    "This is a test email",
		Files:   []File{{ID: 1, Filename: "sample.txt"}},
	})
	emailStore.Unlock()

	mux := http.NewServeMux()
	mux.HandleFunc("/emails", authMiddleware(handleListEmails))
	mux.HandleFunc("/emails/", authMiddleware(handleGetEmail))
	mux.HandleFunc("/hooks", authMiddleware(handleRegisterHook))
	mux.HandleFunc("/files/", authMiddleware(handleGetFile))
	mux.HandleFunc("/metrics", handleMetrics)

	addr := ":8080"
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, logMiddleware(mux)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func token() string {
	t := os.Getenv("API_TOKEN")
	if t == "" {
		t = "secret-token"
	}
	return t
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	expected := token()
	return func(w http.ResponseWriter, r *http.Request) {
		auth := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if auth != expected {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func handleListEmails(w http.ResponseWriter, r *http.Request) {
	metricInc("list_emails")
	emailStore.RLock()
	defer emailStore.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(emailStore.emails)
}

func handleGetEmail(w http.ResponseWriter, r *http.Request) {
	metricInc("get_email")
	idStr := strings.TrimPrefix(r.URL.Path, "/emails/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	emailStore.RLock()
	defer emailStore.RUnlock()
	for _, e := range emailStore.emails {
		if e.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(e)
			return
		}
	}
	http.NotFound(w, r)
}

func handleRegisterHook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	metricInc("register_hook")
	var h Hook
	if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	hookStore.Lock()
	h.ID = len(hookStore.hooks) + 1
	hookStore.hooks = append(hookStore.hooks, h)
	hookStore.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(h)
}

func handleGetFile(w http.ResponseWriter, r *http.Request) {
	metricInc("get_file")
	idStr := strings.TrimPrefix(r.URL.Path, "/files/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	fileStore.RLock()
	f, ok := fileStore.files[id]
	fileStore.RUnlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", f.Filename))
	w.Write(f.Data)
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics.Lock()
	defer metrics.Unlock()
	out := make(map[string]uint64)
	for k, v := range metrics.counters {
		out[k] = atomic.LoadUint64(v)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}
