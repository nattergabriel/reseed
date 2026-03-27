# Reseed Copy

<!-- Central source for all project descriptions. Update here first, then propagate to:
README.md, website, docs, GitHub About, ... -->

## Tagline

A CLI tool for managing and distributing agent skills across projects.

## Description

Reseed manages your [agent skills](https://agentskills.io) across projects. Keep all your skills in one central library, pull in open source ones from GitHub, and install exactly what each project needs. Instead of global skills that bloat every project, skills live in each project. That way also every teammate has access to them. Your library can be a git repo to version and share your collection.

## How it works

1. You create a **library**, a directory where all your skills live.
2. You write your own skills or install open source ones from GitHub.
3. You browse your library and add skills to each project, either through the interactive TUI or via CLI commands.
4. When upstream skills change, you re-run `reseed install` to pull the latest, then `reseed sync` to update your projects.

## Key ideas

- **Skills are copies, not symlinks.** They live in each project so every teammate has access.
- **Your library is yours.** It can be a git repo, a Dropbox folder, or just a directory on your machine.
- **Not every project needs every skill.** You pick what to install per project.
- **Open source skills from GitHub can be updated** by re-running `reseed install`, then `reseed sync` to push changes into projects.

## Why reseed

AI coding agents like Claude Code, Cursor, and Copilot get better when you give them skills -- custom prompts, workflows, and tool configurations tailored to your codebase.

But skills live in the wrong places. Store them in your user config and teammates never see them, plus every project gets the same set whether it needs them or not. Copy them into projects by hand and they go stale the moment you improve one.

Reseed gives you a **central skill library** that you own. Collect skills there, organize them into packs, then pick exactly what each project needs. One command copies them in. Another keeps them in sync when the library changes. Your teammates see the same skills you do, and every project gets only what it actually uses.
