# md-notes — Normalized PRD

## 1. Product name

md-notes

## 2. One-line description

A small CLI for terminal-resident developers that creates, tags, and finds plain-`.md` notes on disk so they can take scratch notes and recover them in seconds without an app lock-in.

## 3. Problem

Scratch notes pile up across random files and folders, and recovering an old one by half-remembered phrase takes minutes instead of seconds. App-based tools (Obsidian, Notion) lock notes into a proprietary store; the user wants plain `.md` files they own and can edit or sync with anything.

## 4. Target users

- **Primary:** the author — a terminal-resident developer who takes frequent scratch notes during the day and loses them in folders.
- **Secondary:** none. The tool is not designed for a broader audience; adoption by others is incidental.

## 5. Goal

Ship a small CLI that creates, tags, and finds markdown notes stored as plain `.md` files in a configurable folder. Search must feel instant on thousands of notes. Distribution is a single binary or `pipx`-installable script — no daemon, no Docker, no web UI.

## 6. User stories / scenarios

- As a user, I run `notes new "some title"` so that a new dated markdown file with frontmatter opens in `$EDITOR`.
- As a user, I run `notes find <query>` so that I can recover a note by half-remembered phrase in seconds, searching titles and bodies.
- As a user, I run `notes find --tag work` (or `notes find #work`) so that I can scope results to a category.
- As a user, I run `notes ls` so that I can see what I touched recently, most-recent first.
- As a user, I configure a notes directory once (default `~/notes/`) so that all commands operate on the same store.

## 7. Core capabilities

- Create a new note from a title, with date and empty-tags frontmatter, opened in `$EDITOR`.
- Full-text search across titles and bodies, fast at thousands-of-notes scale.
- Tag-based filtering of search results.
- List notes by modification recency.
- Configurable notes directory (default `~/notes/`).
- Optional daily-note shortcut (`notes today`) — v1 inclusion unresolved (see Open questions).
- Optional tag mutator (`notes tag <note> <tag>`) — vs. relying on editor-level frontmatter edits, unresolved (see Open questions).

## 8. Non-goals

- Web UI.
- Built-in sync (iCloud / git / Syncthing remain the user's responsibility).
- Encryption at rest or in transit.
- Plugin system.
- Cross-note linking, backlinks, or graph view.
- Built-in TUI viewer for search results.
- Background daemon or long-running index-maintenance service.

## 9. Constraints and preferences

- Notes are plain `.md` files in a folder; no proprietary format or database of record. Files must remain editable in any external editor.
- Distributable as a single binary or single-file `pipx` install. No Docker.
- Configurable notes directory; default `~/notes/`. Configuration mechanism (flag, env var, file, or combination) unresolved — see Open questions.
- Language preference: leaning Go for single-binary distribution; Python is the fallback if iteration speed wins. Final commitment deferred to ADR phase.
- No daemon, no network services, no telemetry.
- Frontmatter at top of each new note (creation date and an empty tag list).
- Editor handoff via `$EDITOR`.
- Search must remain fast on the thousands-of-notes scale.

## 10. Success signals

- The user stops reaching for sticky notes and ad-hoc `.txt` scratch files for short-term thoughts.
- Recovering an old note feels like seconds, not minutes.
- No new mental model required — file-on-disk reasoning is enough to be productive.
- Ships as a single binary or single-file `pipx` install with no runtime services.

## 11. Open questions

- **Tag model:** frontmatter list (`tags: [work, idea]`), inline `#hashtag` in the body, or both? If both, which wins for `--tag` filtering?
- **Filename convention:** `YYYY-MM-DD-slug.md` (date-prefixed, sorts naturally) vs. `slug.md` (cleaner, needs collision strategy)? How are title collisions handled?
- **Search backend:** plain `grep` / `ripgrep` shell-out (no index) vs. SQLite FTS (index that must stay fresh) vs. something else.
- **Configuration surface:** flag, env var, config file, or all three? Precedence order if all three?
- **Daily-note shortcut:** is `notes today` a v1 capability, or deferred to v0.2 / v1.x?
- **Tag mutator command:** is `notes tag <note> <tag>` worth shipping, or is `$EDITOR`-based frontmatter editing sufficient?
- **Release split:** explicit v0.1 (e.g. `new` + `find` + `ls`) → v0.2 (tagging, daily) carve-up, or a single v1 scope?
- **Language commitment:** lock Go now, or defer to the ADR phase?
- **Test discipline:** unit tests from day one, or smoke-tests-only for v0.1?
