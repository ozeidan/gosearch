package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/karrick/godirwalk"
	"github.com/ozeidan/gosearch/request"
)

func main() {
	fileChangeChan := make(chan fileChange, 100)
	requestChan := make(chan request.Request)
	go fanotifyInit(fileChangeChan)
	go start(fileChangeChan, requestChan)
	go request.ListenAndServe(requestChan)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for _ = range c {
		break
	}
}
