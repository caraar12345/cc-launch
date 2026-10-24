# cc-launch

Inline BubbleTea (v1) launcher that picks a directory and `exec`s `claude` in it.

- `main.go`: load/bootstrap config, run picker, `chdir` + `syscall.Exec` claude.
- `config.go`: `~/.config/cc-launch.yml` (preferred) or `.toml`; keys `default_prefix`, `directories`.
- `setup.go`: first-run prompt for `default_prefix`; writes `cc-launch.yml`.
- `picker.go`: main UI. Plain text filters `directories`; leading `/` browses subdirectories of `default_prefix` (nested `/a/b` and Tab-completion supported).
- `upgrade.go`: `claude upgrade` started at launch in the background, hard 30s cap. If the user picks a directory before it finishes, launch waits for it.

Commands: `go build -o cc-launch .`, `go test ./...`.
