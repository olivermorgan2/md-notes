# ADR-003: Filename convention — `YYYY-MM-DD-slug.md` with numeric collision suffix

**Status:** accepted
**Date:** 2026-05-02

## Context

`notes new "<title>"` writes a new `.md` file to the notes directory. The filename convention is unresolved in the PRD (`Design/prd-normalized.md`'s open questions): `YYYY-MM-DD-slug.md` vs. `slug.md`, plus a collision-handling strategy when two notes share a title.

Constraints from the MVP and product principles:

- Files must remain editable in any external editor (no opaque or hash-only names).
- The tool must "stay out of the way" — users should be able to find a note by skimming the directory.
- `notes ls` orders by modification recency, but a date-aware filename is still useful for human scanning, sync conflict resolution, and shell glob (`ls 2026-05-*.md`).
- `notes new` should be idempotent in spirit — calling it twice with the same title should never silently overwrite existing content (Risk 3 in `Design/build-out-plan.md`).

## Options considered

### Option A: `slug.md` (no date prefix)

- Pros:
  - Cleaner, shorter filenames; nicer in `ls` output.
  - Closer to user-given title — "design ideas" → `design-ideas.md`.
- Cons:
  - Title collisions are common in real use ("meeting notes", "todo", "ideas") and need an explicit collision strategy.
  - Loses chronological scan-by-glob ability without reading frontmatter.
  - Two notes with the same title in the same directory can no longer coexist without invented suffixes.

### Option B: `YYYY-MM-DD-slug.md` with numeric suffix on collision

- Pros:
  - Sorts naturally in `ls`; trivial to glob a date range (`2026-05-*.md`).
  - Collision becomes near-impossible — same title same day is rare. When it does happen, `YYYY-MM-DD-slug-2.md`, `-3.md`, … is unambiguous.
  - Sync tools (git, Syncthing) handle date-prefixed names well; conflict copies are obvious.
  - Date is encoded twice — in the filename and in frontmatter — but that redundancy is cheap and useful (filename for shell, frontmatter for parsing).
- Cons:
  - Filenames are longer than they strictly need to be.
  - Title collisions on the same day produce a `-2` suffix that the user may find ugly; mitigable by varying the title.

### Option C: `YYYYMMDDHHMMSS-slug.md` (timestamp prefix)

- Pros:
  - Collisions are vanishingly rare (would need same-second creation).
  - Strict chronological sort.
- Cons:
  - Filenames are noisy and hard to read.
  - Fights human scanning; the "stays out of the way" principle pushes back.
  - Privacy-adjacent — the filename leaks creation time-of-day, which feels excessive for a notes folder users may share or sync.

## Decision

Choose **Option B: `YYYY-MM-DD-slug.md` with numeric suffix on collision (`-2`, `-3`, …).** The date prefix gives natural chronological sort and makes shell globs useful without forcing a parse of frontmatter. Same-title-same-day collisions are rare in practice and the numeric-suffix strategy is deterministic and obvious. This keeps `notes new` idempotent in spirit — a second invocation with the same title creates a sibling file, never overwrites the first.

Slug derivation: lowercase, ASCII-fold (best-effort), replace runs of non-alphanumeric characters with a single `-`, trim leading/trailing `-`, and truncate to 64 characters. Empty slugs (e.g. title was all punctuation) fall back to `untitled` and rely entirely on the date and numeric suffix.

## Consequences

- **Easier:** human scanning of the notes directory; date-based shell globs; collision handling that requires no user thought.
- **Harder:** slugify implementation needs to handle Unicode → ASCII reasonably and have a deterministic collision-suffix loop; needs unit tests covering identical-title-same-day, identical-title-different-day, and empty-slug fallback.
- **Maintain:** the slug rules are documented in `README.md` so that users can predict filenames; `notes new` includes a unit test asserting collision suffixes are sequential and never overwrite an existing file.
- **Deferred:** whether to expose a `--slug` flag for explicit slug override; revisit only if the auto-slug produces visibly bad names in real use.
