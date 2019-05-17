package config

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/pkg/errors"
)

func SetupLogging() error {
	var writers []io.Writer
	if config.FileLogs {
		file, err := setupLogFiles()
		if err != nil {
			return err
		}

		writers = append(writers, file)
	}

	if config.StdoutLogs {
		writers = append(writers, os.Stdout)
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(io.MultiWriter(writers...))
	return nil
}

func setupLogFiles() (*os.File, error) {
	logDirectory := fmt.Sprintf("/var/log/%s", AppName)
	if _, err := os.Stat(logDirectory); os.IsNotExist(err) {
		err := os.Mkdir(logDirectory, os.ModePerm)
		if err != nil {
			return nil, errors.Wrap(
				err,
				"couldn't create logging directory",
			)
		}
	}

	logFilePath := fmt.Sprintf("%s/default", logDirectory)
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, errors.Wrap(
			err,
			"couldn't open logfile",
		)
	}

	// log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	// var output io.Writer
	// switch {
	// case stdoutLogs || fileLogs:
	// 	output = io.MultiWriter(os.Stdout, file)
	// case fileLogs:
	// 	output = os.Stdout
	// case stdoutLogs:
	// 	output = file
	// default:
	// }
	// if printLogs {
	// 	log.SetOutput(io.MultiWriter(os.Stdout, file))
	// } else {
	// 	log.SetOutput(file)
	// }

	return file, nil
}
