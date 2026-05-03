# md-notes — MVP

**Last updated:** 2026-05-01

## Product name

md-notes

## One-line description

A small CLI for terminal-resident developers that creates and finds plain-`.md` notes on disk so they can take scratch notes and recover them in seconds without an app lock-in.

## Product goal

Replace the "sticky notes and random `.txt` scratch files" workflow with a single CLI that writes dated markdown files into a user-owned folder and finds them again by half-remembered phrase. Success for the first release is "I stop reaching for sticky notes, and I can recover an old note in seconds."

## Target users

### Primary user

The author — a terminal-resident developer who takes frequent scratch notes during the day and currently loses them in folders.

## Core problem

Scratch notes pile up across random files and folders, and recovering an old one by half-remembered phrase takes minutes instead of seconds. App-based tools (Obsidian, Notion) lock notes into a proprietary store; the user wants plain `.md` files they own and can edit or sync with anything.

## Product principles

1. **Files on disk are the source of truth.** No proprietary format, no database of record. Any external editor can open and edit a note.
2. **The tool stays out of the way.** No new mental model — file-on-disk reasoning is enough to be productive.
3. **No services, no daemons, no network.** Single binary or `pipx` install; no Docker, no background processes, no telemetry.
4. **Sync, encryption, and collaboration are someone else's job.** The user picks iCloud / git / Syncthing (or none) — the tool doesn't care.
5. **Fast at thousands of notes.** Search latency is a first-class success criterion.

## MVP scope

### In scope

- `notes new "<title>"` — create a new note with date + empty-tags frontmatter, opened in `$EDITOR`.
- `notes find <query>` — full-text search across titles and bodies, fast at thousands-of-notes scale.
- `notes ls` — list notes by modification recency, most-recent first.
- Configurable notes directory (default `~/notes/`).
- Empty `tags: []` written into every new note's frontmatter — so v0.2 tagging does not require a migration of older notes.

### Out of scope

Product-level non-goals from the PRD (permanent):

- Web UI.
- Built-in sync (iCloud / git / Syncthing remain the user's responsibility).
- Encryption at rest or in transit.
- Plugin system.
- Cross-note linking, backlinks, or graph view.
- Built-in TUI viewer for search results.
- Background daemon or long-running index-maintenance service.

Deferred by this MVP scoping (candidates for v0.2):

- `notes find --tag <tag>` and `#tag` filtering.
- `notes tag <note> <tag>` mutator command.
- `notes today` daily-note shortcut.

## Primary outputs

A single binary or `pipx`-installable script. The user runs `notes new`, `notes find`, or `notes ls` from a terminal; notes live as plain `.md` files in a configurable folder on disk.

## Success criteria

The MVP succeeds if the user can:

1. Run `notes new "<title>"` and have a dated, frontmatter-bearing markdown file open in `$EDITOR` within seconds.
2. Run `notes find <half-remembered phrase>` against a directory of hundreds-to-thousands of notes and get the right one back, fast.
3. Run `notes ls` and see the notes they touched today and yesterday at the top.
4. Stop reaching for sticky notes and ad-hoc `.txt` scratch files for short-term thoughts.

## Deferred to later

- **Tag filtering and tag mutation** — v0.2 work. The MVP writes `tags: []` into frontmatter so deferred tagging does not need a migration.
- **Daily-note shortcut (`notes today`)** — v0.2 if used enough to justify; otherwise drop.
- **Filename collision strategy beyond a sensible default** — handled minimally in v0.1; revisited if collisions actually happen.

## Acceptance criteria for this document

This MVP statement is acceptable when it:

- names a clear product and primary user — yes,
- lists what is in and out of scope without ambiguity — yes,
- and can drive the build-out plan, ADRs, and issue backlog without further interpretation — yes (see `Design/build-out-plan.md`).

Open items: tag model, filename convention, search backend, configuration surface, and Go-vs-Python language commitment are unresolved and surfaced as ADR candidates in `Design/build-out-plan.md`.
