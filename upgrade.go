package main

import (
	"context"
	"errors"
	"os/exec"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const upgradeTimeout = 30 * time.Second

type upgradeDoneMsg struct {
	err           error
	timedOut      bool
	before, after string // claude versions around the upgrade, if known
}

var semver = regexp.MustCompile(`\d+\.\d+\.\d+\S*`)

// parseUpgrade reads `claude upgrade` output: the "Current version: X" line is
// the before version and the last version mentioned is the after version.
func parseUpgrade(out string) (before, after string) {
	for _, line := range strings.Split(out, "\n") {
		v := semver.FindAllString(line, -1)
		if len(v) == 0 {
			continue
		}
		if before == "" && strings.HasPrefix(strings.TrimSpace(line), "Current version") {
			before = v[0]
		}
		after = v[len(v)-1]
	}
	after = strings.TrimRight(after, ").,")
	return
}

// upgradeCmd runs `claude upgrade` with a hard cap of upgradeTimeout.
func upgradeCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), upgradeTimeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, "claude", "upgrade")
		cmd.WaitDelay = 2 * time.Second // don't hang on child-held pipes
		out, err := cmd.CombinedOutput()
		msg := upgradeDoneMsg{err: err, timedOut: errors.Is(ctx.Err(), context.DeadlineExceeded)}
		if err == nil {
			msg.before, msg.after = parseUpgrade(string(out))
		}
		return msg
	}
}
