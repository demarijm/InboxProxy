package queue

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Job represents a unit of work to be processed by the queue.
type Job interface {
	ID() string
	Run() error
}

// Status holds metadata about a job's execution state.
type Status struct {
	State     string
	Attempts  int
	LastError string
}

// Queue manages asynchronous job execution with retry and failure logging.
type Queue struct {
	jobs       chan Job
	wg         sync.WaitGroup
	statuses   map[string]*Status
	mu         sync.Mutex
	maxRetries int
	failureLog *os.File
}

// New creates a new Queue.
func New(workerCount, maxRetries int, failureLogPath string) (*Queue, error) {
	logFile, err := os.OpenFile(failureLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	q := &Queue{
		jobs:       make(chan Job, workerCount*2),
		statuses:   make(map[string]*Status),
		maxRetries: maxRetries,
		failureLog: logFile,
	}
	for i := 0; i < workerCount; i++ {
		go q.worker()
	}
	return q, nil
}

// Enqueue adds a job to the queue.
func (q *Queue) Enqueue(j Job) {
	q.mu.Lock()
	q.statuses[j.ID()] = &Status{State: "pending"}
	q.mu.Unlock()
	q.wg.Add(1)
	q.jobs <- j
}

// Wait waits for all jobs to finish processing.
func (q *Queue) Wait() {
	q.wg.Wait()
}

// Status returns the current status of a job, if known.
func (q *Queue) Status(id string) (Status, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	s, ok := q.statuses[id]
	if !ok {
		return Status{}, false
	}
	return *s, true
}

func (q *Queue) update(id string, f func(*Status)) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if s, ok := q.statuses[id]; ok {
		f(s)
	}
}

func (q *Queue) worker() {
	for j := range q.jobs {
		id := j.ID()
		q.update(id, func(s *Status) { s.State = "running" })
		var err error
		attempts := 0
		for {
			attempts++
			err = j.Run()
			if err == nil || attempts > q.maxRetries {
				break
			}
			time.Sleep(time.Second)
		}
		if err != nil {
			q.update(id, func(s *Status) {
				s.State = "failed"
				s.Attempts = attempts
				s.LastError = err.Error()
			})
			fmt.Fprintf(q.failureLog, "%s,%s\n", id, err)
		} else {
			q.update(id, func(s *Status) {
				s.State = "done"
				s.Attempts = attempts
			})
		}
		q.wg.Done()
	}
}

// Close stops accepting new jobs and waits for workers to finish.
func (q *Queue) Close() {
	close(q.jobs)
	q.wg.Wait()
	q.failureLog.Close()
}
