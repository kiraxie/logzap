package logzap

import (
	"regexp"
)

func FilterLogPattern(msg string) string {
	return reFilterToken.ReplaceAllString(msg, "${1}[MASKED]")
}

var reFilterToken = regexp.MustCompile(`([&?]token=)[0-9A-Za-z_-]+`)
