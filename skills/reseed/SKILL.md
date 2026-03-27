---
name: reseed
description: How to use reseed to manage agent skills in the current project. Use this skill whenever the user mentions reseed, agent skills, skill management, adding or syncing skills, browsing their skill library, or setting up skills for a project. Also use it when the user asks about what skills are available, wants to install skills from GitHub, or mentions .agents/skills/ or .claude/skills/.
---

# Reseed - Agent Skills Manager

Reseed manages a personal skill library and installs skills into projects. Skills are directories containing a `SKILL.md` file, following the [Agent Skills spec](https://agentskills.io). They live in `.agents/skills/` within each project and are read by 30+ agents (Claude Code, Cursor, Copilot, Codex, Gemini CLI, etc.).

The user manages their library. Your job is to handle everything on the project side - checking what's installed, adding what's needed, and keeping things in sync.

## Quick reference

```
reseed status                       # what's installed in this project
reseed list -l                      # what's in the user's library (with descriptions)
reseed add <skills-or-packs...>     # copy skills/packs from library into the project
reseed add --all                    # add every skill from the library
reseed remove <skills...>           # remove skills from the project
reseed sync                         # re-copy installed skills from library (get updates)
reseed install <source>             # fetch skills from GitHub into the library
```

## How it works

- The user has a **library** - a directory on their machine containing all their skills. It can include **packs** (named groups of related skills).
- `reseed add` copies skills from the library into the project's skills directory. These are real file copies, not symlinks, so they show up in git and every team member gets them.
- `reseed sync` re-copies skills that exist in both the project and the library, pulling in any updates. Skills not in the library are left alone.
- There's no project manifest. Sync matches by folder name.
- The default target directory is `.agents/skills/`, but different tools read from different locations. Claude Code uses `.claude/skills/`. Override with `--dir` on any command, or set it permanently with `reseed config dir`:

```bash
reseed --dir .claude/skills add commit      # per-command
reseed config dir .claude/skills            # persist for all commands
```

## Typical workflow

### 0. Check the target directory

The default is `.agents/skills/`, but some tools use a different location -- Claude Code reads from `.claude/skills/`. If you need a different directory, pass `--dir` on every command:

```bash
reseed --dir .claude/skills add commit
reseed --dir .claude/skills sync
```

If the user wants to avoid repeating `--dir`, suggest they persist it with `reseed config dir .claude/skills` -- but ask first, since that changes their global default.

### 1. Check what's already here

```bash
reseed status
```

If the skills directory doesn't exist yet, that's fine - reseed creates it automatically when you add skills.

### 2. See what's available

```bash
reseed list -l
```

This shows all skills and packs in the user's library with descriptions. Use this to understand what's available before suggesting what to add.

### 3. Add skills to the project

```bash
reseed add commit review python-base   # mix skills and packs in one command
reseed add --all                       # or just add everything
```

You can pass any number of skills and packs in a single `reseed add` command - reseed expands packs into their individual skills automatically. Pick what matches the project's stack and needs.

### 4. Keep skills up to date

```bash
reseed sync
```

Run this when the user wants to pull the latest versions of their skills into the project.

## Installing skills from GitHub

`install` and `add` do different things:
- `reseed install` fetches from GitHub into the user's **library**. It takes a GitHub reference (`user/repo`).
- `reseed add` copies from the library into the **project**. It takes local skill or pack names.

```bash
reseed install user/repo                    # all skills from the repo
reseed install user/repo/path/to/skills     # skills under a specific directory
reseed install user/repo@v2.0               # pin to a version tag
reseed install user/repo --pack mypack      # group them into a pack
```

After installing, use `reseed add` to bring them into the project.

## When reseed is not installed

If `reseed` is not found, point the user to the install command:

```bash
curl -fsSL https://raw.githubusercontent.com/nattergabriel/reseed/main/install.sh | sh
```

On Windows, they can download the binary from the [latest release](https://github.com/nattergabriel/reseed/releases/latest).

If the user hasn't initialized a library yet, they need to run:

```bash
reseed init ~/skills    # or any path they prefer
```

