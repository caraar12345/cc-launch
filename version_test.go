package main

import (
	"regexp"
	"strings"
	"testing"
)

func TestVersionFile(t *testing.T) {
	if v := strings.TrimSpace(version); !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(v) {
		t.Fatalf("VERSION %q is not MAJOR.MINOR.PATCH", v)
	}
}
