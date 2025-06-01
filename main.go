package main

import (
	"flag"
	"log"

	"github.com/demarijm/inboxproxy/hooks"
)

func main() {
	sender := flag.String("sender", "", "email sender")
	subject := flag.String("subject", "", "email subject")
	fileType := flag.String("file", "", "file type")
	config := flag.String("config", "hooks.json", "path to hooks json")
	flag.Parse()

	hs, err := hooks.LoadHooks(*config)
	if err != nil {
		log.Fatal(err)
	}

	runner := hooks.NewHookRunner(hs)
	runner.Start()
	runner.Process(hooks.EmailEvent{
		Sender:   *sender,
		Subject:  *subject,
		FileType: *fileType,
	})
	runner.Stop()
}
