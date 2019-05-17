package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/ozeidan/gosearch/internal/config"
	"github.com/ozeidan/gosearch/internal/database"
	"github.com/ozeidan/gosearch/internal/fanotify"
	"github.com/ozeidan/gosearch/internal/request"
)

func main() {
	err := config.ParseConfig()
	if err != nil {
		log.Println("failed to initialize configuration", err)
	}

	err = config.SetupLogging()
	if err != nil {
		log.Println(err)
		return
	}

	fileChangeChan := make(chan fanotify.FileChange, 100)
	requestChan := make(chan request.Request)
	go fanotify.Listen(fileChangeChan)
	go database.Start(fileChangeChan, requestChan)
	go request.ListenAndServe(requestChan)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for range c {
		break
	}
}
