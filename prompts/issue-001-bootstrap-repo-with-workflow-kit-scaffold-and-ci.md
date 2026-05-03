You are working in my `md-notes` repository.

Context:
- A small CLI for terminal-resident developers that creates and finds plain-`.md` notes on disk so they can take scratch notes and recover them in seconds without an app lock-in.
- Follow the rules in `CLAUDE.md`.
- The workflow model is described in `{{WORKFLOW_DOC_PATH}}` <!-- TODO: fill in — no generic-project-workflow.md or docs/workflow-guide.md exists yet -->.

ADR:
- File: `Design/adr/adr-001-language-choice.md`
- Decision: Use Go as the implementation language; the MVP's single-binary distribution constraint is the dominant factor, and Python would need extra packaging tooling to approximate it.

GitHub Issue:
- Title: Bootstrap repo with workflow kit scaffold and CI
- Number: #1
- Milestone: M1
- Labels: infra

Goal
A clean Go-module repo into which the rest of v0.1 can land, with CI catching regressions from issue #2 onward.

Why it matters
Every downstream issue compiles and tests against this scaffold. Without it, no issue can declare "tests pass" or "CI green", and the v0.1 ship gate (a working `notes` binary on a clean machine) cannot be validated.

Requirements
- `go mod init github.com/olivermorgan2/md-notes` at the repo root
- `cmd/notes/main.go` with a stub `notes` command that prints help and exits 0
- `Makefile` (or `justfile`) with `build`, `test`, and `lint` targets
- GitHub Actions workflow that runs lint + `go test ./...` on push and PR
- `.gitignore` covering Go build artefacts (`/notes`, `/dist/`, `*.test`, `*.out`)

Acceptance criteria
- `go build ./...` produces a `notes` binary at the repo root
- `go test ./...` passes (no tests yet — exits 0 with "no test files")
- CI runs green on a push to a feature branch

Scope and constraints
- Primary folders to touch: repo root (`go.mod`, `go.sum`, `Makefile`, `.gitignore`), `cmd/notes/`, `.github/workflows/`
- Folders to avoid unless absolutely necessary: `Design/`, `prompts/`, `notes/`, `.claude/` — these are meta-docs and tooling, not part of the Go module
- This issue is scaffolding only — do not introduce business logic for `notes new`, `find`, or `ls`; those are issues #2–#5

Evaluation & testing requirements
- `go build ./...` exits 0 and produces an executable `notes` binary on the author's machine
- A push to a feature branch triggers the CI workflow; the workflow completes green
- All existing tests must continue to pass.
- If a change cannot be unit tested, document the manual verification.

Instructions for you
1. Read the relevant docs and existing files:
   - `CLAUDE.md`
   - `Design/adr/adr-001-language-choice.md`
   - any existing modules under `cmd/notes/`, `.github/workflows/`
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
