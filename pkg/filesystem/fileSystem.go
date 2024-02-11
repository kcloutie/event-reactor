package filesystem

import "strings"

func NormalizeFilePath(path string) string {
	if strings.Contains(path, ":") {
		path = strings.Split(path, ":")[1]
	}
	return strings.ReplaceAll(path, "\\", "/")
}
