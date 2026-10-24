# cc-launch

Inline BubbleTea (v1) launcher that picks a directory and `exec`s `claude` in it.

- `main.go`: load/bootstrap config, run picker, launch claude (`launch_unix.go`: chdir + `syscall.Exec`; `launch_windows.go`: child process with stdio passthrough and exit code propagated, since Windows has no exec).
- `config.go`: `~/.config/cc-launch.yml` (preferred) or `.toml`; keys `default_prefix`, `directories`.
- `setup.go`: first-run prompt for `default_prefix`; writes `cc-launch.yml`.
- `picker.go`: main UI. Plain text filters `directories`; leading `/` browses subdirectories of `default_prefix` (nested `/a/b` and Tab-completion supported).
- `other…` is a trailing row (cursor == len(items)) opening a free-form path input. `cc-launch <path>` skips the picker (`runDirect`: upgrade, then launch).
- `upgrade.go`: `claude upgrade` started at launch in the background, hard 30s cap. If the user picks a directory before it finishes, launch waits for it.

Commands: `go build -o cc-launch .`, `go test ./...`.

Versioning/release: `VERSION` (plain `MAJOR.MINOR.PATCH`) is the single source of truth; it is `go:embed`ded for `cc-launch --version`. Bump it and push to main to release: CI sees no `vX.Y.Z` release exists, builds, signs and notarizes the darwin binaries, tags the commit and publishes a release with auto-generated notes since the previous release. There is no tag trigger.

CI: `.github/workflows/build.yml` jobs: `version` (release or not), `test` (gofmt/vet/test), `build` (cross-builds darwin/linux/windows, amd64/arm64, on one Linux runner; .zip for Windows; darwin entries run in the `production` environment and are signed with indygreg/apple-code-sign-action, skipped on PRs; notarized on release builds only via `APP_STORE_CONNECT_KEY`, secrets `APPSTORE_CERTIFICATES_FILE_BASE64` + `APPSTORE_CERTIFICATES_PASSWORD`), `release`. Action versions are pinned via `actions.lock` (`gh actions-lock`).
