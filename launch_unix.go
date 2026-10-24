//go:build !windows

package main

import (
	"os"
	"syscall"
)

// execClaude replaces this process with claude running in dir.
func execClaude(bin, dir string) error {
	if err := os.Chdir(dir); err != nil {
		return err
	}
	return syscall.Exec(bin, []string{"claude"}, os.Environ())
}
