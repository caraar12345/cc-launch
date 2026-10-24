//go:build windows

package main

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
)

// execClaude runs claude as a child (Windows has no exec) and exits with its code.
func execClaude(bin, dir string) error {
	cmd := exec.Command(bin)
	cmd.Dir = dir
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	signal.Ignore(os.Interrupt) // Ctrl+C is for claude, not us
	err := cmd.Run()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		os.Exit(ee.ExitCode())
	}
	if err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
