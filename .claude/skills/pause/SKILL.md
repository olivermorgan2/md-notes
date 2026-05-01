---
name: pause
description: Refresh Design/state.md to current truth (phase, in-flight issue, recent work, blockers, continue-here) and optionally write a richer notes/handoff-YYYY-MM-DD.md for context-window-exhausting session handoffs.
---

# pause

Refresh `Design/state.md` so it accurately reflects right now —
phase, in-flight issue, recent work, blockers, "continue here"
pointer — and (optionally) write a richer
`notes/handoff-YYYY-MM-DD.md` that a fresh Claude Code session can
read end-to-end to pick up cleanly.

This skill is the **session-handoff** counterpart to `/resume`. The
two together implement ADR-035's session-continuity contract:
`/pause` writes the truth down, `/resume` reads it back. The other
write side — incremental updates to `state.md` during the normal
flow — is owned by `prepare-issue`, `claude-issue-executor`, and
`pr-review-packager`; `/pause` is for the *manual* refresh moments
those skills do not cover.

See [ADR-035](../../Design/adr/adr-035-state-md-session-continuity.md).

## When to use this skill

- **Mid-session, before a context reset.** When the session has
  accumulated enough state that a `/clear` would lose load-bearing
  context.
- **End of day.** Before stepping away from a long-running project,
  to leave a clean pickup point for tomorrow.
- **After a non-trivial detour.** If the session diverged from the
  prepared prompt (e.g. hit a blocker, switched to a sibling issue),
  a manual refresh records what actually happened.
- **First adoption.** The first time a project adopts ADR-035, run
  `/pause` to seed `Design/state.md` from the current `gh` and `git
  log` state.

If the session is short and the next step is in the working tree or
the open prompt, skip this skill — the routine writes from
`prepare-issue` / `claude-issue-executor` / `pr-review-packager` are
sufficient.

## What this skill does not do

- Does not commit `Design/state.md` automatically. The skill writes
  to the working tree; the user decides whether to commit it
  alongside session work or separately.
- Does not push, open PRs, or call `gh pr` / `gh issue` mutators.
  Read-only against GitHub.
- Does not rewrite earlier `notes/handoff-*.md` files. Each handoff
  is its own dated file; old ones stay until the user prunes them.
- Does not create `Design/state.md` from nothing without confirming
  the project intends to adopt ADR-035. On a project where the file
  is genuinely absent, the skill asks before seeding.

## Inputs

- **Required:** a git repo whose `origin` points at the GitHub repo.
- **Optional:** existing `Design/state.md`. If present, only the
  zones whose facts changed are rewritten; other zones (and any
  out-of-fence editorial commentary) are preserved.
- **Optional:** the path or number of the prompt currently being
  worked on. If provided, used to set `in-flight` and
  `continue-here`. If omitted, the skill infers from the most-recent
  branch name and open prompt files.
- **Optional flag:** `--handoff` — also writes
  `notes/handoff-YYYY-MM-DD.md` with a richer narrative.

## Output

- **`Design/state.md`** — refreshed in place. Marker fences
  preserved; only the zones whose facts changed are rewritten.
- **`notes/handoff-YYYY-MM-DD.md`** (with `--handoff` only) — a
  free-form richer handoff modelled on
  [`notes/handoff-2026-04-30-v-next-batch-resume.md`](../../notes/handoff-2026-04-30-v-next-batch-resume.md).
  Covers state of work, what's open, recommended next step, standing
  rules, working-tree state, and "what NOT to do".

## Execution protocol

1. **Confirm the repo context.** `git rev-parse --show-toplevel` and
   `gh repo view --json nameWithOwner` to confirm we are in a kit
   target project. If either fails, stop and tell the user.
2. **Detect or seed `Design/state.md`.** If absent, ask the user
   whether to seed from `templates/state-template.md`. On `no`,
   stop. On `yes`, copy the template and fill the `Last updated`
   header with today's date in `YYYY-MM-DD` form; leave zones empty
   for the next steps to populate.
3. **Determine current phase.** If `Design/build-out-plan.md` exists
   and contains `## Phase` blocks (per ADR-032), pick the earliest
   unfinished phase. Otherwise write the literal `single`. Confirm
   with the user before overwriting an existing non-empty `phase`
   zone.
4. **Determine the in-flight issue.** In order of preference:
   (a) the explicit issue number / prompt path the user passed;
   (b) the most recent `prompts/issue-*.md` whose mtime is within
   the last 14 days and whose corresponding branch exists;
   (c) `gh issue list --assignee @me --state open --limit 1` if (a)
   and (b) are empty.
   If all three are empty, write `none` to the `in-flight` zone.
5. **Refresh recent work.** Run
   `gh pr list --state merged --limit 5 --json number,title,url,mergedAt`,
   pull the linked ADR if any (`grep -oE 'ADR-[0-9]+'` on PR title /
   body), and render one line per PR. Keep the rolling list at most
   5 entries; older entries drop off.
6. **Refresh blockers.** Ask the user: "Any blockers to record?"
   Default `none`. Free-form one line per blocker.
7. **Refresh `continue-here`.** Compose one short paragraph naming
   the next concrete action. Inputs: the in-flight issue's prompt
   path, the branch name, and any outstanding step the user names.
   Show the proposed paragraph and confirm before writing.
8. **Write the file.** Rewrite only the zones whose contents
   changed. Preserve marker fences; preserve out-of-fence content
   verbatim.
9. **Line-cap check.** After writing, run `wc -l Design/state.md`.
   If the file exceeds 100 lines, show the user a one-line warning
   and suggest pruning the oldest `recent` entry. Do not auto-prune
   beyond the 5-entry rolling window.
10. **Optional rich handoff.** If `--handoff` was passed, render
    `notes/handoff-YYYY-MM-DD.md` modelled on the example linked
    above. Sections:
    - **Date** and **Where we left off** (one paragraph),
    - **State of v-current** (or current epic): merged / open / queued tables,
    - **Recommended next step on resume**,
    - **Standing rules** (point at memory or CLAUDE.md),
    - **Working tree state at handoff**,
    - **What NOT to do on resume**.
    If today's handoff file already exists, ask before overwriting.
11. **Report.** Print the path(s) written and a one-line diff
    summary (which zones changed; whether the handoff file was
    written).

## Edge cases

- **`Design/state.md` exists but marker fences are broken.** Stop.
  Tell the user which zone is malformed; do not silently rewrite.
- **`Design/state.md` over the line cap before this run.** Warn but
  do not refuse. The skill can still refresh zones; the warning
  guides manual pruning.
- **No `gh` auth.** The recent-work refresh fails gracefully — the
  zone is left as-is with a one-line note appended ("recent-work
  refresh skipped: gh unavailable"). Other zones are still updated.
- **Detached HEAD or working tree dirty for `Design/state.md`.**
  The skill writes to the working tree as usual; the user picks
  whether to stash, commit, or discard.
- **Multiple open PRs by the user.** The recent-work refresh shows
  the most recent 5 merged PRs only — open PRs do not enter the
  `recent` zone. They surface via `gh` when `/resume` is invoked
  with the staleness sniff.
- **`--handoff` requested but `notes/` directory missing.** Create
  it (one `mkdir -p notes` is acceptable).

## Self-check before writing

- [ ] All required marker fences are intact in the new file
  contents (each zone has matched `:start` / `:end`).
- [ ] No `{{...}}` placeholder remains in any rewritten zone.
- [ ] The `Last updated` header was set to today's date.
- [ ] The `recent` zone has at most 5 entries.
- [ ] The `continue-here` paragraph names a *concrete* next action,
  not a vague aspiration.
- [ ] If `--handoff` was passed, the handoff file is dated, named
  `notes/handoff-YYYY-MM-DD-<slug>.md` if a slug is supplied, else
  `notes/handoff-YYYY-MM-DD.md`.

## Handoff

`/pause` is a leaf — the user reviews the diff (`git diff
Design/state.md`) and decides whether to commit. The next session
runs `/resume` to read what was just written.

If `/pause` was invoked because the context window is about to
exhaust, the natural next step is `/clear` followed by `/resume` in
the fresh session — the optional `notes/handoff-YYYY-MM-DD.md` is
the longer-form supplement when `state.md` alone is insufficient.

See [`example.md`](example.md) for a worked invocation.
