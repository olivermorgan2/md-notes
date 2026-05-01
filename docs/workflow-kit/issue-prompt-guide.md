# Issue prompt guide

How to use [`notes/issue-prompt.md`](../notes/issue-prompt.md) — the
reusable Claude Code session prompt. The underlying decision is
[ADR-006](../Design/adr/adr-006-claude-code-execution-model.md):
issue-by-issue, plan-first execution.

A worked instance is at
[`notes/issue-prompt-sample.md`](../notes/issue-prompt-sample.md).

## What this is

`notes/issue-prompt.md` is the prompt you paste into Claude Code at the
start of each build session. One issue, one session. The template
gathers every piece of context the session needs (project, ADRs,
issue, scope, testing rules) and locks the workflow into a sequence:
read → plan → wait for approval → implement with tests → evaluation
summary.

## When to use it

- At the **start** of every issue session. Fill a fresh copy from the
  template; do not reuse a prompt across sessions.
- **One issue per session.** If the work splits into two issues, run
  two sessions.
- **Even for small issues.** The plan-first gate is cheap and the
  evaluation-summary discipline is what makes PRs reviewable.

If the work is so trivial that a prompt feels like overkill (a typo
fix, a one-line config tweak), skip the prompt and commit directly —
but do not skip it for anything that touches code paths or adds a
module.

## How to fill it before a session

Walk the template top to bottom. Each placeholder maps to a known
source:

| Placeholder | Where the value comes from |
|---|---|
| `{{PROJECT_NAME}}` | The repo name. |
| `{{ONE_LINE_PROJECT_DESCRIPTION}}` | The one-line description in `Design/mvp.md` (or `Design/prd-normalized.md` if no MVP yet). |
| `{{WORKFLOW_DOC_PATH}}` | Usually `Design/build-out-plan.md`. For repos that customise this, point at the actual workflow doc. |
| `{{ADR_FILE}}` and `{{ADR_ONE_LINE_SUMMARY}}` | The ADR(s) that drive this issue. Repeat the two-line block per ADR. If none applies, write `ADR: none — {{REASON}}` and explain. |
| `{{ISSUE_TITLE}}`, `{{ISSUE_NUMBER}}`, `{{MILESTONE}}`, `{{LABELS}}` | Straight from the GitHub issue. |
| `{{ONE_OR_TWO_SENTENCES}}` (Goal) | The "Goal" field of the GitHub issue body. |
| `{{ONE_PARAGRAPH}}` (Why it matters) | The "Why it matters" field of the issue body. |
| `Requirements` and `Acceptance criteria` | Copy from the issue body. Keep them as bullets — Claude treats each one as a discrete check. |
| `{{PRIMARY_FOLDERS}}` and `{{AVOID_FOLDERS}}` | The folders the work should and should not touch. Be specific — `src/parser/, test/parser/` is better than `src/`. The avoid list prevents scope creep. |
| `{{PROJECT_SPECIFIC_CONSTRAINT_OR_DELETE_THIS_LINE}}` | Anything not implied by the ADR or the testing rules. Delete the line if there is nothing to add. |

The `Evaluation & testing requirements` block and the
`Instructions for you` block are **fixed** — do not edit them per
session. They encode ADR-006 and should stay constant across sessions
so Claude's behaviour stays predictable.

## Evaluation summary at the end of a session

Step 5 of the template asks Claude for an evaluation summary. The
session is not done until this is produced. It must contain:

1. **Files changed** — every file created or modified, with a one-line
   note of what changed in each.
2. **Test results** — pasted output from the test runner: total,
   passed, failed, skipped. If no test runner, say so explicitly.
3. **New tests** — list each new test file and what it validates
   (happy path, edge cases, error handling).
4. **Regression check** — confirm pre-existing tests still pass.
5. **Manual verification steps** — only if some behaviour cannot be
   unit-tested. Skip the section if not applicable.
6. **Exact verification commands** — the commands you should run
   yourself to reproduce the test results.

A small example:

```
Evaluation Summary
==================
Files changed:
  - src/parser/gpx.ts (new) — parseGpx() and TrackPoint type
  - test/parser/gpx.test.ts (new)
  - test/fixtures/marathon.gpx (new) — real-race fixture

Test results:
  ✓ 14 passed, 0 failed, 0 skipped (vitest)

New tests:
  - test/parser/gpx.test.ts
    - parses a real-race fixture and returns the expected point count
    - returns first and last points matching hand inspection
    - throws TypedGpxError on malformed input (truncated, wrong root,
      missing timestamps)

Regression check:
  No pre-existing tests in this project yet. Confirmed nothing
  outside the parser module changed.

Commands to verify:
  npx vitest run
```

The summary is the quality gate before opening the PR. If it cannot
be produced honestly (tests skipped, regression not actually checked,
files glossed over), the work is not done — fix the gaps before
declaring the session finished.

## When the plan is wrong

The plan-first gate exists so that bad plans are caught before any
code lands. If the proposed plan misses scope, includes work that
should be a different issue, or wires the wrong module to the wrong
place, **push back before approving**. Re-prompting after code is
written is much more expensive than re-prompting after a plan.

A short "no, do not touch X" or "split this into two — let's do Y
first" is the right move. Approval is not a politeness ritual; it is
the moment the architecture decision becomes irreversible for this
session.
