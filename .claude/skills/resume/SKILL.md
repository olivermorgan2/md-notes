---
name: resume
description: Brief a fresh Claude Code session by reading Design/state.md and emitting a single short summary of phase, in-flight issue, recent work, blockers, and the next concrete action. Falls back to `gh` if state.md is missing, empty, or suspect.
---

# resume

Read `Design/state.md` and emit a one-message brief that brings a
fresh session up to speed in seconds: current phase, in-flight issue
(with its prompt and branch), the last few completed PRs, any
blockers, and the "continue here" pointer that names the next
concrete action.

This skill is the **read** side of the session-continuity contract
decided in [ADR-035](../../Design/adr/adr-035-state-md-session-continuity.md).
The **write** side lives in `prepare-issue`,
`claude-issue-executor`, `pr-review-packager` (each updates its zone
during the normal flow), and `pause` (which refreshes everything
mid-session for context-window-exhausting handoffs).

## When to use this skill

- At the **start** of a fresh Claude Code session on a long-running
  project, when the user wants context faster than re-running `gh`
  and grepping `prompts/`.
- After resolving a context reset, when `/pause` was used to write a
  rich handoff in the prior session.
- Any time the user wants a one-line snapshot of "where are we?"
  without leaving Claude Code.

If the project is small enough that everything in flight is in the
last commit message and the working tree, skip this skill — `git
log -1` is faster.

## What this skill does not do

- Does not modify any files. Read-only.
- Does not call `gh` on the happy path. The whole point of
  `state.md` is to make the brief instant; `gh` is a fallback only.
- Does not draft prompts, run tests, create branches, or open PRs.
  The brief points the user at the next action; the user (or
  another skill) takes it.
- Does not repair a malformed `state.md`. If the file is broken, the
  skill flags it and tells the user to fix it (or run `/pause` to
  refresh).

## Inputs

- **Required:** `Design/state.md`. If absent, the skill takes the
  `gh` fallback path described under **Edge cases**.
- **Optional (fallback only):** `gh` access to the GitHub repo whose
  remote is `origin`. Used when `state.md` is missing, empty, or
  flagged as suspect.

## Output

A single short message to the user, structured as:

- **Phase:** {{phase or `single`}}
- **In-flight:** `#NNN` ({{status}}) — prompt `prompts/issue-NNN-*.md`,
  branch `{{branch}}`
- **Recent:** up to 5 PRs, one line each (most recent first)
- **Blockers:** one line per blocker, or `none`
- **Continue here:** the verbatim text from the `continue-here` zone

The brief is plain markdown — no tables, no banners. The aim is for
the user to skim it in two seconds and know what to do next.

## Execution protocol

1. **Locate `Design/state.md`.** If absent, jump to step 7
   (`gh` fallback).
2. **Read the file.** Confirm the marker fences are intact:
   `state:phase`, `state:in-flight`, `state:recent`,
   `state:blockers`, `state:continue-here` each have a balanced
   `:start` / `:end` pair. If any pair is missing or unbalanced,
   stop with a one-line message naming the broken zone and suggest
   `/pause` to refresh.
3. **Parse each zone.** Pull the content between fences for every
   zone listed above. Trim whitespace. Skip placeholders that still
   look like `{{...}}` — treat them as empty.
4. **Sniff for staleness.** The file is suspect if any of:
   - the in-flight issue is `none` but `continue-here` names a
     specific issue;
   - the `Last updated` date is more than 14 days older than the
     most recent commit on `main` (cheap proxy for "the file
     forgot to update");
   - the in-flight issue's PR is already merged on GitHub (one
     `gh pr list --state merged --search "#NNN"` call, only when
     the file otherwise looks fine — this is the single allowed
     `gh` call on the happy path).
   If suspect, print a one-line warning before the brief and append
   "(consider `/pause` to refresh)".
5. **Render the brief.** Use the **Output** shape above. Keep each
   line under 100 characters. If the `recent` zone has fewer than 5
   entries, render only what's there — do not pad.
6. **Stop.** This skill never asks the user a follow-up question;
   the brief is the deliverable.
7. **`gh` fallback** (only when `Design/state.md` is missing or
   empty). Run, in order:
   - `gh pr list --state open --limit 5 --json number,title,url`
     for in-flight signals;
   - `gh pr list --state merged --limit 5 --json number,title,url,mergedAt`
     for recent work;
   - `gh issue list --state open --limit 5 --json number,title,url`
     for outstanding issues.
   Render the same shape as on the happy path, but mark the brief
   with a one-line header noting the fallback was used and that the
   project should consider running `/pause` to seed `Design/state.md`.

## Edge cases

- **`Design/state.md` missing.** Take the `gh` fallback path
  (step 7). Do not create the file — that is `/pause`'s job.
- **`Design/state.md` empty or all placeholders.** Same as missing —
  `gh` fallback.
- **Marker fences malformed.** Stop. Print which zone is broken and
  suggest `/pause` to refresh.
- **Outside a git repo / no `gh` auth.** On the happy path this is
  fine. On the fallback path, render whatever is in `state.md` (if
  anything) and note that `gh` was unavailable — do not block.
- **Conflict markers (`<<<<<<<`) in the file.** Stop. Tell the user
  to resolve the merge conflict per ADR-035's rule (most-recently-
  merged PR's version wins for the conflicting zone).
- **Phase zone literal `single`.** Render as `single` — projects
  without ADR-032 phases are first-class.

## Self-check before printing

- [ ] All five zones were located and parsed (or the broken-zone
  message was printed instead).
- [ ] No `{{...}}` placeholder leaked into the brief.
- [ ] At most one `gh` call was made on the happy path (the staleness
  sniff in step 4).
- [ ] The brief fits on a screen — under ~15 lines including the
  recent list.
- [ ] The "continue here" line is rendered verbatim from the file,
  not paraphrased.

## Handoff

`/resume` is a leaf — its output is the user's read. The user picks
up from the "continue here" line: typically by invoking
`/prepare-issue NNN`, opening a specific prompt, or reviewing an
open PR. If the user wants to refresh the file before continuing,
they run `/pause`.

See [`example.md`](example.md) for a worked invocation against a
populated `Design/state.md`.
