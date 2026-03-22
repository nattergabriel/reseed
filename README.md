<h3 align="center">reseed</h3>

<p align="center">
  A CLI tool for managing and distributing <a href="https://agentskills.io">agent skills</a> across projects
</p>

<p align="center">
  <a href="https://reseed.mintlify.app"><img src="https://img.shields.io/badge/docs-mintlify-blue" alt="Docs" /></a>
  <a href="https://github.com/nattergabriel/reseed/releases"><img src="https://img.shields.io/github/v/release/nattergabriel/reseed?include_prereleases" alt="Release" /></a>
  <a href="https://github.com/nattergabriel/reseed/actions/workflows/ci.yml"><img src="https://github.com/nattergabriel/reseed/actions/workflows/ci.yml/badge.svg" alt="CI" /></a>
  <a href="https://github.com/nattergabriel/reseed/blob/main/LICENSE"><img src="https://img.shields.io/github/license/nattergabriel/reseed" alt="License" /></a>
</p>

---

AI coding agents like Claude Code, Cursor, and Copilot can be customized with [agent skills](https://agentskills.io): markdown files that tell them how to work in your project. But managing skills across multiple projects quickly gets tedious.

Reseed gives you a single library to organize your skills. Write your own, pull in open source ones from GitHub and keep them synced, bundle related skills into packs, and install exactly what each project needs. Skills are copied into git so your whole team benefits without needing reseed.

## Install

**macOS and Linux:**

```bash
curl -fsSL https://raw.githubusercontent.com/nattergabriel/reseed/main/install.sh | sh
```

**Windows:** download the binary from the [latest release](https://github.com/nattergabriel/reseed/releases/latest) and add it to your PATH.

## Getting started

For full documentation, visit [reseed.mintlify.app](https://reseed.mintlify.app).

Your **library** is a directory where all your skills live. It can be any folder on your machine (and can itself be a git repo to version and share your collection). From there, you install skills into any project's `.agents/skills/` directory.

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
