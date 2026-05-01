---
name: pr-review-packager
description: Draft a pull-request body from templates/pr-template.md, auto-fill Closes #N and ADR references from branch and commit history, show the draft for approval, then open the PR via gh pr create
---

# pr-review-packager

Package a feature branch into a well-structured pull request. The skill
reads [`templates/pr-template.md`](../../templates/pr-template.md),
fills in `Closes #N` and ADR references from the branch name and commit
history, derives a change summary from `git log main..HEAD`, shows the
finished body in chat for approval, and only then calls
`gh pr create`.

This skill completes the implementation pipeline:
`/prepare-issue` → `claude-issue-executor` → **`/pr-review-packager`**.
See [ADR-015](../../Design/adr/adr-015-pr-review-packager-skill.md) for
the decision record.

## When to use this skill

Use when an implementation branch is ready for review and the user
wants to open the PR without hand-filling the template. Typical
invocation:

```
/pr-review-packager
```

Optional pass-through flags are supported and forwarded verbatim to
`gh pr create`:

```
/pr-review-packager --label feature --milestone "MVP" --reviewer @alice
```

The skill is **read-only** with respect to `templates/pr-template.md`
(ADR-005) — it consumes the file without editing it.

## What this skill does not do

- Does not edit `templates/pr-template.md`. Ever. It is a read-only
  consumer.
- Does not modify the local branch or rewrite commits. If the commit
  history needs cleaning up, the user does that before invoking the
  skill.
- Does not push the branch. Pushing is the user's responsibility; the
  skill fails cleanly if the upstream branch does not exist yet and
  reports the exact `git push -u origin <branch>` command to run.
- Does not create a draft PR by default. If the user wants one they
  can pass `--draft`.
- Does not review the PR content for correctness (that is `/review`,
  a built-in). It only packages the body.
- Does not run CI, tests, or linters. It only reports what `git log`
  shows.
- Does not open the PR without explicit user approval of the rendered
  body. The approval gate from ADR-015's chosen option (Option B) is
  mandatory.

## Inputs

- **Required:** none — the skill reads everything it needs from the
  current git worktree, the repo, and `gh`.
- **Optional flags** (passed through to `gh pr create`):
  - `--label <name>` (repeatable)
  - `--milestone <name>`
  - `--reviewer <user-or-team>` (repeatable)
  - `--draft`
  - `--base <branch>` (defaults to `main`)
  - `--assignee <user>` (repeatable)

Flags the skill does **not** forward: anything that would suppress the
body (`--fill`, `--body-file`), anything that would bypass the approval
gate, and anything destructive.

## Output

- A new pull request on GitHub, opened via
  `gh pr create --title "..." --body "..." --base <base> --head <current-branch>`
  plus any pass-through flags.
- A one-line confirmation in chat including the PR URL returned by
  `gh`.

## Data sources and how the skill reads them

The skill consults four sources, in this order:

1. **Current branch and git history.**
   - `git symbolic-ref --short HEAD` → current branch name. A detached
     HEAD aborts the skill (see Edge cases).
   - `git rev-parse --abbrev-ref --symbolic-full-name @{u}` → upstream
     ref, used to confirm the branch has been pushed. If no upstream
     is set, the skill stops with a clear message telling the user to
     run `git push -u origin <branch>` first.
   - `git log <base>..HEAD --format="%H%x09%s"` → commits ahead of
     base, used for the change summary and as a secondary source for
     issue and ADR tokens. `<base>` is `main` unless the user passed
     `--base`.
2. **PR template.** Read `templates/pr-template.md` verbatim. If it is
   missing, abort (see Edge cases).
3. **Issue prompt file** (fallback). Glob `prompts/issue-NNN-*.md` and
   pick the most recently modified file. Useful when the branch name
   does not encode the issue number.
4. **ADR files.** For each `ADR-NNN` token found, glob
   `Design/adr/adr-<NNN>-*.md` and resolve the full filename.

`gh` is the only allowed tool for creating the PR. Do not call the
REST/GraphQL API directly.

## Execution protocol

Run these steps in order. Stop on the first failure unless otherwise
noted.

1. **Verify working tree.** Run `git status --porcelain`. If it is
   non-empty, warn the user and ask whether to continue — the
   uncommitted changes will not be included in the PR.
2. **Resolve the branch.** Run `git symbolic-ref --short HEAD`. On a
   detached HEAD, abort with:
   `"HEAD is detached. Check out a branch before running /pr-review-packager."`
3. **Confirm base.** The base branch is `main` unless `--base` was
   passed. If the current branch equals the base, abort with:
   `"Current branch is <base>. Check out a feature branch before packaging a PR."`
4. **Confirm the branch has been pushed.** Run
   `git rev-parse --abbrev-ref --symbolic-full-name @{u}`. If it
   errors, abort with:
   `"No upstream for <branch>. Run: git push -u origin <branch>"`.
5. **List commits ahead of base.** Run
   `git log <base>..HEAD --format="%H%x09%s"`. If the output is empty,
   abort with:
   `"No commits on <branch> ahead of <base>. Nothing to package."`
6. **Extract the issue number.** In priority order, use the first hit:
   1. **Branch name.** Match the pattern
      `issue-(\d+)-` or a trailing `-(\d+)$` against the current
      branch name, or an explicit `#(\d+)` token.
   2. **Commit messages.** Scan the commit subjects in
      `git log <base>..HEAD` for `#(\d+)` tokens (typically the
      trailing `(ADR-NNN, #N)` convention this repo uses). Use the
      first unique number seen, most-recent commit first.
   3. **Most recent issue prompt.** Glob `prompts/issue-(\d+)-*.md`,
      pick the file with the newest mtime, and read `NNN` from its
      filename. Strip leading zeros for the PR body.
   If no hit is found, leave `{{ISSUE_NUMBER}}` as
   `<!-- TODO: fill in issue number -->` and flag it during the
   approval gate rather than blocking.
7. **Extract ADR references.** Scan the same sources (branch name,
   commit subjects, most recent prompt file) for `ADR-(\d+)` tokens
   (case-insensitive). Deduplicate, preserve first-seen order. For
   each unique `NNN`, glob `Design/adr/adr-<NNN>-*.md`. If a file
   matches, record the path. If none match, keep the `ADR-NNN` token
   and flag it as a TODO in the approval gate. If no ADR tokens are
   found anywhere, set the `## ADR` body to `none`.
8. **Derive the change summary.** From the commit subjects, group by
   conventional-commit verb prefix: `feat`, `fix`, `docs`, `refactor`,
   `chore`, `test`, `perf`, `ci`, `build`, `style`. Any commit without
   a recognised prefix goes under a final `other` group. Within each
   group, keep one bullet per commit, preserving commit order
   (oldest-first) and stripping the trailing `(ADR-NNN, #N)` suffix
   for readability. If there is only one group, omit the group
   heading and just list the bullets.
9. **Derive the PR title.** Use the newest commit subject on the
   branch, stripped of the trailing `(ADR-NNN, #N)` suffix. If the
   title still ends with something like `(#NN)`, strip that too. The
   user can override during approval.
10. **Draft the Summary paragraph.** One sentence: take the verb and
    object from the PR title, plus the issue number if known. Example:
    `"Adds the pr-review-packager skill that packages feature
    branches into well-structured PRs (closes #17)."` This is a
    best-effort draft — the user will edit if needed.
11. **Fill the template.** Read `templates/pr-template.md` and
    substitute each placeholder:

    | Placeholder | Filled with |
    |---|---|
    | `{{One-paragraph description of what this PR changes and why.}}` | the Summary draft from step 10 |
    | `{{ISSUE_NUMBER}}` | issue number from step 6 (no `#`) |
    | `` `Design/adr/adr-{{NNN}}-{{short-title}}.md` `` line | one resolved ADR path per line, each as `Related ADR: Design/adr/adr-NNN-short-title.md`. If no ADRs, the whole line becomes `Related ADR: none`. |
    | `{{Bullet the substantive changes, not every file touched.}}` block | the grouped bullets from step 8 |
    | `{{Paste test-runner output: total / passed / failed / skipped.}}` | left as a placeholder `<!-- TODO: paste test-runner output or delete the code fence and write "no code changes — docs only" -->` |
    | `{{Steps a reviewer should run to convince themselves the change works.}}` | `<!-- TODO: list verification steps, or write "none needed" -->` |

    Strip the template's HTML comments that instruct the author (they
    are not part of the rendered body the skill produces).

12. **Show the filled body to the user** as a fenced markdown block
    plus the derived title on a line above it. Flag any TODO markers
    and any unresolved tokens (e.g. ADR file missing, no issue
    number). Ask explicitly:
    `"Open PR with this title and body? (yes / edit / cancel)"`.
    **Do not call `gh pr create` before this confirmation.**
13. **Handle the response.**
    - `yes` → proceed to step 14.
    - `edit` → ask the user what to change, apply it, re-render, re-ask.
      Loop until `yes` or `cancel`.
    - `cancel` → stop without calling `gh`.
14. **Open the PR.** Call:
    ```
    gh pr create \
      --title "<title>" \
      --body  "<body>" \
      --base  "<base>" \
      --head  "<current-branch>" \
      [pass-through flags]
    ```
    Pass-through flags are the `--label`, `--milestone`, `--reviewer`,
    `--assignee`, `--draft` values the user provided at invocation.
15. **Update `Design/state.md` if present.** Per
    [ADR-035](../../Design/adr/adr-035-state-md-session-continuity.md),
    close out the issue in the state file:
    - prepend a one-line entry to the `state:recent` zone of the
      form `#<PR> — ADR-NNN — <commit-summary first sentence>`
      (use `none` for the ADR token if no ADR was resolved). If the
      zone now has more than five entries, drop the oldest so the
      rolling list stays at five.
    - rewrite the `state:in-flight` zone to `Issue: none`,
      `Prompt: n/a`, `Branch: n/a`, `Status: none`.
    - rewrite the `state:continue-here` zone to one short paragraph
      pointing at the just-opened PR, e.g. `"Review and merge #<PR>;
      then pick the next issue from the queue."`.
    Marker fences bound each zone; rewrite only the bytes between
    the fences. If `Design/state.md` is absent, skip silently. If
    marker fences are broken, do not rewrite; surface the broken
    zone in the final report and suggest `/pause`.
16. **Report.** Print the PR URL returned by `gh`, plus a one-line
    summary of what was filled vs. what was left as TODO. Done.

## Issue-link extraction — detailed rules

Priority order (first hit wins):

1. **Current branch name.** Regexes, tried in order:
   - `^issue-0*(\d+)-` (e.g. `issue-17-pr-review-packager` → `17`)
   - `-#?(\d+)$` (e.g. `add-auth-17` → `17`)
   - `#(\d+)` anywhere in the branch name.
2. **Commit subjects from `git log <base>..HEAD`.** Walk
   newest-first. For each subject, extract all `#(\d+)` tokens.
   Accept the first one that is not also preceded by `ADR-`
   (guards against false positives from things like
   `ADR-015#section`). Typical hits come from the repo's
   `feat(skills): add X (ADR-NNN, #17)` convention.
3. **Most recent issue prompt file.** Glob `prompts/issue-0*(\d+)-*.md`,
   sort by mtime descending, take the top match, parse the leading
   digits.

Strip leading zeros before writing into the PR body — GitHub's
`Closes #17` does not want `#017`.

## ADR-link extraction — detailed rules

Sources, in the same priority order as issue extraction: branch name,
then commit subjects, then the most recent prompt file.

Parse tokens with the regex `ADR-0*(\d+)` case-insensitive. Preserve
first-seen order, deduplicate.

For each unique `NNN`, glob `Design/adr/adr-0*<NNN>-*.md`. If exactly
one file matches, use it. If multiple match (should never happen,
ADR numbers are unique), use the shortest filename. If none match,
keep the `ADR-NNN` token in the body and append
`<!-- TODO: ADR file not found in Design/adr/ -->`.

If multiple ADRs are found, emit one `Related ADR:` line each, in the
order they were discovered:

```
Related ADR: Design/adr/adr-015-pr-review-packager-skill.md
Related ADR: Design/adr/adr-005-template-architecture.md
```

If zero ADRs are found anywhere, emit the single line
`Related ADR: none`.

## Change-summary derivation

Given commit subjects from `git log <base>..HEAD --format="%s"`
(oldest-first):

1. For each subject, identify the conventional-commit verb prefix by
   matching `^(feat|fix|docs|refactor|chore|test|perf|ci|build|style)(\([^)]*\))?:\s*`.
   Everything after the `:` is the bullet text. If no prefix matches,
   the whole subject is the bullet text and the group is `other`.
2. Strip a trailing ` (ADR-NNN, #N)` or ` (#N)` suffix from each
   bullet.
3. Group bullets by prefix. Group order in the output:
   `feat`, `fix`, `refactor`, `perf`, `docs`, `test`, `ci`, `build`,
   `chore`, `style`, `other`.
4. If only one group has any bullets, omit the group heading and emit
   just the bullets.
5. If multiple groups have bullets, emit each as a short subheading
   in bold (`**Features**`, `**Fixes**`, `**Docs**`, etc.) followed
   by the bullets.

One bullet per commit. Do not attempt to merge related commits — the
user can edit during the approval gate.

## Edge cases

- **Detached HEAD** → abort with
  `"HEAD is detached. Check out a branch before running /pr-review-packager."`.
- **Current branch is the base branch** → abort with
  `"Current branch is <base>. Check out a feature branch before packaging a PR."`.
- **No upstream for the branch** → abort with
  `"No upstream for <branch>. Run: git push -u origin <branch>"`.
- **No commits ahead of base** → abort with
  `"No commits on <branch> ahead of <base>. Nothing to package."`.
- **`templates/pr-template.md` missing** → abort with
  `"templates/pr-template.md not found. The kit's PR template must exist before this skill can run."`.
- **Uncommitted changes in the working tree** → warn and ask to
  continue; do not abort automatically.
- **No issue number anywhere** → proceed, but fill
  `Closes #<!-- TODO: fill in issue number -->` and flag it at the
  approval gate. Do not guess.
- **ADR token found but no file** → keep the token in the body and
  append `<!-- TODO: ADR file not found -->`. Flag at the approval
  gate.
- **Multiple issue numbers on the branch** → emit one `Closes #N`
  line per unique number, in first-seen order. GitHub accepts
  multiple.
- **`gh pr create` exits non-zero** (e.g. duplicate PR, auth error) →
  print `gh`'s stderr verbatim and stop. Do not retry.
- **An existing PR for the branch** — `gh pr create` will report this.
  Surface the existing PR's URL from `gh pr view --json url` so the
  user can decide whether to update it manually.

## Review-before-create checkpoint

The user always sees the rendered title and body in chat before
`gh pr create` is called. This preserves the approval gate from
ADR-015's chosen option. Never skip this step — not even when the
draft looks clean, not even with a `--yes` flag (the skill does not
accept one).

## Self-check before calling `gh`

- [ ] Current branch is not the base branch and is not detached.
- [ ] Upstream exists for the current branch.
- [ ] At least one commit is ahead of the base branch.
- [ ] `templates/pr-template.md` was read successfully.
- [ ] Every placeholder in the rendered body is either filled or
      explicitly marked with a `<!-- TODO: ... -->` comment.
- [ ] Every referenced ADR either resolves to a file in
      `Design/adr/` or is explicitly marked TODO.
- [ ] **ADR index is in sync.** Run `bin/sync-adr-index --check`. If
      it reports drift (exit 1), the branch added or modified an
      ADR without updating `Design/adr/README.md`. Run
      `bin/sync-adr-index`, commit the index update on the same
      branch, and re-show before opening the PR. (ADR-023.)
- [ ] The user explicitly confirmed the body with `yes`.

If any fail, fix and re-show before calling `gh`.

## Relationship to other skills and commands

- **`/review`** (built-in): reviews an existing PR. This skill
  creates one. They complement each other — package first, review
  after.
- **`claude-issue-executor`** (ADR-014, Issue #16): runs the
  implementation session for an issue. When the implementation is
  done and committed, the user (or the executor's handoff step)
  invokes `/pr-review-packager` to open the PR.
- **`/changelog`** (ADR-016, Issue #18): complementary — consumes
  merged PRs to update `CHANGELOG.md`. This skill is upstream of it.
- **`/release`** (ADR-017, Issue #19): downstream — cuts a release
  from merged PRs. This skill is also upstream of it.

## Handoff

The PR is the handoff. Once opened, the review and merge are
human-driven (or handled by `/review`). This skill is done when
`gh pr create` returns a URL and the URL is reported to the user.

See [`example.md`](example.md) for a worked invocation on this very
issue.
