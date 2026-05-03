# ADR-001: Language choice — Go

**Status:** accepted
**Date:** 2026-05-02

## Context

md-notes is a small CLI for taking and finding markdown notes. The MVP (`Design/mvp.md`) and normalized PRD (`Design/prd-normalized.md`) name distribution as a constraint: ship as a **single binary or `pipx`-installable script**, no Docker, no daemon, no network services. The user signalled a lean toward Go for single-binary distribution, with Python as the fallback if iteration speed wins. Search must remain fast at the thousands-of-notes scale, which makes a runtime-free language attractive.

The choice affects every downstream ADR — distribution shape, packaging story, available search-backend libraries, filename / frontmatter parsing libraries, and developer iteration speed.

## Options considered

### Option A: Go

- Pros:
  - Compiles to a single static binary; matches the MVP distribution constraint exactly.
  - No runtime dependency on the user's machine — no Python/Node/runtime version mismatch.
  - Strong stdlib for filesystem traversal, regex, YAML (via `gopkg.in/yaml.v3`), and process control (`exec.Command` for `$EDITOR`).
  - Fast cold-start; relevant for a CLI run interactively many times a day.
  - Cross-compilation for macOS/Linux/Windows is one `GOOS`/`GOARCH` flag.
- Cons:
  - Slightly more ceremony than Python for a small project (build step, explicit error handling, no REPL).
  - Smaller frontmatter / markdown ecosystem than Python's; some libraries are thinner.
  - Author iteration speed is somewhat slower than Python on the first few iterations.

### Option B: Python (with `pipx`)

- Pros:
  - Faster to write and easier to extend; closer to author's iteration preference.
  - Mature ecosystem for YAML (`PyYAML`), markdown, and CLI scaffolding (`click`, `typer`).
  - `pipx` makes single-user installs nearly as ergonomic as a binary.
- Cons:
  - Runtime dependency on Python ≥ 3.x being installed; cross-machine portability is less crisp.
  - `pipx` install time is non-trivial on cold environments.
  - Cold-start time is meaningfully higher than Go for a CLI invoked many times per day.
  - Single-file binary distribution requires extra tooling (`pyinstaller`, `shiv`) and complicates the release path.

## Decision

Choose **Option A: Go.** The MVP's distribution constraint (single binary, no runtime deps) is the dominant factor; Go matches it natively while Python needs extra packaging tooling to approximate. The slight ergonomic loss vs. Python is acceptable for a tool the author uses many times a day where cold-start latency is felt directly. Python remains the sensible fallback if Go ergonomics block progress.

## Consequences

- **Easier:** single-binary distribution; cross-compilation; consistent cold-start latency; trivial `$EDITOR` handoff via `exec.Command`.
- **Harder:** initial scaffolding (build step, module setup); some small-utility code is more verbose than Python equivalents.
- **Maintain:** a `go.mod`/`go.sum`, a `Makefile` or `justfile` for common tasks, and a release workflow that emits per-OS binaries.
- **Deferred:** any decision about a Python-based extension or scripting layer; revisit only if a v1.x feature genuinely needs one.
