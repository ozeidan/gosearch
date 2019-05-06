package main

import (
	"fmt"
	"strings"

	"github.com/karrick/godirwalk"
)

func start() {
	dirname := "/"
	err := godirwalk.Walk(dirname, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			// fmt.Printf("%s\n", osPathname)
			if de.IsDir() {
				dirCount++
			}
			if strings.HasPrefix(osPathname, "/proc") {
				return filterError
			}
			return nil
		},
		Unsorted: true, // (optional) set true for faster yet non-deterministic enumeration (see godoc)
		ErrorCallback: func(_ string, err error) godirwalk.ErrorAction {
			if err == filterError {
				return godirwalk.SkipNode
			}
			fmt.Println(err)
			return godirwalk.SkipNode
		},
	})
	if err != nil {
		fmt.Println(err)
	}
}
