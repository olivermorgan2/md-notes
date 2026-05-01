---
name: claude-issue-executor
description: Drive a disciplined implementation session from a prepared issue prompt — plan-first, branch, incremental commits, tests alongside, evaluation summary
---

# claude-issue-executor

Run a full implementation session for a single GitHub issue, starting from
a prepared prompt file. The skill parses the prompt, proposes a plan and
**blocks for explicit user approval**, creates a feature branch from `main`,
implements in incremental commits that reference the governing ADR and
issue number, writes tests alongside code where the project supports them,
and produces an evaluation summary at the end.

This skill is the enforcement mechanism for the execution model decided
in [ADR-006](../../Design/adr/adr-006-claude-code-execution-model.md) and
the orchestration contract decided in
[ADR-014](../../Design/adr/adr-014-claude-issue-executor-skill.md).

## When to use this skill

- After `/prepare-issue` (ADR-013) has written a prompt file — typically
  `prompts/issue-NNN-short-title.md`.
- When starting an implementation session for a single GitHub issue.
- When the user invokes `/claude-issue-executor <path-to-prompt>`.

If no prompt file exists yet, stop and tell the user to run
`/prepare-issue` first (or to hand-author a prompt that matches
`prompts/_template.md`).

## What this skill does not do

- Does not create prompts. That is `/prepare-issue` (ADR-013).
- Does not create pull requests. That is `/pr-review-packager` (ADR-015).
  This skill **suggests** `/pr-review-packager` in its final summary but
  never auto-invokes it.
- Does not draft, amend, or accept ADRs. ADRs are authored by
  `adr-writer` and accepted by a human.
- Does not modify GitHub issues, labels, milestones, or the issue body.
- Does not duplicate logic from other skills. It is an orchestrator:
  prompt parsing, plan gate, branching, commits, tests, summary, handoff.

## Invocation contract

```
/claude-issue-executor <path-to-prompt-file-OR-issue-number>
```

- **Argument (optional):** either (a) a path to a prepared prompt file
  — preferred location `prompts/issue-NNN-short-title.md`, legacy
  fallback `notes/issueN-prompt.md`; or (b) a GitHub issue number, in
  which case the skill resolves the prompt path itself and may
  auto-chain `/prepare-issue` per the auto-chain protocol below
  (ADR-038).
- **No argument given:** list candidate prompt files from `prompts/`
  and, as a fallback, `notes/issue*-prompt.md`. Ask the user which to
  use. Do nothing else until they answer.

### Auto-chain `prepare-issue` (per [ADR-038](../../Design/adr/adr-038-tighten-prompt-step.md))

When the resolved issue has no `prompts/issue-NNN-*.md` on disk, the
skill auto-invokes `/prepare-issue NNN` and proceeds. The prep step
is logged prominently so the user sees it happen — auto-chain reduces
ceremony, not visibility.

When a prompt exists but is **stale** — its mtime is older than the
issue body's `updatedAt` (via `gh issue view <NNN> --json updatedAt`)
or older than any linked ADR's mtime — the skill asks the user
whether to regenerate. Default action: **regenerate with
confirmation, never silently**. On `n`, proceed with the existing
prompt and surface a one-line note in the eventual evaluation
summary.

The auto-chain runs *before* significance classification, so plan-
mode rhythm (ADR-039) gates the regenerated prompt as well.

## Inputs

- **Required:** a prepared prompt file containing the sections the skill
  relies on (see **Prompt validation**).
- **Required in the environment:** a clean working tree on `main` (or on
  a branch that tracks `main`). Nothing uncommitted.
- **Read (not modified):** `CLAUDE.md`, the ADR named in the prompt, and
  any other files the prompt points at.

## Outputs

- A feature branch created from `main`.
- A series of focused commits on that branch, each referencing the ADR
  and issue from the prompt.
- Tests alongside implementation, where the project has a test runner.
- An evaluation summary printed at the end of the session.

This skill does **not** push, does **not** open a PR, and does **not**
merge. Those belong to `/pr-review-packager` and the user's own review.

## Prompt validation

Before doing anything else, read the prompt file and confirm it contains
the sections the rest of the skill depends on. Required sections, matched
case-insensitively by heading or leading bold label:

1. **Context** — what project and what this session is about.
2. **ADR** — the governing ADR, e.g. `Design/adr/adr-NNN-...md`.
3. **GitHub Issue** — title and number, e.g. `Number: #16`.
4. **Goal** — what the issue should achieve.
5. **Requirements** — the work to be done.
6. **Scope and constraints** (or "Scope") — bounds of the session.
7. **Instructions for you** (or "Instructions") — the plan-first,
   commit-incrementally, evaluate-at-the-end directive.

If any required section is missing, **stop**. Report which sections were
not found and ask the user whether to fix the prompt (`/prepare-issue`
can regenerate it) or proceed without that section. Do not invent
missing content.

Also extract two fields for later:

- **ADR number** — the `NNN` in the ADR filename. Used in commit messages.
- **Issue number** — the `#N` from the GitHub Issue section. Used in
  commit messages and the branch name.

If either is absent, the prompt is malformed — treat as above.

## Branch naming

Derive the branch name from the prompt filename:

- `prompts/issue-017-pr-review-packager.md` → `pr-review-packager-skill` is
  one acceptable form; so is `issue-017-pr-review-packager`. Prefer the
  short, descriptive form that matches the repo convention in
  [`docs/github-setup.md` section 5](../../docs/github-setup.md#5-branch-naming-and-pull-requests):
  kebab-case, descriptive, no leading `feature-` or numeric-only names.
- If the prompt filename already matches repo convention, use it verbatim
  with any `issue-NNN-` prefix stripped when the short title alone is
  already distinctive.
- When in doubt, show the user the proposed branch name in the plan and
  let them amend it before approval.

Create the branch with `git checkout -b <branch>` from `main`. If the
user is not on `main`, switch to `main` and pull first — but only after
confirming the working tree is clean.

## Commit model

- **One logical change per commit.** Do not batch unrelated edits.
- **Commit message format:**
  `<verb>(<scope>): <what> (ADR-NNN, #issue)`
  where `<verb>` is `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, etc.
  and `<scope>` points at the primary folder touched
  (`skills`, `docs`, `templates`, etc.).
- **ADR-NNN and #issue come from the prompt**, not from the current
  conversation. Do not invent them.
- **Tests live in the same commit as the code they cover**, not in a
  separate "add tests" commit (see **Test-alongside**).
- **ADR index sync.** Before any commit that includes a file under
  `Design/adr/adr-*.md` (new, modified, or status-changed), run
  `bin/sync-adr-index` and stage `Design/adr/README.md` alongside the
  ADR files. Per ADR-023; keeps the index in sync with the
  filesystem on every commit that touches ADRs.

Example for this very issue:

```
feat(skills): add claude-issue-executor skill (ADR-014, #16)
docs(skills): wire claude-issue-executor into skills/README.md (ADR-014, #16)
```

## Test-alongside

Detect whether the project has a test runner by checking, in order:

- `package.json` → `scripts.test` (Node/JS/TS projects)
- `pytest.ini`, `pyproject.toml` with a `[tool.pytest.ini_options]` section
  (Python)
- `Cargo.toml` with a `[lib]` or `[[test]]` section (Rust)
- `go.mod` (Go — `go test ./...` is always available)
- any file ending in `*test*.*` in a conventional tests folder

If any is present, write tests as part of the logical commit that
introduces the code they cover. If the project has **no** test runner
(common for docs-only kits like this one — see this very repo), note
that in the evaluation summary and rely on documentation-style
verification instead.

Tests for a prose-only skill (a SKILL.md that contains no runnable code)
are usually unnecessary; a walk-through in the evaluation summary is the
accepted substitute.

## The plan gate — how plan-first is enforced

This is the single most important rule of the skill. Plan-first is
enforced by following this exact sequence every time:

1. Read the prompt file.
2. Validate it (see **Prompt validation**).
3. Read `CLAUDE.md` and the ADR referenced in the prompt.
4. **Produce a step-by-step plan in a single message** that includes:
   - the branch name you propose to create,
   - every file you will create or modify, with a one-line reason,
   - the commit sequence you intend (message + which files land in each),
   - whether tests will be written and where,
   - the verification steps you will run at the end.
5. **Stop.** Do not call any tool that writes to disk, runs git, runs
   tests, or otherwise mutates state. Wait for the user to answer.
6. Only proceed once the user replies with an explicit approval — e.g.
   "yes", "go ahead", "approved", "LGTM, proceed". A clarifying reply
   ("why are you doing X?") is **not** approval; address the question
   and re-present the plan.
7. If the user asks for changes, revise the plan and return to step 4.
8. If the user denies the plan outright, stop and offer to revise or to
   hand back control.

Operationally, when operating inside Claude Code: between steps 4 and 6
the assistant must not invoke Write, Edit, Bash (other than read-only
commands the user has already authorised), or any other mutating tool.
The only tool calls allowed between "plan proposed" and "plan approved"
are read-only ones needed to answer follow-up questions from the user.

This rule is load-bearing. Skipping it violates ADR-006 and defeats the
purpose of the skill.

See **Plan-mode rhythm (per ADR-039)** below for harness-level
enforcement that runs alongside this rule.

## Plan-mode rhythm (per ADR-039)

The chat plan-gate above is **convention** — the assistant follows the
8-step rule because the skill says so. Claude Code also ships a
**harness-level** mechanism: plan mode (toggled with
`shift+tab shift+tab`) locks the assistant out of all mutating tools
until the user explicitly exits plan mode with approval. The two are
complementary: the chat plan-gate operates *inside* plan mode when the
user has entered it, so both run together rather than competing.

This section defines when the executor requests plan-mode entry and
how it routes sessions of different sizes. The rules below implement
[ADR-039](../../Design/adr/adr-039-plan-mode-for-significant-tasks.md).

### Significant checklist

A session is **significant** if it meets any of:

- modifies 3+ files, OR
- edits any `skills/*/SKILL.md`, OR
- edits any `templates/*` file, OR
- edits `bin/*` (scripts, installer), OR
- modifies `.claude/settings*.json` or other harness config, OR
- otherwise carries blast radius beyond a single small fix.

### Trivial checklist

A session is **trivial** if it is one of:

- single typo,
- single-line doc tweak,
- status-line / config-default tweak,
- single-file rename within scratch space,
- `feature-ideas.md` status flip,
- ADR status flip (proposed → accepted),
- single-PR scope with no design decisions and no ADR linkage.

This list is the **single source of truth shared with ADR-038's
`--no-prompt` mode**. When either ADR's checklist evolves, both must
move together — ADR-038's mandatory content-boundary review is the
enforcement point.

### Hybrid path

When the executor parses the invocation and reads the prompt's Scope
section, it classifies the session against the two checklists and
routes to one of three branches:

- **Clearly-significant** → the assistant *requests* the user toggle
  plan mode and waits. The assistant cannot enter plan mode itself —
  only the user can press `shift+tab shift+tab`. After the user
  toggles, the assistant proposes the plan inside plan mode (the chat
  plan-gate's 8-step rule above runs here, with the harness-level
  lock providing belt-and-braces enforcement). The user approves and
  exits plan mode (`shift+tab` once); optionally enables auto-accept
  edits (`shift+tab` again) for the execution phase.
- **Clearly-trivial** → the executor proceeds with the chat plan-gate
  alone. No plan-mode request is made; the friction of toggling is
  not warranted for the work in question.
- **Borderline** → the executor asks once: *"Significant? yes / no /
  decide for me based on scope."* Proceed accordingly.

### Alignment-review obligation

When the trivial checklist above is amended, ADR-038's
`--no-prompt` criteria must be reviewed and aligned in the same change.
ADR-038 carries a mandatory content-boundary review obligation that
serves as the enforcement point for keeping the two checklists in
lockstep.

## `--no-prompt` mode

Per [ADR-038](../../Design/adr/adr-038-tighten-prompt-step.md), the
executor accepts a `--no-prompt` flag that **skips prompt generation
entirely** and runs from the issue body alone. The prompt artefact is
not written. A one-line breadcrumb — `issue executed without prompt
per ADR-038` — is appended to the first commit's message so the
audit trail survives.

### When `--no-prompt` is appropriate

The criteria for `--no-prompt` are exactly the **Trivial checklist**
above (lines beginning *"A session is **trivial** if it is one of:"*).
That checklist is the **single source of truth**; the `--no-prompt`
mode does not duplicate it. If the trivial checklist evolves, this
mode's criteria evolve in lockstep — see the **Alignment-review
obligation** section just above.

### Auto-detect (with confirmation)

When the user invokes `/claude-issue-executor <issue-number>` (no
explicit `--no-prompt`), the executor auto-detects candidates
*conservatively*:

- The issue has **zero `ADR-NNN` references** in its body, AND
- The issue has at least one of the labels `chore`, `docs`, or
  `bugfix-trivial`.

When both conditions hold, the executor *suggests* `--no-prompt`
mode and asks for confirmation: *"Issue looks trivial — skip prompt
generation? (yes / no)"*. On `yes`, proceed without prompt. On `no`,
fall back to the standard auto-chain path. The suggestion never
short-circuits silently.

### Override

Explicit `--no-prompt` overrides auto-detection — no confirmation is
asked. This is for the user who already knows the issue is trivial
and wants the lowest-ceremony path. The breadcrumb is still left in
the commit message.

### Interactions

- **Plan-mode rhythm** (ADR-039) still applies. `--no-prompt`
  affects the *prompt* step, not the plan-mode gate. A trivial
  issue with `--no-prompt` typically also classifies as
  clearly-trivial against the significance checklist, so plan mode
  is not requested. But if the executor is ever invoked with
  `--no-prompt` on something that *is* significant (e.g. a multi-
  file refactor mislabelled `chore`), the significance gate still
  fires — the two are independent.
- **`/check-plan`** (ADR-034) does not run when `--no-prompt` is
  set, because there is no rendered prompt to check. The skip is
  noted in the evaluation summary alongside the breadcrumb.
- **`Design/state.md`** updates (ADR-035) still happen. The
  `state:in-flight` zone records `Status: verified` (or `executing`
  mid-session) regardless of whether a prompt was generated.

## Session protocol — end to end

1. **Resolve the prompt (auto-chain `prepare-issue` per
   [ADR-038](../../Design/adr/adr-038-tighten-prompt-step.md)).**
   - If the argument is a *path*, treat as the prompt file directly.
   - If the argument is an *issue number*, resolve the prompt at
     `prompts/issue-NNN-*.md`. If absent, **auto-invoke
     `/prepare-issue NNN`** and log the prep step prominently. If
     present, run the staleness check: compare the prompt's mtime
     against `gh issue view <NNN> --json updatedAt` and against the
     mtime of every linked ADR. If any is newer, ask the user
     whether to regenerate (default: yes; never silent). On `n`,
     proceed with the existing prompt and note this in the
     evaluation summary.
   - If no argument was given, list candidate prompt files from
     `prompts/` and (fallback) `notes/issue*-prompt.md`. Ask the
     user which to use.
2. **Parse invocation and classify.** With the prompt resolved
   (or `--no-prompt` set — see **`--no-prompt` mode** below),
   **classify the session against the significance checklist** (see
   **Plan-mode rhythm** above): if clearly-significant, request the
   user toggle plan mode before continuing; if borderline, ask once;
   if clearly-trivial, proceed with the chat plan-gate alone.
3. **Preflight.** Confirm the working tree is clean. If dirty, stop and
   ask the user to commit, stash, or discard. Confirm the current branch
   is `main` or agree with the user to switch.
4. **Read and validate the prompt.** If malformed, stop and report.
   Skipped when `--no-prompt` is set; the issue body is used directly.
5. **Read the referenced ADR** and `CLAUDE.md`.
6. **Propose the plan** (see **The plan gate**).
7. **Wait for approval.** No edits until explicit yes.
8. **Create the branch** from `main`.
9. **Implement, commit incrementally.** Each commit is one logical
   change with the required message format. Tests land with code. If
   a new significant boundary appears mid-session (scope expands
   beyond what was classified at start, e.g. an unforeseen
   `skills/*/SKILL.md` edit or a new `templates/*` change), pause and
   re-flag for plan-mode entry. Do not silently cross the boundary.
   When `--no-prompt` is set, append the breadcrumb `issue executed
   without prompt per ADR-038` to the first commit's message.
10. **Verify.** Run the project's tests if any exist. Run any
    verification step called out in the prompt's "Evaluation & testing
    requirements" section.
11. **Update `Design/state.md` if present.** Per
    [ADR-035](../../Design/adr/adr-035-state-md-session-continuity.md),
    rewrite the `state:in-flight` zone: keep `Issue` and `Prompt`,
    set `Branch` to the current branch, set `Status: verified`. The
    `recent` zone is **not** touched here — it tracks merged PRs and
    is the responsibility of `/pr-review-packager`. If the file is
    absent, skip silently. If marker fences are broken, do not
    rewrite; report the broken zone in the evaluation summary's
    follow-ups list and suggest `/pause`.
12. **Evaluation summary.** Print the final summary (see **Evaluation
    summary** below). For `--no-prompt` runs, the summary explicitly
    notes the bypass for the audit trail.
13. **Suggest handoff.** Tell the user the next step is
    `/pr-review-packager` to package a PR. Do not invoke it.

## Evaluation summary

Print a single structured message that covers:

- **What changed** — bulleted list of files created or modified, grouped
  by purpose.
- **Commits** — the commit messages in order, with SHAs when known.
- **Verification performed** — test runner output, build output, or a
  walk-through if the project has no runnable tests.
- **Follow-ups** — anything noted during the session that is out of
  scope for this issue but worth capturing (e.g. "this also enables
  issue #18"). Do not silently defer requirements that were in scope.
- **Commands the user should run** to reproduce the verification
  themselves. Concrete shell lines, not prose.
- **Next step** — `/pr-review-packager` once the session is reviewed.

The format mirrors the "evaluation summary" step in every
`notes/issueN-prompt.md` — see those files for reference.

## Edge cases

- **Prompt file not found.** Search `prompts/` and `notes/` for the
  closest match and ask the user which they meant. Never guess silently.
- **Prompt file malformed.** Report the missing sections by name and
  ask whether to fix the prompt or to proceed with explicit user
  guidance for the gaps.
- **Working tree dirty on entry.** Refuse to proceed. Do not auto-stash.
  Ask the user what to do.
- **Target branch already exists.** Ask whether to (a) switch to the
  existing branch and continue, (b) pick a new name, or (c) delete the
  existing one (only with explicit confirmation and only when the user
  is sure nothing there is needed). Never force-delete silently.
- **User denies the plan.** Offer to revise. If the user wants to stop
  entirely, stop — do not keep pushing variants.
- **Tests fail during implementation.** Stop, report, and ask — do not
  paper over failing tests to keep the commit cadence.
- **ADR or issue number missing from the prompt.** Malformed — see
  **Prompt validation**. Do not substitute a placeholder in commit
  messages.

## Handoff

On successful completion, the final message suggests:

> Next step: `/pr-review-packager` to draft a pull request from this
> branch, using the commit history and the governing ADR.

This skill does not invoke `/pr-review-packager` itself. The user opens
the next session deliberately, preserving a human review checkpoint
between implementation and PR publication (ADR-015).

## Alignment check

Before finishing the session, confirm:

- [ ] The plan was proposed and explicitly approved before any edits.
- [ ] A feature branch was created from `main` (not committed to `main`
  directly).
- [ ] Every commit message includes `ADR-NNN` and `#issue` from the
  prompt.
- [ ] Tests, if applicable to the project, landed with the code.
- [ ] An evaluation summary was printed.
- [ ] `/pr-review-packager` was suggested, not auto-invoked.

If any box is unchecked, the skill has drifted — say so in the
evaluation summary rather than hiding it.

See [`example.md`](example.md) for a worked session driven by
`prompts/issue-017-pr-review-packager.md`.
