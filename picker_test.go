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
