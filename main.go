package main

import (
	"fmt"
	"log"

	"github.com/demarijm/inboxproxy/internal/jobs"
	"github.com/demarijm/inboxproxy/internal/queue"
)

func main() {
	q, err := queue.New(2, 3, "failures.log")
	if err != nil {
		log.Fatalf("queue init: %v", err)
	}
	defer q.Close()

	fileJob := jobs.NewSaveFileJob("sample.txt", []byte("hello"))
	q.Enqueue(fileJob)

	hookJob := jobs.NewHookJob(func() error {
		fmt.Println("hook executed")
		return nil
	})
	q.Enqueue(hookJob)

	q.Wait()

	if s, ok := q.Status(fileJob.ID()); ok {
		fmt.Printf("file job status: %+v\n", s)
	}
	if s, ok := q.Status(hookJob.ID()); ok {
		fmt.Printf("hook job status: %+v\n", s)
	}
}
