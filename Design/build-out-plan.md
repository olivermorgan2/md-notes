# md-notes — Build-Out Plan

**Last updated:** 2026-05-01
**Granularity:** coarse

## Objective

Ship the MVP of md-notes defined in `Design/mvp.md`: a small CLI that creates dated markdown notes with frontmatter, lists them by recency, and finds them by full-text search across a configurable folder of `.md` files. The plan covers everything from repo bootstrap to a locally-installable v0.1 release in a single one-week delivery cut.

## Build strategy

1. Bootstrap the repo with the workflow kit's standard scaffold.
2. Resolve the five foundational ADRs (language, configuration surface, filename convention, frontmatter shape, search backend) before writing implementation code.
3. Implement `notes new`, `notes find`, and `notes ls` against the chosen backend, sharing the same notes directory and config loader.
4. Validate on a real-use corpus, package for distribution, write README, tag v0.1.

## Scope

- In scope: every in-scope capability from `Design/mvp.md`.
- Out of scope: every item in the "Out of scope" list of `Design/mvp.md` (both permanent non-goals and deferred-to-v0.2 items).
- Assumptions: the user has a local toolchain for the chosen language; the user has `$EDITOR` set; the user has at least a few real notes (or is willing to create them) to dry-run search against.

## Success criteria

The plan is complete when the user can:

1. Install md-notes as a single binary or via `pipx`.
2. Run `notes new "<title>"`, get a dated frontmatter-bearing `.md` in their notes directory, and edit it in `$EDITOR`.
3. Run `notes find <phrase>` against a corpus of hundreds-to-thousands of notes and find the right one in under a second.
4. Run `notes ls` and see the most recently modified notes at the top.
5. Tag the repo `v0.1` with passing tests and a README.

## Repository structure

```text
md-notes/
  src/             ← CLI entry point and command implementations
  test/            ← unit + smoke tests
  Design/          ← mvp.md, build-out-plan.md, adr/, prd-normalized.md
  notes/           ← per-issue prompts and working notes
  README.md
  CLAUDE.md
```

(Concrete sub-layout — e.g. `cmd/notes/` for Go vs. `md_notes/` for Python — is settled by the language ADR.)

## Phases

1 phase chosen — within the `coarse` band — because md-notes is a one-week CLI with three small commands sharing one notes directory and one config loader; splitting into multiple phases would create artificial seams in work that ships together. Per the skill's single-phase fallback, this is the back-compat flat-plan path: downstream skills (`issue-planner`, `workflow-docs`, `/release`) treat this as one implicit phase.

### Phase 1 — md-notes v0.1

- **Goal:** ship md-notes v0.1 — a CLI that creates, finds, and lists plain `.md` notes in a configurable folder.
- **ADR dependencies:** language; configuration surface; filename convention; frontmatter shape; search backend (all five resolved before implementation, see "Decisions needing ADRs" below).
- **Deliverables:**
  - Repo scaffold and CI.
  - Config loader resolving the notes directory (default `~/notes/`, overridable per the configuration-surface ADR).
  - `notes new "<title>"` — writes a dated `.md` file with frontmatter (`date`, `tags: []`) and opens it in `$EDITOR`.
  - `notes find <query>` — full-text search across titles and bodies via the chosen backend.
  - `notes ls` — list notes by modification recency, most-recent first.
  - Unit tests for filename generation, frontmatter rendering, config precedence, and find/ls output.
  - Smoke test on a synthetic ~1,000-note corpus capturing find latency.
  - One-week dry-run on the author's real notes.
  - Packaging as a single binary (Go) or `pipx`-installable script (Python).
  - README with install + usage + configuration + known limits.
  - CHANGELOG seeded; v0.1 git tag and GitHub release.
- **Exit criterion:** **MVP ships.** Concretely, all four MVP success criteria from `Design/mvp.md` hold during the one-week dry-run on the author's real notes, and a clean install (binary download or `pipx install`) on a fresh machine yields a working `notes` command.

## Milestone recommendation

| Milestone | Focus |
|---|---|
| M1 | md-notes v0.1 — full MVP delivery |

## Initial issue backlog

### M1 — md-notes v0.1

- Bootstrap repo with workflow kit scaffold and CI
- ADR: language choice (Go vs. Python)
- ADR: configuration surface (flag / env / file precedence)
- ADR: filename convention and collision handling
- ADR: frontmatter shape and v0.2-tagging compatibility
- ADR: search backend (grep/ripgrep vs. SQLite FTS vs. other)
- Implement config loader with documented precedence
- Implement `notes new "<title>"` (filename, frontmatter, `$EDITOR` handoff)
- Implement `notes find <query>` against the chosen backend
- Implement `notes ls` ordered by modification recency
- Unit tests for filename generation, frontmatter, config, find, and ls
- Smoke test on a synthetic 1,000-note corpus; record search latency
- Dry-run on real notes for one week; capture issues as v0.1.x or v0.2 candidates
- Package as single binary (Go) or `pipx`-installable script (Python)
- Write README (install, usage, configuration, known limits)
- Seed CHANGELOG; tag `v0.1`; publish GitHub Release

## Testing strategy

- Unit tests for the create path (filename generation, frontmatter rendering, config precedence) and the find/list path (query parsing, output ordering).
- One smoke test on a synthetic ~1,000-note corpus to validate the search-latency success criterion.
- Manual dry-run on the author's real notes for one week as the only honest check that the tool stays out of the way.
- No integration tests required for v0.1; there is no backend, no daemon, and no network.

## Risks and mitigations

### Risk 1 — Search backend choice forces a v0.2 rewrite

Picking the simplest thing (e.g. `grep` shell-out) may not survive once tagging lands; picking the heaviest (SQLite FTS) over-invests for v0.1. *Mitigation:* the search-backend ADR must explicitly consider v0.2's tag-filtering use case and pick a backend that survives it, even if v0.1 only uses a subset.

### Risk 2 — Frontmatter shape in v0.1 forces migration in v0.2

If v0.1 omits `tags: []` from new notes, v0.2 has to migrate. *Mitigation:* the MVP explicitly writes `tags: []` into every new note's frontmatter, codified as a unit-test check.

### Risk 3 — Filename collisions silently overwrite notes

Two notes with the same title could clobber each other. *Mitigation:* the filename-convention ADR must define a deterministic collision strategy (e.g. date-prefixed names, suffix on collision), and `notes new` includes a unit test for it.

### Risk 4 — Single-phase plan masks scope creep

With only one delivery cut, "just one more thing" pressure is harder to resist. *Mitigation:* the in-scope list is fixed at the five capabilities in `Design/mvp.md`; anything else captured during the dry-run becomes a v0.1.x or v0.2 issue, not an in-flight scope expansion.

### Risk 5 — Packaging eats the back half of the week

`pipx`/binary packaging on a fresh OS can absorb a day. *Mitigation:* if packaging is fighting back at the end of the week, ship as `pip install -e .` (or `go install`) for v0.1 and revisit polished packaging in v0.1.x — the exit criterion is "a working `notes` command on a clean machine", not "perfect distribution".

## Acceptance criteria for this document

This build-out plan is acceptable when it:

- matches the MVP statement — yes,
- sequences work in a realistic shape — yes (single phase, one-week cut, `coarse` band),
- identifies initial ADRs or decisions — yes (see below),
- and produces a practical milestone and issue structure — yes (1 milestone, ~16 issues).

## Decisions needing ADRs

Surfaced for handoff to `adr-writer`. Each item is a single architectural question.

1. **Language choice: Go vs. Python.** Affects distribution shape (single binary vs. `pipx`), iteration speed, and stdlib availability for filename / frontmatter / search work. User leans Go.
2. **Configuration surface: flag / environment variable / config file precedence.** Affects UX and how the notes directory is resolved on every command.
3. **Filename convention and collision handling.** `YYYY-MM-DD-slug.md` vs. `slug.md`; collision strategy when two notes share a title. Affects sort behaviour, readability, and idempotence of `notes new`.
4. **Frontmatter shape and v0.2-tagging compatibility.** Exact YAML shape (`date`, `tags: []`, anything else) so v0.2 tagging does not need a migration.
5. **Search backend: grep/ripgrep shell-out vs. SQLite FTS vs. other.** Affects latency at thousands of notes, install footprint, and whether v0.2 tag filtering needs an index.
