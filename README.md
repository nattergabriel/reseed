<p align="center">
  <img src=".github/logo.svg" alt="reseed" height="40" />
</p>

<p align="center">
  A CLI tool for managing and distributing <a href="https://agentskills.io">agent skills</a> across projects
</p>

<p align="center">
  <a href="https://reseed.mintlify.app"><img src="https://img.shields.io/badge/docs-mintlify-59617F" alt="Docs" /></a>
  <a href="https://github.com/nattergabriel/reseed/releases"><img src="https://img.shields.io/github/v/release/nattergabriel/reseed?include_prereleases&color=59617F" alt="Release" /></a>
  <a href="https://github.com/nattergabriel/reseed/blob/main/LICENSE"><img src="https://img.shields.io/github/license/nattergabriel/reseed?color=59617F" alt="License" /></a>
  <a href="https://github.com/nattergabriel/reseed/actions/workflows/ci.yml"><img src="https://github.com/nattergabriel/reseed/actions/workflows/ci.yml/badge.svg" alt="CI" /></a>
</p>

---

Reseed manages your [agent skills](https://agentskills.io) across projects. Keep all your skills in one central library, pull in open source ones from GitHub, and install exactly what each project needs. Instead of global skills that bloat every project, skills live in each project so every teammate has access. Your library can be a git repo to version and share your collection.

## Install

**macOS and Linux:**

```bash
curl -fsSL https://raw.githubusercontent.com/nattergabriel/reseed/main/install.sh | sh
```

**Windows:** download the binary from the [latest release](https://github.com/nattergabriel/reseed/releases/latest) and add it to your PATH.

## Getting started

For full documentation, visit [reseed.mintlify.app](https://reseed.mintlify.app).

Your **library** is a directory where all your skills live. It can be any folder on your machine (and can itself be a git repo to version and share your collection). From there, you install skills into any project's `.agents/skills/` directory. You can change the default directory with `reseed config dir .claude/skills`.

### 1. Create your library

This only needs to be done once.

```bash
reseed init ~/my-skills  # specific directory
reseed init              # or use the current directory
```

### 2. Add skills to your library

Write your own skills (any folder with a `SKILL.md` file) or pull in open source ones from GitHub. These are tracked in your library and can be updated when new versions are published.

```bash
reseed install user/repo             # all skills from a repo
reseed install user/repo/skill-name  # a specific skill
reseed install user/repo@v1.0        # pinned to a tag
```

### 3. Organize skills into packs

Bundle related skills together for easy installation. The `pack` command opens an interactive selector.

```bash
reseed pack frontend  # create or edit the "frontend" pack
```

### 4. Use skills in a project

From inside a project, add skills or packs from your library. Run `reseed add` without arguments for an interactive selector.

<p align="center">
  <img src=".github/screenshot.png" alt="reseed add interactive selector" width="500" />
</p>

```bash
reseed add                  # interactive selection
reseed add skill-a skill-b  # add by name
reseed add my-pack skill-a  # mix packs and skills
reseed add --all            # add everything in your library
```

### 5. Keep things up to date

Fetch the latest versions of open source skills into your library, then push those updates into your projects.

```bash
reseed fetch  # fetch latest versions from GitHub into library
reseed sync   # re-copy library skills into project
```

### Other commands

```bash
reseed library            # list all skills in your library
reseed config dir <path>  # change default skills directory
```

## Contributing

Requires Go 1.24+ and [golangci-lint](https://golangci-lint.run/). Run `make setup` to enable pre-commit hooks.

## License

[MIT](LICENSE). Free to use, modify, and distribute.
