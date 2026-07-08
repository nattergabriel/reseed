# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

reseed is an open source Go CLI (Cobra) for managing agent skills across projects. Users keep skills in a central library directory and copy/sync them into individual projects' skills directories (`.agents/skills/` by default). A "skill" is any directory containing a `SKILL.md` marker file, per the [Agent Skills spec](https://agentskills.io).

## Commands

```bash
make build    # go build ./...
make test     # go test ./...
make lint     # go vet ./... && golangci-lint run
make setup    # enable pre-commit hooks (build + vet + golangci-lint)
```

Run a single test: `go test ./internal/reseed -run TestName`

Requires Go 1.24+. CI runs build, test, and golangci-lint on every PR.

## Architecture

Two sides connected by copy operations:

- **Library** (source): a user-chosen directory with a `skills/` subdir. Its path is stored in a global YAML config at `os.UserConfigDir()/reseed` (`internal/config`).
- **Project** (destination): the current working directory's skills dir, default `.agents/skills`, overridable by the `--dir` persistent flag or the `dir` field in global config (`internal/project.SkillsPath`).

Data flow: `reseed install` downloads skills from GitHub into the library; `reseed add` and `reseed sync` copy from library to project. Copies are full replacements (`RemoveAll` then recursive copy, see `skill.Copy`).

Packages under `internal/`:

- `reseed`: the whole domain in one package. `Skill{Name, Group, Path}` is the currency on the library side (never bare names there); the project side deals in bare installed names, since project skills have no group and their path is derivable. `Library` is walked once at `OpenLibrary` into `Library.Skills`; `Resolve` expands a skill name or folder path into skills and errors on ambiguous names (basenames must be unique across the library). `Project` resolves its skills dir once at `OpenProject` (flag > config > default; absolute paths respected) and has `Add`/`Remove`/`Sync`/`Installed`. `Config` load/save lives here too; a missing config file is an empty config, a corrupt one is an error.
- `github`: installs from GitHub. `ParseRef` accepts `user/repo[/path]` only (no URLs, no versions). Downloads the default-branch tarball via the GitHub API; uses `GITHUB_TOKEN` if set.

`cmd/` has one file per subcommand plus `tui.go`. Commands are grouped into "library" and "project" Cobra groups (`cmd/root.go`); the `--dir` persistent flag is `cmd.flagDir`, passed to `OpenProject`. Running `reseed` with no args launches an interactive Bubble Tea TUI (`cmd/tui.go`) that shares the same `Library`/`Project` methods as the CLI commands.

`skills/reseed/SKILL.md` is the agent skill shipped with the repo that teaches agents how to use reseed; keep it in sync when command behavior changes.

## Key design decisions

- Skills are **copies**, not symlinks. They land in the project's repo, so teammates get them via git without needing reseed.
- No project manifest. The skills directory is the source of truth; `sync` matches by folder name against the library.
- `add` is local and fast (no network); only `install` fetches from GitHub.

## Documentation

- Keep the README concise: only what users need to get started, not every flag or feature. Full details belong in the docs (reseed-docs, published at reseed.mintlify.app).

## Conventions

- Keep it simple. No over-engineering, no unnecessary abstractions.
- No backwards compatibility. When behavior changes or a feature is removed, make a clean break: no shims, aliases, deprecation paths, or code that only exists to handle the old behavior.
- No third-party dependencies unless strictly necessary (currently Cobra, yaml.v3, and the Charm TUI stack: bubbletea, lipgloss, huh).
- Wrap errors with `fmt.Errorf` and context: `fmt.Errorf("doing thing: %w", err)`.
- Conventional commit messages (`feat:`, `fix:`, `docs:`, ...).
- All lint checks must pass before committing (`make lint`); the pre-commit hook (enabled via `make setup`) runs build, vet, and golangci-lint automatically.
- Releases are tagged and built with GoReleaser (`.goreleaser.yaml`); the version is injected via ldflags into `cmd.version`.
