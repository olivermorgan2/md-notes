---
name: prepare-issue
description: Auto-fill an issue prompt from a GitHub issue number, linked ADRs, and the build-out plan, then write it to prompts/issue-NNN-short-title.md
---

# prepare-issue

Take a GitHub issue number, pull the issue's title, body, labels,
milestone, and linked ADR references via `gh`, read any relevant
project context (most importantly `Design/build-out-plan.md`), fill
`prompts/_template.md`, and write the result to
`prompts/issue-NNN-short-title.md` for the user to paste into a fresh
Claude Code session.

This skill automates the biggest point of friction between "issue
exists on GitHub" and "Claude Code session is briefed" — see
[ADR-013](../../Design/adr/adr-013-prepare-issue-skill.md).

## When to use this skill

Use when the user has a GitHub issue ready to work on and wants a
filled prompt file in `prompts/`. Typical invocation:

```
/prepare-issue 17
```

The argument is a single GitHub issue number (no `#`). The skill is
**read-only** with respect to GitHub — it never creates, edits, or
comments on issues.

If the user does not yet have a GitHub issue backlog, run
`issue-planner` first (ADR-011). If the filled prompt already exists
and is still accurate, skip this skill and re-use the existing file.

## What this skill does not do

- Does not create or modify GitHub issues. Use `issue-planner` for
  that.
- Does not start the implementation session. The filled prompt is a
  handoff artifact; the user (or `claude-issue-executor`, ADR-014)
  runs it in a fresh session.
- Does not modify `prompts/_template.md`. It reads it only.
- Does not invent content when issue metadata is missing. Missing
  fields are left as clearly-marked TODOs.
- Does not overwrite an existing prompt file without explicit user
  confirmation.

## Inputs

- **Required:** a single GitHub issue number, passed as the skill's
  argument (e.g. `/prepare-issue 17`).
- **Implicit repo context:** the current working directory must be
  inside a git repo whose `origin` is the GitHub repo containing the
  issue. `gh` infers the repo from git remotes.
- **Optional flag:** `--skip-check` — opt out of the `/check-plan`
  pre-write gate (per [ADR-034](../../Design/adr/adr-034-plan-checker.md)).
  Default is on (gate runs); the flag is documented as opt-out for
  rapid iteration on known-good drafts only. When set, the skill
  writes the prompt despite any deterministic-criteria failures and
  appends a one-line breadcrumb to the prompt body —
  `<!-- /check-plan was skipped via --skip-check per ADR-034 -->` —
  so the bypass is auditable.

## Output

- **File:** `prompts/issue-NNN-short-title.md` in the target project.
- **Shape:** exactly the section order of `prompts/_template.md`
  (ADR-008): Context, ADR, GitHub Issue, Goal, Why it matters,
  Requirements, Acceptance criteria, Scope, Evaluation, Instructions.
- **Filename convention:** zero-padded to 3 digits, kebab-case short
  title derived from the issue title (see "Short-title derivation"
  below).

## Data sources and how the skill reads them

The skill consults three sources, in priority order:

1. **GitHub issue** (required) — via
   `gh issue view <N> --json title,body,labels,milestone,url`.
   This is the primary source for almost every field.
2. **Linked ADRs** (optional) — any `ADR-NNN` token or
   `Design/adr/adr-NNN-*.md` link mentioned anywhere in the issue
   body, title, or labels. For each referenced ADR, read the file
   from `Design/adr/` and extract its title and one-line decision
   summary.
3. **Build-out plan** (optional) — `Design/build-out-plan.md`, if it
   exists. Grep it for mentions of the issue number or the issue's
   core noun phrase and pull a short contextual paragraph.

`gh` is the only allowed tool for GitHub reads. Do not call the
REST/GraphQL API directly.

## Execution protocol

Run these steps in order. Stop on the first failure unless otherwise
noted.

1. **Validate input.** Confirm the argument is a positive integer.
   If not, ask the user for the issue number and stop.
2. **Check the template exists.** Read `prompts/_template.md`. If it
   is missing, tell the user to run Issue #12's work first and stop.
3. **Fetch the issue.** Run
   `gh issue view <N> --json title,body,labels,milestone,url`. If
   `gh` exits non-zero (issue not found, no auth, wrong repo), print
   the error message verbatim and stop. Do not attempt to guess or
   proceed with partial data.
4. **Parse the issue JSON.** Extract:
   - `title` → issue title string
   - `body` → markdown body
   - `labels[].name` → comma-separated label list
   - `milestone.title` → milestone name (or "none")
   - `url` → canonical GitHub URL (not used in the template but
     useful for logging)
5. **Resolve ADR references.** Scan the issue title + body for both
   forms:
   - `ADR-NNN` tokens (case-insensitive, with or without zero-pad).
   - Explicit path links like `Design/adr/adr-NNN-short.md` or
     `adr-NNN-short.md`.
   For each unique `NNN`, glob `Design/adr/adr-<NNN>-*.md`. If a
   match is found, read the file and extract:
   - the filename (for the `ADR.File` line),
   - the ADR title from the `# ADR-NNN: Title` heading,
   - a one-line decision summary from the `## Decision` section
     (first sentence is usually enough).
   If no ADR references are found, mark the ADR section with
   "ADR: none — no ADR referenced in this issue." per the template
   comment.
6. **Read build-out plan if present.** If
   `Design/build-out-plan.md` exists, grep it for the issue number
   and for noun phrases from the issue title. Capture any short
   matching paragraph as supplementary context to include in the
   "Context" section or the "Why it matters" section where it fits
   naturally. If the file does not exist, skip silently — this is
   normal in a freshly-initialized kit repo.
7. **Derive the short title.** See "Short-title derivation" below.
8. **Fill the template.** Read `prompts/_template.md` and substitute
   every `{{PLACEHOLDER}}` with the extracted value. For placeholders
   where data is not available, leave the placeholder in place and
   append ` <!-- TODO: fill in -->` so the user sees the gap
   immediately.
9. **Show the user the filled prompt in the chat** as a fenced
   markdown block. Ask explicitly: "Write this to
   `prompts/issue-NNN-short-title.md`? (yes / edit / cancel)".
   **Do not write the file before this confirmation.**
10. **Pre-write check (per [ADR-034](../../Design/adr/adr-034-plan-checker.md)).**
    Unless `--skip-check` was passed, after the user confirms with
    `yes`, invoke `/check-plan` against the in-memory filled
    prompt. On pass, proceed to step 11 (file-exists check). On
    fail, surface the failures (each citing its criterion ID — e.g.
    `PROMPT-C1`) to the user, ask how to revise, apply the
    revision in memory, re-show the updated prompt for confirmation,
    and re-invoke `/check-plan` for round 2. After 3 failed rounds,
    yield: surface the remaining failures and stop without writing
    the file. Warnings are surfaced but do not block.
    `--skip-check` short-circuits the gate and the skill proceeds
    to step 11 with a one-line breadcrumb appended to the prompt
    body.
11. **Handle the file-exists case.** Before writing, check whether
    `prompts/issue-NNN-short-title.md` already exists. If it does,
    show a diff between the existing file and the new content and
    ask explicitly whether to overwrite. Default to "no".
12. **Write the file** only after explicit confirmation. Report the
    absolute path and a one-line summary of what was filled vs. left
    as TODO.
13. **Update `Design/state.md` if present.** Per
    [ADR-035](../../Design/adr/adr-035-state-md-session-continuity.md),
    rewrite two zones:
    - `state:in-flight` → set `Issue: #NNN`, `Prompt:` to the just-
      written path, `Branch: n/a` (executor will set it later),
      `Status: prepared`.
    - `state:continue-here` → one short paragraph naming the next
      action: `"Run /claude-issue-executor prompts/issue-NNN-…md"`.
    Marker fences (`<!-- state:<zone>:start --> / :end -->`) bound
    each zone; rewrite only the bytes between the fences. Other
    zones (`phase`, `recent`, `blockers`) are left untouched. If
    `Design/state.md` is absent, skip silently — this is normal in
    a kit repo or a project that has not adopted ADR-035. If the
    file exists but its marker fences are broken, do not attempt to
    repair; tell the user and suggest `/pause` to refresh.

## Short-title derivation

Derive the short title from the issue title:

1. Lowercase the title.
2. Strip a trailing `(adr-nnn)` or `(#nn)` suffix if present.
3. Strip leading verb boilerplate ("build ", "add ", "create ",
   "write ") only if the remaining title is still descriptive —
   otherwise keep the verb.
4. Replace any run of non-alphanumeric characters with a single `-`.
5. Trim leading and trailing `-`.
6. Truncate to ~50 characters at the nearest `-` boundary.

Examples:

| Issue title                                                    | Short title                       |
|---------------------------------------------------------------|-----------------------------------|
| "Build pr-review-packager skill (ADR-015)"                    | `pr-review-packager-skill`        |
| "/prepare-issue skill (ADR-013)"                              | `prepare-issue-skill`             |
| "Write docs/workflow-guide.md — end-to-end workflow guide"    | `docs-workflow-guide-end-to-end`  |
| "Create repository skeleton and project-local installation"   | `repository-skeleton-project-local` |

The resulting filename is
`prompts/issue-NNN-<short-title>.md` with `NNN` zero-padded to three
digits.

## Template-filling rules

The template uses `{{PLACEHOLDER}}` slots. Map each slot as follows.

| Slot                               | Source                                                                 |
|------------------------------------|------------------------------------------------------------------------|
| `{{PROJECT_NAME}}`                 | Git repo name (from `git remote get-url origin` or `basename $PWD`).   |
| `{{ONE_LINE_PROJECT_DESCRIPTION}}` | First line of the project's `README.md` if available; else TODO.       |
| `{{WORKFLOW_DOC_PATH}}`            | `generic-project-workflow.md` if it exists; else `docs/workflow-guide.md` if it exists; else TODO. |
| `{{ADR_FILE}}`                     | Filename of the first resolved ADR (e.g. `adr-013-prepare-issue-skill.md`). |
| `{{ADR_ONE_LINE_SUMMARY}}`         | First sentence of the `## Decision` section of that ADR.               |
| `{{ISSUE_TITLE}}`                  | Issue title verbatim.                                                  |
| `{{ISSUE_NUMBER}}`                 | The input number (not zero-padded here — the template already has `#`). |
| `{{MILESTONE}}`                    | Milestone title from `gh`, or "none".                                  |
| `{{LABELS}}`                       | Comma-separated label names.                                           |
| `{{ONE_OR_TWO_SENTENCES}}`         | The "Goal" section of the issue body if present; else first paragraph. |
| `{{ONE_PARAGRAPH}}`                | The "Why it matters" section of the issue body if present; else TODO. |
| `{{REQUIREMENT_N}}`                | Bullets from the issue body's "Requirements" section.                  |
| `{{CRITERION_N}}`                  | Bullets from the issue body's "Acceptance criteria" section.           |
| `{{PRIMARY_FOLDERS}}`              | From the issue body's "Scope" section, if present; else TODO.          |
| `{{AVOID_FOLDERS}}`                | From the issue body's "Scope" section, if present; else TODO.          |
| `{{PROJECT_SPECIFIC_CONSTRAINT_OR_DELETE_THIS_LINE}}` | Delete line if no project-specific constraint; else fill. |
| `{{EVALUATION_REQUIREMENT_N}}`     | Bullets from the issue body's "Evaluation" section.                    |

If the issue body closely mirrors the template (as most issues
produced by `issue-planner` will), the mapping is near-direct. If the
body is freeform, do best-effort extraction and mark gaps as TODOs
rather than inventing content.

**Multiple ADRs:** if the issue references more than one ADR, repeat
the `- File:` / `- Decision:` pair for each, in the order they first
appear in the issue body.

## Edge cases

- **Invalid or non-numeric argument** → ask for the issue number and
  stop.
- **Issue not found or auth error** → print `gh`'s error verbatim and
  stop. Do not retry.
- **No ADR referenced** → write
  `ADR: none — no ADR referenced in this issue.` per the template
  comment. Do not block on this.
- **ADR referenced but file missing** → include the reference token
  verbatim with ` <!-- TODO: ADR file not found in Design/adr/ -->`.
- **`Design/build-out-plan.md` missing** → skip silently; the kit
  repo itself does not have this file.
- **`prompts/_template.md` missing** → stop with a clear message
  pointing at Issue #12.
- **File already exists at target path** → show a diff and ask
  before overwriting. Default no.
- **Multi-line label list** → join with `, `.
- **No milestone** → fill with the string `none`.
- **`Design/state.md` missing** → skip the state-update step
  silently. Consistent with how the build-out-plan read is handled.
- **`Design/state.md` marker fences broken** → do not rewrite. Tell
  the user which zone is malformed and suggest `/pause` to refresh
  the file in place.
- **`/check-plan` gate fails** → the gate runs *after* user-confirm
  but *before* disk write, so a failed check leaves the working
  tree clean. Iterate up to 3 rounds in step 10, then yield without
  writing.
- **`/check-plan` gate yields** (3 rounds exceeded) → stop. Surface
  remaining failures to the user as items to fix manually. The
  in-memory prompt is discarded; nothing is written.
- **`--skip-check`** → bypass the gate entirely for this invocation.
  Append the documented breadcrumb to the prompt body before
  writing, and proceed to step 11. The flag is single-use; future
  `/prepare-issue` invocations re-enable the gate.

## Review-before-write checkpoint

The user always sees the filled prompt as chat output before the
file is written. This preserves the review gate from ADR-013's
Option A (the chosen option). Never skip this step, even if the
filled prompt looks clean.

## Self-check before writing

- [ ] The issue number was fetched successfully via `gh`.
- [ ] Every `{{PLACEHOLDER}}` in the template is either filled or
  explicitly marked `<!-- TODO: fill in -->`.
- [ ] The ADR section names a file that exists in `Design/adr/`, or
  explicitly says "none".
- [ ] The short title is kebab-case, ≤ 50 chars, and starts with an
  alphanumeric.
- [ ] The filename is `prompts/issue-NNN-<short-title>.md` with
  `NNN` zero-padded to three digits.
- [ ] The user explicitly confirmed the write.
- [ ] `/check-plan` passed (deterministic criteria), or
  `--skip-check` was explicitly set with a recorded breadcrumb in
  the prompt body.

If any fail, fix and re-show before writing.

## Handoff

The filled `prompts/issue-NNN-*.md` is the handoff point into a
Claude Code implementation session. The user either pastes it into a
fresh session manually, or (once ADR-014 lands) runs
`claude-issue-executor` to consume it. This skill is done once the
file is written and the path is reported.

See [`example.md`](example.md) for a worked invocation on a real
issue.
