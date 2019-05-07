package config

import "strings"

var filteredPathPrefixes []string = []string{"/proc", "/home/omar/.cache", "/var/cache"}

func IsPathFiltered(path string) bool {
	for _, filterString := range filteredPathPrefixes {
		if strings.HasPrefix(path, filterString) {
			return true
		}
	}

	return false
}
