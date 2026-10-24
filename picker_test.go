package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFilterDirs(t *testing.T) {
	dirs := []string{"/a/foo-api", "/a/bar", "/b/api-foo", "/c/other"}
	got := filterDirs(dirs, "api")
	want := []string{"/b/api-foo", "/a/foo-api"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if got := filterDirs(dirs, "zzz"); len(got) != 0 {
		t.Fatalf("expected no matches, got %v", got)
	}
	if got := filterDirs(dirs, ""); len(got) != len(dirs) {
		t.Fatalf("empty query should keep all, got %v", got)
	}
}

func TestBrowseDirs(t *testing.T) {
	root := t.TempDir()
	for _, d := range []string{"cc-launch", "cc-launch-old", "other", ".hidden", "work/api"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	os.WriteFile(filepath.Join(root, "file"), nil, 0o644)

	got := browseDirs(root, "cc-launch")
	if len(got) != 2 || got[0] != filepath.Join(root, "cc-launch") {
		t.Fatalf("exact match should rank first, got %v", got)
	}
	if got := browseDirs(root, ""); len(got) != 4 { // no hidden, no files
		t.Fatalf("got %v", got)
	}
	if got := browseDirs(root, "work/"); len(got) != 1 || got[0] != filepath.Join(root, "work", "api") {
		t.Fatalf("nested browse failed: %v", got)
	}
}

func TestParseUpgrade(t *testing.T) {
	b, a := parseUpgrade("Current version: 2.1.310\nChecking for updates to latest version...\nClaude Code is up to date (2.1.310)\n")
	if b != "2.1.310" || a != "2.1.310" {
		t.Fatalf("got %q %q", b, a)
	}
	b, a = parseUpgrade("Current version: 2.1.309\nChecking...\nSuccessfully updated from 2.1.309 to version 2.1.310\n")
	if b != "2.1.309" || a != "2.1.310" {
		t.Fatalf("got %q %q", b, a)
	}
}

func TestParseUpgradeRealOutput(t *testing.T) {
	// Wording taken from anthropics/claude-code#40987 and #73830.
	b, a := parseUpgrade("Current version: 2.1.86\nChecking for updates to latest version...\nSuccessfully updated from 2.1.86 to version 2.1.87\n")
	if b != "2.1.86" || a != "2.1.87" {
		t.Fatalf("got %q %q", b, a)
	}
	b, a = parseUpgrade("Current version: 2.1.200\nChecking for updates to latest version...\nUpdating to 2.1.199...\nClaude Code is up to date (2.1.200\n")
	if b != "2.1.200" || a != "2.1.200" {
		t.Fatalf("got %q %q", b, a)
	}
}
