package main

import (
	"os"
	"os/signal"

	"github.com/ozeidan/gosearch/internal/database"
	"github.com/ozeidan/gosearch/internal/fanotify"
	"github.com/ozeidan/gosearch/internal/request"
)

func main() {
	fileChangeChan := make(chan fanotify.FileChange, 100)
	requestChan := make(chan request.Request)
	go fanotify.FanotifyInit(fileChangeChan)
	go database.Start(fileChangeChan, requestChan)
	go request.ListenAndServe(requestChan)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for _ = range c {
		break
	}
}
