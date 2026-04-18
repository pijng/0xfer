package fetch

import (
	"net/url"
	"strings"
)

func FilenameFromHeader(cd, remotePath string) string {
	filename := extractFilenameFromDisposition(cd)
	if filename != "" {
		return filename
	}

	return extractFilenameFromPath(remotePath)
}

func extractFilenameFromDisposition(cd string) string {
	idx := strings.Index(cd, "filename=")
	if idx == -1 {
		return ""
	}

	namePart := cd[idx+len("filename="):]
	if len(namePart) == 0 {
		return ""
	}

	if namePart[0] == '"' {
		namePart = namePart[1:]
		idx = strings.Index(namePart, "\"")
		if idx == -1 {
			return ""
		}

		return namePart[:idx]
	}

	idx = strings.Index(namePart, ";")
	if idx != -1 {
		return namePart[:idx]
	}

	return strings.TrimSpace(namePart)
}

func extractFilenameFromPath(path string) string {
	if path == "" {
		return ""
	}

	u, err := url.Parse(path)
	if err != nil {
		return ""
	}

	filename := u.Path
	idx := strings.LastIndex(filename, "/")
	if idx == -1 || idx+1 >= len(filename) {
		return ""
	}

	return filename[idx+1:]
}