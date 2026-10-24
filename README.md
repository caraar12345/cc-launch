# cc-launch

Inline terminal UI to launch [Claude Code](https://claude.com/claude-code) in a directory.

```
go install github.com/caraar12345/cc-launch@latest
cc-launch
```

- Type to filter your saved directories.
- Type `/` to browse under your `default_prefix`: `/cc-launch` + enter opens Claude Code in `<default_prefix>/cc-launch`. Tab completes, so `/foo/` descends further.
- On start it runs `claude upgrade` (max 30s) in the background; launching waits for it.

## Config

`~/.config/cc-launch.yml` (preferred) or `~/.config/cc-launch.toml`. Created on first run.

```yaml
default_prefix: ~/src/github.com/caraar12345
directories:
  - ~/src/github.com/caraar12345/cc-launch
  - ~/src/other
```
