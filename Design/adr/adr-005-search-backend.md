# ADR-005: Search backend — in-process scan in Go

**Status:** proposed
**Date:** 2026-05-02

## Context

`notes find <query>` searches across the notes directory. The MVP requires search to feel "fast" at the thousands-of-notes scale — explicitly a first-class success criterion. The PRD's open questions framed three candidates: plain `grep` / `ripgrep` shell-out (no index, simple); SQLite FTS (index, must stay fresh); or "something else". The MVP also constrains the tool to **no daemon, no background process, no index-maintenance service** — which makes any persistent index that needs to be kept up-to-date suspect.

Risk 1 in `Design/build-out-plan.md` is "search backend choice forces a v0.2 rewrite". The chosen backend must survive v0.2's tag-filtering use case (which is `--tag work` and `#work` matching against frontmatter and inline body content respectively).

ADR-001 chose Go. ADR-004 chose minimal YAML frontmatter (`title`, `date`, `tags: []`). Both inform this decision.

## Options considered

### Option A: Shell out to `ripgrep`

- Pros:
  - Extremely fast at thousands-of-notes scale; ripgrep is the gold standard.
  - Minimal implementation work in md-notes itself — parse args, exec ripgrep, format output.
- Cons:
  - Adds a runtime dependency the user must install separately; breaks the "single binary, just runs" story.
  - On a fresh machine, `notes find` fails until ripgrep is installed — bad first-run UX.
  - Cross-platform shell-quoting and PATH issues are non-trivial.
  - Output parsing is fragile if ripgrep changes its format.

### Option B: In-process scan in Go (walk + read + match)

- Pros:
  - Zero runtime dependencies; preserves the single-binary distribution promise.
  - At thousands-of-notes scale (~1k–10k files, each ≤ a few KB), an in-process walk-and-match completes in tens of milliseconds on a warm cache and ~100–300 ms on a cold cache. Well inside the "feels fast" target.
  - Trivial to make tag-aware in v0.2 — frontmatter is already YAML and the scan is already reading every file; adding a `--tag` filter is a small change to the same code path.
  - Implementation is small: `filepath.WalkDir` + read + `bytes.Contains` (literal) or `regexp` (regex), with optional case-insensitivity.
  - Goroutine-fanout if needed later, without adding any user-visible state.
- Cons:
  - Slower than ripgrep at the 100k-notes scale — but the MVP target is thousands, not hundreds-of-thousands. Re-evaluable in a later ADR if the corpus grows.
  - No fancy features (regex highlighting, glob filters) without extra implementation.

### Option C: SQLite FTS with on-write index updates

- Pros:
  - Sub-millisecond search latency at any realistic scale.
  - Easy to add tag-aware queries in v0.2 by writing tags as a separate FTS column.
- Cons:
  - Index must stay fresh — `notes new` updates the index, but **any external edit** (the user opening `~/notes/*.md` in vim, or syncing from another machine) leaves the index stale. Detecting and reconciling this requires either an `mtime` walk on every search (which is most of an in-process scan anyway) or a watcher daemon (an explicit MVP non-goal).
  - Adds a SQLite dependency; in Go this is either cgo (`mattn/go-sqlite3`, complicates cross-compilation) or pure-Go (`modernc.org/sqlite`, larger binary).
  - Database file in the notes folder muddies the "notes are just files" principle; placing it elsewhere creates a hidden state file the user must understand.
  - For the actual MVP corpus size (thousands), FTS solves a problem the user does not have, at a real maintenance cost.

## Decision

Choose **Option B: in-process scan in Go.** The MVP corpus size (thousands) is exactly the regime where in-process scanning is fast enough to feel instant and where index maintenance is pure overhead. The single-binary distribution is preserved, the "no daemon, no index" constraint is honoured, and v0.2 tag filtering slots into the same code path with a small change. ripgrep would be faster but at the cost of a runtime dependency that breaks first-run UX. SQLite FTS would be faster still but introduces freshness problems that conflict with the "notes are just files" principle.

Implementation sketch: `filepath.WalkDir` over the configured notes directory, skipping hidden files; for each `.md` file, read into memory and match query (literal substring by default, regex with `-r`/`--regex`); print `path: line` for matches with a configurable context. Match is case-insensitive by default.

The Phase 2 smoke test from `Design/build-out-plan.md` — search latency on a synthetic 1,000-note corpus — is the empirical exit criterion. If latency exceeds ~500 ms on the author's laptop, the implementation needs a goroutine fan-out before this ADR is considered satisfied.

## Consequences

- **Easier:** single-binary distribution stays single-binary; v0.2 tag filtering is a small extension of the same scanner; no index freshness problem; "notes are just files" principle holds without exception.
- **Harder:** if the corpus grows to 100k+ notes, this approach will need to be revisited via a superseding ADR (likely toward an opt-in index, with the index treated as a cache the user can delete and rebuild).
- **Maintain:** a small scanner module with unit tests for literal and regex match, case-insensitivity, and hidden-file skipping; one smoke test on a 1,000-note synthetic corpus capturing latency on the author's machine; documented latency budget in `README.md` so a future regression is obvious.
- **Deferred:** ripgrep integration as an opt-in fast path (`--engine=rg`) and SQLite FTS as a cache layer — both are valid v0.2+ ADRs if real usage demands them. Neither is needed for the MVP.
