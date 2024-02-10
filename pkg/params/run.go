package params

import (
	"strings"
)

const ()

type Run struct {
}

func StringToBool(s string) bool {
	if strings.ToLower(s) == "true" ||
		strings.ToLower(s) == "yes" || s == "1" {
		return true
	}
	return false
}

func New() *Run {
	return &Run{}
}
