package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/ozeidan/gosearch/internal/database"
	"github.com/ozeidan/gosearch/internal/fanotify"
	"github.com/ozeidan/gosearch/internal/request"
	"github.com/pkg/errors"
)

const appName = "goSearch"

func main() {
	err := setupLogFiles()

	if err != nil {
		fmt.Println(err)
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

func setupLogFiles() error {
	logDirectory := fmt.Sprintf("/var/log/%s", appName)
	if _, err := os.Stat(logDirectory); os.IsNotExist(err) {
		err := os.Mkdir(logDirectory, os.ModePerm)
		if err != nil {
			return errors.Wrap(
				err,
				"couldn't create logging directory",
			)
		}
	}

	logFilePath := fmt.Sprintf("%s/default", logDirectory)
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(
			err,
			"couldn't open logfile",
		)
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(file)

	return nil
}
