# Workflow Guide — Idea to Shipped Release

A practical walkthrough of the full Claude Code Workflow Kit flow, from a
rough idea to a tagged release. Follow this guide end-to-end the first
time you use the kit; after that, treat it as a cheat sheet.

**Who this is for:** a solo developer or small team who has installed
the kit into a new target project and wants to ship a feature without
piecing the flow together from individual skill docs.

**What you end up with:** a working, tested, merged feature on `main`
with an accompanying ADR, issue, PR, and (at milestone boundaries) a
tagged GitHub Release.

This guide is the condensed, action-oriented companion to the detailed
[`generic-project-workflow.md`](../generic-project-workflow.md). Where
the reference doc explains *why*, this guide focuses on *what to run
next*. Skills are referenced by their slash name (e.g. `/idea-to-prd`);
each one has its own `SKILL.md` under `.claude/skills/<name>/` in your
target project with the full interface spec.

---

## 1. Prerequisites

You should have already:

- Installed Git, GitHub CLI, and Claude Code, and authenticated `gh`.
- Created the target project on GitHub and cloned it locally.
- Copied the kit's skills into `.claude/skills/` and rendered
  `CLAUDE.md` at the project root.

If any of that is missing, stop here and complete
[`docs/install.md`](install.md) first. The rest of this guide assumes
you are sitting inside the target project's root with Claude Code open.

---

## 2. The happy path, one pass

One pass = one feature shipped. Steps (a) through (c) are usually done
once per MVP scoping round. Steps (d) through (g) loop per issue. Steps
(h) through (j) close out a milestone and a release.

### 2.a From idea or PRD to MVP

Pick the skill that matches what you have in hand:

| You have… | Run | You produce |
|---|---|---|
| A rough idea, no doc | `/idea-to-prd` | A lightweight PRD in `Design/` |
| A standard or custom PRD | `/prd-normalizer` | A normalized PRD in the kit's internal format |
| A normalized PRD | `/prd-to-mvp` | An MVP scope statement: in-scope, out-of-scope, success signals |

Run them in sequence if you have a rough idea: `/idea-to-prd` →
`/prd-normalizer` (if you later want to reshape it) → `/prd-to-mvp`.

**Granularity (per ADR-036).** `prd-to-mvp` and `/planning` accept
an optional `--granularity={coarse|standard|fine}` flag that sets a
target band for phase count: `coarse` aims for 1–3 phases, `standard`
(the default) for 5–8, `fine` for 8–12. Bands are *targets*, not hard
caps — the skill picks the actual count for your project and includes
an inline justification in the rendered build-out-plan. The choice is
recorded as a `**Granularity:**` line in `Design/build-out-plan.md`,
so re-runs are consistent without you having to remember the flag.
Precedence (highest first): explicit flag, then the stored value in
the build-out-plan, then the default `standard`. Pick `coarse` for
weekend or one-week projects, `standard` for typical multi-month
builds, `fine` for large multi-quarter scope where each phase warrants
its own milestone.

**What unblocks next:** a written MVP statement that scopes the first
release. You now know what decisions you need to lock in.

### 2.b Design decisions to ADRs

For each meaningful architectural choice surfaced by the MVP — a library
pick, an integration, a data model, a deployment target — write an ADR.

- **Run:** `/adr-writer`
- **What you see:** a drafted ADR in `Design/adr/adr-NNN-<short-title>.md`
  following the kit's template (context, options, decision, consequences).
- **What you produce:** one accepted ADR per decision.

One ADR per decision. If the skill drafts several, accept them one at a
time — each ADR should stand alone.

**What unblocks next:** a small set of accepted ADRs that define the
backlog's shape.

### 2.c Plan to backlog

Turn the MVP plus ADRs into a tracked GitHub backlog.

- **Run:** `/issue-planner`
- **What you see:** a batch of drafted issues (title, body, labels,
  milestone, ADR reference) presented for your approval before any are
  created.
- **What you produce:** GitHub issues created via `gh issue create`,
  plus a GitHub Project board (per ADR-012) organising them.

The skill is GitHub-first: issues live on GitHub, not in local files.
The Project board gives you a kanban view of the backlog.

**Phased vs single-phase plans (per ADR-032).** When
`Design/build-out-plan.md` contains multiple `## Phase N: <name>`
blocks, `/issue-planner` creates one GitHub milestone per phase and
assigns each issue to its phase's milestone; `/release` defaults to
one tag per phase. When the plan has only one Phase block — or no
Phase headings at all — every downstream skill behaves identically
to the pre-ADR-032 flat path: one milestone for the whole project,
one release at the end. The single-phase fallback is the right shape
for small projects (weekend or one-week scope); multi-phase shape
fits multi-month builds where each phase delivers standalone value.
See [`examples/projects/phased-podcast-pipeline/`](../examples/projects/phased-podcast-pipeline/)
for a worked 3-phase example, or [`examples/projects/kb-lookup/`](../examples/projects/kb-lookup/)
and [`examples/projects/slug-utils/`](../examples/projects/slug-utils/)
for the single-phase shape.

**Anti-pattern: tuning granularity to dodge scope.** The
`--granularity` flag (documented in step 2.a) is a knob for *how
many phases your scope decomposes into*, not a knob for shipping a
large project faster. If you find yourself reaching for `coarse` on
a multi-month project so the build-out plan looks shorter, the right
move is to cut MVP scope (re-run `prd-to-mvp` with a tighter scope
question), not to compress the build-out into fewer phases. Phases
are how the work is *paced*; scope is what is *delivered*. Don't
confuse the two. ADR-036's "Consequences" section calls this out as
the explicit anti-pattern this section guards against.

**What unblocks next:** a list of issue numbers you can work through
one at a time.

### 2.d One issue at a time to a prompt

Before writing any code, generate a per-issue Claude Code session
prompt.

- **Run:** `/prepare-issue <issue-number>`
- **What you see:** the skill pulls the issue body and its linked ADR
  via `gh`, reads the relevant build-out plan section, and fills the
  prompt template.
- **What you produce:** `prompts/issue-NNN-<short-title>.md` (per
  ADR-008), ready to paste into a fresh Claude Code session.

This is a deliberate handoff — the prompt is written but execution
starts separately, so you get a review checkpoint between planning and
building.

**Auto-chain (per ADR-038).** Since ADR-038, this step is *optional*:
`/claude-issue-executor <issue-number>` will auto-invoke
`/prepare-issue` itself when no `prompts/issue-NNN-*.md` exists for
the issue, logging the prep step prominently. You can still run
`/prepare-issue` explicitly if you want a review checkpoint between
prompt generation and execution — that is the recommended path for
significant work. **Stale prompts** (the issue body or any linked ADR
has been updated since the prompt was last written) trigger a
regeneration prompt with confirmation; the executor never silently
ignores a stale prompt.

**What unblocks next:** a prompt file you can kick off whenever you
have time to build — or skip this step and let the executor
auto-chain.

### 2.e Prompt to implementation

Open a fresh Claude Code session and hand it the prompt.

- **Run:** `/claude-issue-executor <path-to-prompt>`
- **What you see:** the skill enters plan mode, proposes a step-by-step
  plan, and waits for your approval. Only after you approve does it
  create a feature branch and start implementing.
- **What you produce:** a branch with incremental commits (each
  referencing the ADR and issue number), tests written alongside code,
  and an evaluation summary at the end.

This is the plan-first execution model (ADR-006). The evaluation
summary is your quality gate — it lists files changed, test results,
and any manual verification steps. Do not skip reading it.

**Plan-mode rhythm (per ADR-039).** For significant sessions, the
executor pauses at the start and asks you to enter Claude Code plan
mode — the harness-level lock that gates all mutating tools — before
proposing the plan. The full rhythm:

1. The executor reports *"Significant — toggle plan mode when ready."*
2. You toggle plan mode (`shift+tab shift+tab`).
3. The executor proposes the plan inside plan mode.
4. You approve and exit plan mode (`shift+tab` once).
5. Optionally, toggle auto-accept edits (`shift+tab` again) so the
   execution phase runs without per-tool approval prompts.
6. The executor implements; if a new significant boundary appears
   mid-session it pauses and re-flags before crossing it.

Plan mode is requested automatically when the session crosses the
"significant" threshold defined in
[`skills/claude-issue-executor/SKILL.md`](../skills/claude-issue-executor/SKILL.md)
(edits a `skills/*/SKILL.md`, modifies 3+ files, edits `templates/*`,
etc.). Trivial sessions (single typo, ADR status flip, doc tweak)
skip plan mode and rely on the chat plan-gate alone.

**`--no-prompt` mode (per ADR-038).** Genuinely-trivial issues —
typo fixes, dependency bumps, doc tweaks, ADR status flips — don't
earn ADR-shaped ceremony. Pass `--no-prompt` to skip prompt
generation entirely; the executor reads the issue body directly and
leaves a one-line breadcrumb in the first commit's message
(`issue executed without prompt per ADR-038`) for the audit trail.
The criteria for when `--no-prompt` is appropriate are exactly the
**Trivial checklist** in ADR-039 (single source of truth — see
`skills/claude-issue-executor/SKILL.md`). The executor *auto-detects*
candidates conservatively: zero `ADR-NNN` references in the issue
body **and** label in `chore` / `docs` / `bugfix-trivial`. When both
hold, it suggests `--no-prompt` and asks for confirmation; explicit
`--no-prompt` overrides the auto-detect without asking. Plan-mode
rhythm still applies — `--no-prompt` skips the prompt step, not the
significance gate.

**Branch naming:** the skill creates branches following the kit's
convention (see [GitHub Flow](https://docs.github.com/en/get-started/using-github/github-flow)) —
e.g. `add-auth-middleware` or `NN-add-auth-middleware`.

**What unblocks next:** a working, tested branch ready for PR.

### 2.f Branch to PR

Package the branch into a reviewable pull request.

- **Run:** `/pr-review-packager`
- **What you see:** the skill drafts a PR body using the kit's PR
  template — summary, ADR link, file-level changelog, test results,
  `Closes #NN` trailer — and shows it to you before creating the PR.
- **What you produce:** a PR opened via `gh pr create`, with a clean
  body and the linked issue set to auto-close on merge.

For work that is not yet done, open as a draft PR (`gh pr create
--draft`) to get early feedback; convert with `gh pr ready <number>`
when implementation and tests are complete.

**What unblocks next:** a PR you can self-review and merge.

### 2.g Merge, then loop back

Review the diff on GitHub, resolve any CodeRabbit or human review
comments, then merge.

```bash
gh pr merge <number> --squash --delete-branch
git checkout main
git pull
```

- Squash merge is the default — it keeps `main`'s history clean.
- Use a regular merge only when the per-commit history of the branch
  carries real value.
- Delete the branch on merge so the repo stays tidy.

**Loop back to step 2.d** for the next issue. Keep going until the
milestone's issues are all merged.

### 2.h Close out the milestone

Before cutting a release, close the GitHub milestone the merged
issues belong to. ADR-037 ships three composable skills for this,
chained in order:

| Step | Run | Produces |
|---|---|---|
| 1. Verify the milestone is finishable | `/audit-milestone <N>` | Pass/fail report — open issues, ADRs without merged PRs, phases still `in-progress` |
| 2. Draft the retrospective | `/milestone-summary <N>` | `Design/milestones/<N>-<slug>.md` filled from `git log`, the GitHub milestone, and accepted ADRs in the date range; `lessons` zone left for you to author |
| 3. Close and (optionally) release | `/complete-milestone <N> [--release]` | GitHub milestone closed; `Design/state.md` archived (in-flight cleared, recent prepended, continue-here updated); chains `/release --milestone-phase=N` if `--release` is set |

The audit is **advisory**, not gating — gaps surface clearly but
the user has the final say. The summary file's `lessons` zone is
**user-authored** and preserved verbatim across re-runs (even with
`--overwrite`). All three skills run standalone or in any subset;
`/complete-milestone` chains the previous two as a convenience.

**One milestone == one release (default).** Per ADR-032,
`/issue-planner` creates one GitHub milestone per phase, and
`/release` defaults to one tag per phase via `--milestone-phase=N`.
Pass `--release` to `/complete-milestone` to chain `/release` after
the close, or omit it to close the milestone without cutting a tag
(e.g. when bundling multiple phases into one release).

For a worked walkthrough — including the fail case where Phase 2 has
an open issue and an ADR without a merged PR — see
[`skills/audit-milestone/example.md`](../skills/audit-milestone/example.md).

**What unblocks next:** a closed milestone with an archived state
pointer and a retrospective on disk; either a release just cut (if
`--release` chained) or a clean state to run `/release` separately.

### 2.i Tagged release

Once a milestone is closed (or as a standalone step on a single-phase
project), cut a release.

| Step | Run | Produces |
|---|---|---|
| 1. Draft release notes | `/changelog` | Markdown grouped by verb, ADR, and issue, written to stdout, `CHANGELOG.md`, or a GitHub Release body |
| 2. Tag and publish | `/release` | Semver-bumped annotated git tag, pushed to origin, GitHub Release created via `gh release create` |

`/release` calls `/changelog` internally, so you can run it directly
once you are happy with the tag and bump (patch / minor / major). It
can also mark the milestone phase complete in your build-out plan.

**What unblocks next:** a tagged, published release that stakeholders
can read and that future `/changelog` runs will pick up as the lower
bound.

### 2.j Docs sync

Before closing out the release, regenerate the docs that summarise the
current state of the project.

- **Run:** `/workflow-docs`
- **What you see:** the skill reads the PRD, MVP spec, accepted ADRs,
  and `CLAUDE.md`, then re-renders `README.md` and
  `Design/ai-summary.md` from the kit's templates.
- **What you produce:** an updated `README.md` and
  `Design/ai-summary.md` committed to `main`.

Re-run this whenever the project's shape has changed meaningfully — new
ADRs accepted, a major feature shipped, a pivot in scope. The `ai-summary.md`
in particular is what you paste into external AIs for second-opinion
design reviews.

---

## 3. Across sessions: `Design/state.md`, `/resume`, `/pause`

Long-running projects span many Claude Code sessions. Without a
canonical "where are we?" pointer, every fresh session reconstructs
state from `gh` calls and prompt files. ADR-035 closes that gap with
a single committed artefact, `Design/state.md`, plus two skills.

### The artefact

`Design/state.md` is a small (under ~100 lines), committed pointer
holding five zones, each wrapped in marker fences so individual
skills can rewrite their zone idempotently without disturbing the
others:

| Zone | Holds | Updated by |
|---|---|---|
| `phase` | Current phase name (per ADR-032) or `single` | `/pause` |
| `in-flight` | Issue number, prompt, branch, status | `/prepare-issue`, `/claude-issue-executor`, `/pr-review-packager`, `/pause` |
| `recent` | Rolling list of the last 5 merged PRs | `/pr-review-packager`, `/pause` |
| `blockers` | One line per blocker, or `none` | `/pause` |
| `continue-here` | One paragraph naming the next concrete action | `/prepare-issue`, `/pr-review-packager`, `/pause` |

The format spec is [`templates/state-template.md`](../templates/state-template.md).
The kit ships the template; each target project instantiates its
own `Design/state.md` (typically by running `/pause` for the first
time).

### Reading it: `/resume`

- **Run:** `/resume` at the start of a fresh session.
- **What you see:** a one-message brief — phase, in-flight issue,
  last five PRs, blockers, "continue here" — with no `gh` calls on
  the happy path.
- **Fallback:** if `Design/state.md` is missing, empty, or looks
  suspect (in-flight issue already merged, file mtime stale),
  `/resume` falls back to `gh pr list` / `gh issue list` and
  recommends running `/pause` to seed the file.

### Refreshing it: `/pause`

- **Run:** `/pause` (or `/pause --handoff`) before a context reset,
  end of day, after a non-trivial detour, or to seed the file the
  first time you adopt ADR-035.
- **What you see:** each zone is recomputed from current truth
  (build-out plan for phase, prompt + branch for in-flight, `gh pr
  list --state merged` for recent, you for blockers and continue-
  here) and written back. With `--handoff`, an additional richer
  `notes/handoff-YYYY-MM-DD.md` is written for context-window-
  exhausting handoffs.
- **What you produce:** a refreshed `Design/state.md` ready to
  commit.

### Conflict-resolution rule

`Design/state.md` is committed, so parallel branches can produce
merge conflicts in it. ADR-035's rule: **the most recently merged
PR's version wins for the conflicting zone.** If the result looks
suspect after the merge, `/resume` re-derives from `gh` and points
you at `/pause` to refresh.

### Optional CI line-cap check

`bin/check-state-cap` exits 1 if `Design/state.md` exceeds the
line cap (default 100). Wire it into CI alongside `bin/sync-adr-index`
if you want a guard rail; it exits 0 silently when the file is
absent, so adopting it is no-op for projects that have not yet
turned on ADR-035.

---

## 4. When you don't need an ADR

Not every change is an architectural decision. Bug fixes, chores,
dependency bumps, CI/CD tweaks, and doc-only changes generally do not
warrant an ADR — forcing one creates ceremony without insight.

**ADR-free changes still follow the same GitHub flow:**

| Stage | ADR-driven work | ADR-free work |
|---|---|---|
| Decision | Write ADR, accept | Skip |
| Issue | `/issue-planner` or manually via `gh issue create` | Manually via `gh issue create` |
| Branch | Feature branch from `main` | Feature branch from `main` |
| Prompt | `/prepare-issue <N>` | `/prepare-issue <N>` (still fills the template; ADR section is marked "none") |
| Build | `/claude-issue-executor` | `/claude-issue-executor` |
| PR | `/pr-review-packager` | `/pr-review-packager` |
| Merge | `gh pr merge --squash` | `gh pr merge --squash` |

So: still need an issue. Still need a branch. Still need a PR. You just
skip `/adr-writer` and mark the ADR section of the issue as "none".

The kit's issue template explicitly supports this — see the HTML
comment under the `## ADR` heading in
[`templates/issue-template.md`](../templates/issue-template.md): *"If
no ADR applies, write 'none' and say why."*

**Examples of ADR-free work:**

| Change | Why no ADR |
|---|---|
| Bug fix in an existing module | Restoring intended behaviour, not deciding new architecture |
| Dependency version bump | Unless it forces a behaviour change, it's a chore |
| CI/CD tweak (add a lint step, change a runner) | Operational, not architectural |
| README typo fix, docs rephrasing | Doc-only |
| Refactor with no behaviour change | Unless it reshapes the architecture, it's cleanup |
| Config tweak (log level, timeouts) | Tuning, not deciding |

If you are unsure, err on the side of writing an ADR — they are cheap,
and a rejected ADR is still useful as a record of what was considered.
See `generic-project-workflow.md` §4.2 for the full "when to write an
ADR" table.

---

## 5. Where to go deeper

| You want to… | Go to |
|---|---|
| See the full reference workflow with every option and edge case | [`generic-project-workflow.md`](../generic-project-workflow.md) |
| Learn how to run Claude Code skills from inside a target project | `docs/claude-code-guide.md` (coming in a later issue) |
| Understand a specific skill's interface, inputs, and outputs | `.claude/skills/<name>/SKILL.md` in your target project, or [`skills/README.md`](../skills/README.md) in the kit |
| Revisit a design decision that shaped this workflow | [`Design/adr/`](../Design/adr/) |
| Set up GitHub labels, milestones, and branch protection | [`docs/github-setup.md`](github-setup.md) |
| Fill a Claude Code session prompt by hand | [`docs/issue-prompt-guide.md`](issue-prompt-guide.md) |

If the flow above doesn't match the skills you have installed, check
[`skills/README.md`](../skills/README.md) for the current inventory —
some skills land in later milestones and the guide lists the full v-next
set.
