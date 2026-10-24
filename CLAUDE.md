# cc-launch

Inline BubbleTea (v1) launcher that picks a directory and `exec`s `claude` in it.

- `main.go`: load/bootstrap config, run picker, launch claude (`launch_unix.go`: chdir + `syscall.Exec`; `launch_windows.go`: child process with stdio passthrough and exit code propagated, since Windows has no exec).
- `config.go`: `~/.config/cc-launch.yml` (preferred) or `.toml`; keys `default_prefix`, `directories`.
- `setup.go`: first-run prompt for `default_prefix`; writes `cc-launch.yml`.
- `picker.go`: main UI. Plain text filters `directories`; leading `/` browses subdirectories of `default_prefix` (nested `/a/b` and Tab-completion supported).
- `other…` is a trailing row (cursor == len(items)) opening a free-form path input. `cc-launch <path>` skips the picker (`runDirect`: upgrade, then launch).
- `upgrade.go`: `claude upgrade` started at launch in the background, hard 30s cap. If the user picks a directory before it finishes, launch waits for it.

Commands: `go build -o cc-launch .`, `go test ./...`.

CI: `.github/workflows/build.yml` runs gofmt/vet/test, cross-builds darwin/linux/windows (amd64/arm64; .zip for Windows) on one Linux runner, signs the darwin binaries in the build job (indygreg/apple-code-sign-action; secrets `APPSTORE_CERTIFICATES_FILE_BASE64` and `APPSTORE_CERTIFICATES_PASSWORD`; skipped on PRs), and on a `v*` tag publishes a GitHub release with the archives. Action versions are pinned via `actions.lock` (`gh actions-lock`).
