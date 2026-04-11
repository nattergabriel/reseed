---
name: reseed
description: How to use reseed to manage agent skills in a project. Use when the user mentions reseed, agent skills, SKILL.md, .agents/skills/, or .claude/skills/, or wants to install, add, remove, sync, or list skills.
---

# reseed: Agent Skills Manager

reseed manages a personal skill library and installs skills into projects. Skills are directories containing a `SKILL.md` file, following the [Agent Skills spec](https://agentskills.io). They live in each project's skills directory (`.agents/skills/` by default, or `.claude/skills/` for Claude Code) and are read by 30+ agents (Claude Code, Cursor, Copilot, Codex, Gemini CLI, etc.).

Your job is mainly on the project side: checking what's installed, adding what's needed, and keeping things in sync. You can also run `reseed install` when the user wants to pull skills from GitHub into their library, but leave the library's structure and contents to them otherwise.

Note: bare `reseed` (no args) opens an interactive TUI. That is for the user to run in their own terminal, not for you. If they want to browse their library visually, tell them to run `reseed` themselves. For everything else, use the CLI commands below.

## Quick reference

```
reseed                              # interactive TUI (for the user to run, not you)
reseed init <path>                  # create a new library (first-time setup)
reseed status                       # what's installed in this project
reseed list -l                      # what's in the user's library (with descriptions)
reseed add <skills-or-packs...>     # copy skills/packs from library into the project
reseed add --all                    # add every skill from the library
reseed remove <skills...>           # remove skills from the project
reseed sync                         # re-copy installed skills from library (get updates)
reseed install <user/repo or URL>   # fetch skills from GitHub into the library
```

## How it works

- The user has a **library**: a directory on their machine containing all their skills. It can include **packs** (named groups of related skills).
- `reseed add` copies skills from the library into the project's skills directory. These are real file copies, not symlinks, so they show up in git and every team member gets them.
- `reseed sync` re-copies skills that exist in both the project and the library, pulling in any updates. Skills not in the library are left alone, including orphans: if a skill was removed from the library after it was added to the project, `sync` will not delete the stale copy. Use `reseed remove` for that.
- There's no project manifest. Sync matches by folder name.

## Typical workflow

### 0. Confirm the target directory

Before running project-side commands, check what directory reseed is targeting:

```bash
reseed config dir
```

If this prints `.agents/skills (default)` but the user is on Claude Code (which reads from `.claude/skills/`), suggest running `reseed config dir .claude/skills` once to switch the global default. Ask before running it since it changes user-wide config. If they prefer not to change the default, pass `--dir .claude/skills` on each command instead:

```bash
reseed --dir .claude/skills add commit
```

If `reseed config dir` already prints the right path, no override is needed. Run the commands below as-is.

### 1. Check what's already here

```bash
reseed status
```

If the skills directory doesn't exist yet, that's fine. reseed creates it automatically when you add skills.

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

You can pass any number of skills and packs in a single `reseed add` command, reseed expands packs into their individual skills automatically. Pick what matches the project's stack and needs.

### 4. Keep skills up to date

```bash
reseed sync
```

Run this when the user wants to pull the latest versions of their skills into the project.

## Installing skills from GitHub

`install` and `add` do different things:
- `reseed install` fetches from GitHub into the user's **library**. It takes a GitHub reference (short form or full URL).
- `reseed add` copies from the library into the **project**. It takes local skill or pack names.

```bash
reseed install user/repo                                          # all skills from the repo
reseed install user/repo/path/to/skills                           # skills under a specific directory
reseed install https://github.com/user/repo/tree/main/src/skills  # full GitHub URL also works
reseed install user/repo@v2.0                                     # pin to a version tag
reseed install user/repo --pack mypack                            # group them into a pack
```

After installing, use `reseed add` to bring them into the project.

GitHub's public API allows only ~60 unauthenticated requests per hour, so running several installs in quick succession can hit the rate limit. If `reseed install` fails with a rate-limit error, suggest setting `GITHUB_TOKEN` in the environment to authenticate.

## When reseed is not installed

If `reseed` is not found, point the user to the install command:

```bash
curl -fsSL https://raw.githubusercontent.com/nattergabriel/reseed/main/install.sh | sh
```

On Windows, they can download the binary from the [latest release](https://github.com/nattergabriel/reseed/releases/latest).

## When there's no library yet

If `reseed` is installed but commands fail with "library not initialized", the user needs to create a library once:

```bash
reseed init ~/skills    # or any path they prefer
```

After that, the CLI commands start working against the new library.

