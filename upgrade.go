package main

import (
	"context"
	"errors"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const upgradeTimeout = 30 * time.Second

type upgradeDoneMsg struct {
	err      error
	timedOut bool
}

// upgradeCmd runs `claude upgrade` with a hard cap of upgradeTimeout.
func upgradeCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), upgradeTimeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, "claude", "upgrade")
		cmd.WaitDelay = 2 * time.Second // don't hang on child-held pipes
		err := cmd.Run()
		return upgradeDoneMsg{err: err, timedOut: errors.Is(ctx.Err(), context.DeadlineExceeded)}
	}
}
