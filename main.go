// cc-launch is an inline TUI for launching Claude Code in a chosen directory.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "cc-launch:", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "-h" || arg == "--help" {
			fmt.Println("usage: cc-launch [path]")
			return nil
		}
		dir, err := filepath.Abs(expandHome(arg))
		if err != nil || !isDir(dir) {
			return fmt.Errorf("%s is not a directory", arg)
		}
		if dir, err = runDirect(dir); err != nil || dir == "" {
			return err
		}
		return launch(dir)
	}

	path, exists, err := configPath()
	if err != nil {
		return err
	}

	var cfg *Config
	if exists {
		if cfg, err = loadConfig(path); err != nil {
			return err
		}
	} else {
		prefix, err := runSetup()
		if err != nil {
			return err
		}
		if prefix == "" {
			return nil // cancelled
		}
		cfg = &Config{DefaultPrefix: prefix, Directories: []string{}}
		if err := saveConfig(path, cfg); err != nil {
			return err
		}
		fmt.Printf("Saved %s\n", tildify(path))
	}
	if cfg.DefaultPrefix == "" {
		return fmt.Errorf("%s: default_prefix is not set", tildify(path))
	}

	dir, err := runPicker(cfg)
	if err != nil {
		return err
	}
	if dir == "" {
		return nil
	}
	return launch(dir)
}

// launch replaces this process with claude running in dir.
func launch(dir string) error {
	bin, err := exec.LookPath("claude")
	if err != nil {
		return errors.New("claude not found in PATH")
	}
	if err := os.Chdir(dir); err != nil {
		return err
	}
	return syscall.Exec(bin, []string{"claude"}, os.Environ())
}
