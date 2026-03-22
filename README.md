<h3 align="center">reseed</h3>

<p align="center">
  A CLI tool for managing and distributing <a href="https://agentskills.io">agent skills</a> across projects.<br>
  Build a personal skill library and install skills into any project.
</p>

<p align="center">
  <a href="https://github.com/nattergabriel/reseed/actions/workflows/ci.yml"><img src="https://github.com/nattergabriel/reseed/actions/workflows/ci.yml/badge.svg" alt="CI" /></a>
  <a href="https://github.com/nattergabriel/reseed/blob/main/LICENSE"><img src="https://img.shields.io/github/license/nattergabriel/reseed" alt="License" /></a>
  <a href="https://github.com/nattergabriel/reseed/releases"><img src="https://img.shields.io/github/v/release/nattergabriel/reseed?include_prereleases" alt="Release" /></a>
</p>

---

Reseed maintains a personal skill library and lets you install skills into any project's `.agents/skills/` directory. Skills follow the [Agent Skills](https://agentskills.io) spec, compatible with 30+ agents including Claude Code, Cursor, Copilot, Codex, and Gemini CLI.

## Install

**macOS and Linux:**

```bash
curl -sSfL https://raw.githubusercontent.com/nattergabriel/reseed/main/install.sh | sh
```

**Windows:** download the binary from the [latest release](https://github.com/nattergabriel/reseed/releases/latest).

## Getting started

Your **library** is the central place where you manage all your skills. It can be any directory on your machine (and can itself be a git repo to version and share your collection). You create your own skills or pull in open source skills from GitHub. When open source skills get updated upstream, reseed keeps your library in sync so you don't have to track changes manually.

From your library, you install skills into any project. Each project gets its own copy in `.agents/skills/`, committed to git so your whole team benefits without needing reseed.

### 1. Create your library

Pick a directory to store your skills. This only needs to be done once.

```bash
reseed init ~/my-skills
```

### 2. Add skills to your library

You can write your own skills (any folder with a `SKILL.md` file) or pull in open source skills from GitHub. These are tracked in your library and can be updated automatically when new versions are published.

```bash
reseed install user/repo              # all skills from a repo
reseed install user/repo/skill-name   # a specific skill
reseed install user/repo@v1.0         # pinned to a tag
```

### 3. Use skills in a project

From inside a project, add skills (or packs of skills) from your library. This copies them into the project's `.agents/skills/` directory.

```bash
reseed add skill-name
reseed add my-pack
```

### 4. Keep things up to date

Fetch the latest versions of open source skills into your library, then push those updates into your projects.

```bash
reseed update                         # fetch latest versions from GitHub into library
reseed sync                           # re-copy library skills into current project
```

### Other commands

```bash
reseed library                        # list all skills in your library
reseed list                           # list skills in current project
reseed inspect pack-name              # show skills in a pack
reseed remove skill-name              # remove a skill from current project
```

## Contributing

Requires Go 1.24+ and [golangci-lint](https://golangci-lint.run/). Run `make setup` to enable pre-commit hooks.

## License

[MIT](LICENSE). Free to use, modify, and distribute.
