package config

import "strings"

var filteredPathPrefixes = []string{"/proc", "/home/omar/.cache", "/var/cache"}

// IsPathFiltered determiens returns whether the given path is filtered
// by the user's configuration
func IsPathFiltered(path string) bool {
	for _, filterString := range filteredPathPrefixes {
		if strings.HasPrefix(path, filterString) {
			return true
		}
	}

	return false
}
