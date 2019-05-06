package main

import (
	"errors"
	"fmt"
	"time"
)

var dirCount = 0
var filterError error = errors.New("directory filtered")

func main() {
	start := time.Now()
	elapsed := time.Since(start)
	fmt.Println("elapsed time:", elapsed)
	fmt.Println("counted", dirCount, "directories")
	// err := fanotifyInit()
	// if err != nil {
	// 	fmt.Println(err)
	// }
}
