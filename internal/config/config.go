package config

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

type serverConfig struct {
	PrefixFilters     []string `json:"prefix_filters"`
	SubstringFilters  []string `json:"substring_filters"`
	RegexFilters      []string `json:"regex_filters"`
	IgnoreHiddenFiles bool     `json:"ignore_hidden_files"`
}

const configPath = "/etc/goSearch/config" // TODO: XDG_CONFIG_DIRS?
var config = serverConfig{[]string{}, []string{}, []string{}, false}

var regexFilters []*regexp.Regexp

// ParseConfig initializes the configuration of the program
// by reading and parsing the config file
func ParseConfig() error {
	file, err := os.Open(configPath)

	if os.IsNotExist(err) {
		err = createConfigStub()
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}

	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return err
	}

	parseFilters()
	return nil
}

func createConfigStub() error {
	err := os.Mkdir("/etc/goSearch", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return errors.Wrap(err, "can't create config directory")
	}
	f, err := os.Create(configPath)
	if err != nil {
		return errors.Wrap(err, "can't create config file")
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "    ")

	err = enc.Encode(&config)
	if err != nil {
		return err
	}

	return nil
}

func parseFilters() {
	for _, filterString := range config.RegexFilters {
		r, err := regexp.Compile(filterString)
		if err != nil {
			log.Println("failed to parse regex filter:", err)
		}
		regexFilters = append(regexFilters, r)
	}
}

// IsPathFiltered determines returns whether the given path is filtered
// by the user's configuration
func IsPathFiltered(path string) bool {
	for _, prefixFilter := range config.PrefixFilters {
		if strings.HasPrefix(path, prefixFilter) {
			return true
		}
	}

	for _, substringFilter := range config.SubstringFilters {
		if strings.Contains(path, substringFilter) {
			return true
		}
	}

	for _, r := range regexFilters {
		if r.MatchString(path) {
			return true
		}
	}

	if config.IgnoreHiddenFiles {
		parts := strings.Split(path, "/")
		for _, part := range parts {
			if len(part) > 0 && part[0] == '.' {
				return true
			}
		}
	}

	return false
}
