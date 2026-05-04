You are working in my `md-notes` repository.

Context:
- A small CLI for terminal-resident developers that creates and finds plain-`.md` notes on disk so they can take scratch notes and recover them in seconds without an app lock-in.
- Follow the rules in `CLAUDE.md`.
- The workflow model is described in `{{WORKFLOW_DOC_PATH}}` <!-- TODO: fill in — no generic-project-workflow.md or docs/workflow-guide.md exists; closest candidate is docs/workflow-kit/workflow-guide.md -->.

ADR:
- File: `Design/adr/adr-002-configuration-surface.md`
- Decision: Use all three configuration layers — `--dir` flag > `MD_NOTES_DIR` env > `~/.config/md-notes/config.toml` (`[paths] dir`) > default `~/notes/` — with TOML as the file format and only `[paths].dir` recognised in v0.1.

GitHub Issue:
- Title: Implement config loader with documented precedence (ADR-002)
- Number: #2
- Milestone: M1
- Labels: feature

Goal
A `Config` struct populated by a deterministic resolver that every command uses, implementing the four-layer precedence chain locked in ADR-002.

Why it matters
Per ADR-002, this is the foundation every command depends on; getting precedence wrong here breaks every other command. `notes new`, `find`, and `ls` (issues #3–#5) all read the resolved notes directory before doing anything else, so a wrong-precedence regression here would silently route every operation to the wrong folder.

Requirements
- `internal/config/config.go` with `Load(flag string) (*Config, error)` — `flag` is the `--dir` value (empty string when not set), and `Config` carries at minimum the resolved `NotesDir` field
- XDG-aware path resolver for `~/.config/md-notes/config.toml` (respects `XDG_CONFIG_HOME` when set)
- TOML parsing with a single `[paths]` section — only `dir` recognised; unknown keys ignored, not errors
- Default fallback to `~/notes/` (expand `~` → `$HOME`) when none of flag/env/file are set
- Unit tests covering each precedence layer and each combination

Acceptance criteria
- `flag > env`: `Load("/tmp/x")` with `MD_NOTES_DIR=/tmp/y` returns `/tmp/x`
- `env > file`: `Load("")` with `MD_NOTES_DIR=/tmp/y` and a TOML file pointing elsewhere returns `/tmp/y`
- `file > default`: `Load("")` with no env and a TOML file `dir = "/tmp/z"` returns `/tmp/z`
- `default-when-nothing-set`: `Load("")` with no env and no file returns the expanded `~/notes/`
- Missing config file is **not** an error — the resolver falls through to the default
- Malformed TOML produces a clear error mentioning the file path

Scope and constraints
- Primary folders to touch: `internal/config/` (new package)
- Folders to avoid unless absolutely necessary: `cmd/notes/` (entry point gets one wiring line at most), `Design/`, `prompts/`, `notes/`, `.claude/`, `docs/`
- This issue introduces the project's first runtime dependency (a TOML parser). Pick one of `github.com/BurntSushi/toml` or `github.com/pelletier/go-toml/v2`; the first introduces `go.sum`. Either is fine — call it out in the plan.

Evaluation & testing requirements
- `go test ./internal/config/...` passes; coverage on the resolver is ≥80%
- Each of the six acceptance-criteria scenarios has its own test (table-driven is fine)
- Tests use `t.TempDir()` and `t.Setenv()` so they do not touch the user's real `$HOME` or environment
- All existing tests must continue to pass.
- If a change cannot be unit tested, document the manual verification.

Instructions for you
1. Read the relevant docs and existing files:
   - `CLAUDE.md`
   - `Design/adr/adr-002-configuration-surface.md`
   - any existing modules under `internal/config/`
   - any existing tests related to the modules you will change
2. Propose a short, step-by-step implementation PLAN for this issue, including:
   - new files or modules to create,
   - existing files to modify,
   - key functions or structures,
   - your verification or test plan.
3. Wait for my approval of the plan before making any edits.
4. After I approve, implement the plan:
   - keep changes focused on this issue's scope,
   - commit incrementally with messages referencing the ADR and issue
     (e.g. "feat(scope): add thing (ADR-NNN, #NN)").
5. At the end, provide an evaluation summary:
   - what changed,
   - verification steps performed,
   - any follow-up work needed for later issues,
   - exact commands I should run to inspect the result myself.

Do not start editing files until I explicitly approve your plan.
