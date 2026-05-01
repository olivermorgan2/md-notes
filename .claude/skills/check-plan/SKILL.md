---
name: check-plan
description: Validate an ADR or issue prompt against version-locked criteria, return pass/fail with specific revisions, and iterate with the user up to 3 rounds. Chained from adr-writer and prepare-issue as a pre-commit gate; --skip-check opts out.
---

# check-plan

Read an ADR or issue prompt, detect its type, run the matching
checklist from
[`skills/check-plan/criteria.md`](criteria.md), and return either a
pass or a list of specific revisions per failed criterion. When
chained from `adr-writer` (ADR-007) or `prepare-issue` (ADR-013), it
runs against an in-memory rendering before disk write — so a failed
check leaves the working tree clean. Standalone, it operates on any
artefact already on disk.

This skill is the quality gate decided in
[ADR-034](../../Design/adr/adr-034-plan-checker.md). The contract is
opt-out (default on), with `--skip-check` available on the chained
producers for known-good rapid iteration.

## When to use this skill

- **Standalone, ad-hoc.** Run on any ADR or prompt already on disk
  to audit its quality: `/check-plan path/to/artefact.md`. Useful
  during ADR review, when reading an old prompt before re-running it,
  or when investigating execution-time defects.
- **Chained, automatic.** `adr-writer` and `prepare-issue` invoke
  this skill internally as their final gate before disk write. The
  user does not invoke it directly in that mode.
- **Re-validating after a manual edit.** When you've hand-edited an
  artefact and want a quick structural sanity check before committing.

If you only need a structural lint (no quality judgement), the
warnings will not fire — focus on the deterministic failures. If you
want a deeper review of an ADR's reasoning, use `/review` (built-in)
on the PR that introduces it; this skill is upstream of that.

## What this skill does not do

- Does not modify the artefact. Read-only with respect to the
  artefact under check.
- Does not commit, push, open PRs, or create branches.
- Does not invent criteria. The checklist is `criteria.md`; if a
  dimension you want is not there, add a row.
- Does not deeply reason about content quality (e.g. "is this
  decision wise?"). It runs structural checks plus a small set of
  textual heuristics. Human review is the layer above.
- Does not block on warnings. Only deterministic failures fail the
  gate. Warnings are surfaced and the user decides.

## Inputs

- **Required (standalone mode):** path to an artefact under check —
  either `Design/adr/adr-NNN-*.md` or `prompts/issue-NNN-*.md` or
  `prompts/_template.md`-shaped file at any path.
- **Required (chained mode):** the in-memory rendered artefact
  (passed by the calling skill) plus the artefact's would-be path
  for type detection.
- **Optional flag:** `--skip-check` — only relevant when the skill is
  invoked from a chained producer (`adr-writer` or `prepare-issue`).
  Standalone invocations ignore the flag (the user is already
  asking for a check).
- **Optional flag:** `--max-rounds=N` (default 3) — overrides the
  iteration cap. Reserved for advanced use; the default of 3 matches
  ADR-034's documented cap.

## Output

A single pass-or-fail report:

- **On pass:** `✅ check-plan PASS — <path>` plus a summary line
  listing any warnings that fired (each with its criterion ID).
  Warnings do not fail the gate.
- **On fail:** `❌ check-plan FAIL — <path>` plus one bullet per
  failed criterion, formatted as
  `<ID> (<determinism>): <fix hint with concrete pointer>`.
- **On yield (3-round cap reached):** `⚠ check-plan YIELD — <path>:
  N rounds exceeded` plus the unresolved failures listed for manual
  fix.

The report is plain markdown — no code fences, no banners.

## Execution protocol

Run these steps in order. Stop on the first hard error unless
otherwise noted.

1. **Resolve the artefact.** In standalone mode, take the argument
   path; verify the file exists. In chained mode, take the in-memory
   text plus the would-be path. If neither, stop with usage.
2. **Detect type.**
   - If the path matches `Design/adr/adr-*.md` AND the body has a
     `**Status:**` line, type = `adr`.
   - Else if the path matches `prompts/issue-*.md` OR the body has
     the canonical prompt section sequence (Context, ADR, GitHub
     Issue, Goal — first four headings in order), type = `prompt`.
   - Else stop with `"Cannot determine artefact type from path or
     body. Pass an ADR under Design/adr/ or a prompt under prompts/."`
     Never guess silently.
3. **Load criteria.** Read `skills/check-plan/criteria.md`. Parse
   the `## ADR criteria` and `## Prompt criteria` tables. Select the
   table matching the detected type. If `criteria.md` is missing,
   stop with `"skills/check-plan/criteria.md not found. The
   checklist must exist before this skill can run."`
4. **Run criteria.** For each row in the selected table:
   - Apply the check described in the *What to check* column.
   - On fail, record `(ID, determinism, fix hint)`.
   - On pass with warning conditions met, record a warning entry
     (only for `determinism=warning` rows that *could* fail
     deterministically were the check stricter — see criteria.md).
5. **Triage results.**
   - 0 deterministic failures → pass. Surface warnings if any.
   - ≥1 deterministic failures → fail.
6. **Standalone mode.** Print the report and stop.
7. **Chained mode — iterate.** Return the result to the caller.
   The caller (`adr-writer` or `prepare-issue`) presents the failures
   to the user, asks for revisions, applies them to the in-memory
   artefact, and re-invokes this skill for round 2. Round counter is
   maintained by the caller and passed in. After `--max-rounds`
   failed rounds (default 3), return a YIELD result; the caller
   surfaces the remaining failures and stops the loop.
8. **Audit trail.** Whether pass, fail, or yield, the result is
   surfaced verbatim to the user — never swallowed silently. When
   `--skip-check` is set on the chained producer, this skill is not
   invoked at all and the producer leaves a one-line breadcrumb
   (commit message or prompt body) per ADR-034.

## Type detection — detailed rules

ADR detection (in priority order):

1. **Path match.** `Design/adr/adr-(\d+)-.+\.md`. Sufficient.
2. **Body match.** First non-blank line is `# ADR-NNN: <title>` AND
   a `**Status:**` line appears in the first 10 lines. Sufficient.
3. **Frontmatter match.** YAML frontmatter declares `type: adr`.
   Used only if path and body don't match.

Prompt detection (in priority order):

1. **Path match.** `prompts/(issue-(\d+)-.+|_template)\.md`.
2. **Body match.** The first four section headings (case-
   insensitive, allowing `:` or `# `) appear in this order:
   Context, ADR, GitHub Issue, Goal. Sufficient.

If both ADR and prompt patterns match (highly unlikely), prefer the
path-match result. If neither matches, stop per step 2 above.

## Edge cases

- **Artefact path does not exist.** Standalone mode only. Stop with
  the original `gh`-style error verbatim.
- **`criteria.md` malformed (table parse fails).** Stop with the
  parse error and the offending line. Do not run partial checks.
- **All criteria are warnings.** Treated as pass with surfaced
  warnings. Per ADR-034, warnings never block.
- **`--skip-check` on a standalone invocation.** Ignore — the user
  is explicitly asking for a check. Print a one-line note that the
  flag is only meaningful in chained mode.
- **Round counter exceeds `--max-rounds`.** Yield. Do not run
  another round even if the caller asks.
- **Conflicting type signals (file under `Design/adr/` but body
  looks like a prompt, or vice versa).** Stop and ask the user to
  rename or fix the type ambiguity. Don't run the wrong checklist.
- **Empty artefact.** All deterministic structural criteria fail —
  fail-fast; do not surface every individual failure, just one
  summary line.

## Self-check before returning a result

- [ ] The artefact type was detected from a deterministic signal
  (path or body), not guessed.
- [ ] The selected criteria table matches the detected type.
- [ ] Every result entry cites a stable criterion ID from
  `criteria.md`.
- [ ] No criterion was silently skipped — `PROMPT-C4`'s
  conditional skip (build-out-plan absent) is documented in the
  output.
- [ ] The pass/fail/yield decision matches the deterministic-failure
  count: 0 → pass, ≥1 → fail, ≥1 after max rounds → yield.
- [ ] Warnings were surfaced even on a pass result — they are not
  swallowed.

## How chained mode works

`adr-writer` and `prepare-issue` invoke `/check-plan` as their final
gate before disk write. The flow:

1. Producer renders the artefact in memory.
2. Producer invokes `/check-plan` with the rendered text and the
   would-be path.
3. On pass, producer proceeds to disk write.
4. On fail, producer surfaces the failures, asks the user how to
   revise, applies the revision in memory, and re-invokes (round
   counter +1).
5. After 3 failed rounds OR an unrecoverable revision, the producer
   either yields (and stops with the working tree clean) or, if the
   user passed `--skip-check`, writes anyway and leaves a one-line
   breadcrumb in the audit trail.

The check runs *before* disk write, so a failed gate never produces
a half-written artefact. The user can always retry from a clean
state.

## Handoff

`/check-plan` is a leaf — its output is the final word per
invocation. In standalone mode, the user reads the report and
decides whether to revise the artefact. In chained mode, the
calling producer reads the result and either writes (pass) or
iterates (fail) per the protocol above.

If a criterion fails repeatedly across many artefacts, that is
information about either `criteria.md` (too strict?) or the
upstream template (too lax?). Take it to the next planning round
and decide whether to amend either.

See [`example.md`](example.md) for a worked walk-through (pass +
fail, standalone + chained).
