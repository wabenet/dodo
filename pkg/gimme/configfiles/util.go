package configfiles

import (
	"path/filepath"
	"strings"
)

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	index := 0
	for _, item := range values {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = true
		values[index] = item
		index++
	}
	return values[:index]
}

func isFSRoot(path string) bool {
	return strings.HasSuffix(filepath.Clean(path), string(filepath.Separator))
}
