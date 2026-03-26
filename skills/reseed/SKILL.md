---
name: reseed
description: How to use the reseed CLI to manage agent skills. Use when the user wants to add, remove, install, or sync skills, or mentions reseed, skill packs, or .agents/skills directories.
---

# Reseed CLI

Reseed manages agent skills across projects. It maintains a personal skill library and copies skills into a project's skills directory (`.agents/skills/` by default). Skills are regular directories identified by a `SKILL.md` marker file.

## Key concepts

- **Library**: A directory on the user's machine where all their skills live. Created once with `reseed init`.
- **Skills**: Directories containing a `SKILL.md` file. They get copied (not symlinked) into projects.
- **Packs**: Named groups of skills within the library. A pack is just a subdirectory containing multiple skills.
- **Project skills directory**: Where skills live in a project. Defaults to `.agents/skills/`, configurable via `reseed config dir`.

## Commands

### Library management

```bash
# Initialize a library (one-time setup)
reseed init ~/skills

# Install skills from GitHub into your library
reseed install user/repo
reseed install user/repo/path/to/skills      # specific subdirectory
reseed install user/repo@v2.0                 # pinned to a tag
reseed install user/repo --pack my-pack       # group into a pack

# List everything in the library
reseed list
```

### Project management

Run these from inside a project directory.

```bash
# Add skills or packs to the current project
reseed add commit-message
reseed add my-pack                  # adds all skills in the pack
reseed add skill-one skill-two      # multiple at once
reseed add --all                    # everything in the library

# Remove skills from the project
reseed remove commit-message
reseed remove skill-one skill-two   # multiple at once

# See what's installed in the current project
reseed status

# Re-copy library skills into the project (update to latest)
reseed sync
```

### Configuration

```bash
# Get the current skills directory
reseed config dir

# Change the default skills directory
reseed config dir .claude/skills
```

The `--dir` flag on any command overrides the skills directory for that invocation:

```bash
reseed add commit-message --dir .claude/skills
```

## Output format

Commands print one line per skill with a prefix indicating the action:

- `+ skill-name` for additions (`add`, `install`)
- `- skill-name` for removals (`remove`)
- `~ skill-name` for updates (`sync`)

After the list, a summary line follows: `Added 3 skills.`

`reseed list` prints skill names one per line, with packs shown as sections:

```
commit-message
writing-style

my-pack:
  code-review
  generate-tests
```

`reseed status` prints installed skill names one per line, no formatting.

## Error handling

Common errors and what to do:

- `library not initialized: run 'reseed init <path>' first` - The user needs to create a library before using any other command.
- `skill "X" not found in library` - The skill name doesn't match anything in the library. Run `reseed list` to see available names.
- `skill "X" not installed` - Tried to remove a skill that isn't in the project. Run `reseed status` to check.

## Typical workflows

**Setting up a new project with skills:**
```bash
reseed status                    # check what's already installed
reseed list                      # see what's available
reseed add commit-message        # add what you need
```

**Updating skills after library changes:**
```bash
reseed sync
```

**Installing skills from a public repo:**
```bash
reseed install anthropics/skills/skills --pack anthropic
reseed add anthropic
```
