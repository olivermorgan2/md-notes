# pr-review-packager — worked example

A single end-to-end walkthrough showing how `pr-review-packager`
turns a feature branch into a pull request. This is reference
material, not a runtime artifact.

---

## 1. Setup

The user has just finished implementing Issue #17 on a feature
branch:

```
Branch:   issue-17-pr-review-packager
Base:     main
Upstream: origin/issue-17-pr-review-packager (pushed)
Commits ahead of main:
  abc1234  feat(skills): add pr-review-packager skill (ADR-015, #17)
```

Working tree is clean. The user runs:

```
/pr-review-packager
```

No flags.

## 2. Data the skill gathers

**From git:**

- `git symbolic-ref --short HEAD` → `issue-17-pr-review-packager`
- `git rev-parse --abbrev-ref --symbolic-full-name @{u}` → `origin/issue-17-pr-review-packager`
- `git status --porcelain` → empty
- `git log main..HEAD --format="%H%x09%s"` →
  ```
  abc1234	feat(skills): add pr-review-packager skill (ADR-015, #17)
  ```

**From the filesystem:**

- `templates/pr-template.md` exists and is read.
- `Design/adr/adr-015-pr-review-packager-skill.md` exists (matched by
  glob `Design/adr/adr-0*15-*.md`).

## 3. Derived values

- **Issue number:** `17`, extracted from the branch name
  (`issue-17-pr-review-packager` → `17`). The branch-name source hit
  first; the skill did not need to fall through to commit messages
  or prompt files.
- **ADR references:** `ADR-015`, extracted from the commit subject's
  `(ADR-015, #17)` tail. Resolved to
  `Design/adr/adr-015-pr-review-packager-skill.md`.
- **Change summary groups:** one commit, prefix `feat`. Because there
  is only one group, the skill omits the group heading and emits a
  single bullet:
  ```
  - add pr-review-packager skill
  ```
- **PR title:** `feat(skills): add pr-review-packager skill`
  (derived from the newest commit subject, with the trailing
  `(ADR-015, #17)` stripped).
- **Summary draft:** `Adds the pr-review-packager skill that
  packages feature branches into well-structured PRs (closes #17).`

## 4. Rendered body (shown to user for approval)

The skill prints the following to chat and asks
`"Open PR with this title and body? (yes / edit / cancel)"`.

> **Title:** `feat(skills): add pr-review-packager skill`

````markdown
## Summary

Adds the pr-review-packager skill that packages feature branches into well-structured PRs (closes #17).

## Closes

Closes #17

## ADR

Related ADR: Design/adr/adr-015-pr-review-packager-skill.md

## Changes

- add pr-review-packager skill

## Test results

```
<!-- TODO: paste test-runner output or delete the code fence and write "no code changes — docs only" -->
```

## Manual verification

<!-- TODO: list verification steps, or write "none needed" -->
````

The skill flags the two TODO comments so the user can fill them
during the `edit` loop if desired.

## 5. User confirmation

The user replies `edit` and types in test results and verification
steps. The skill re-renders the body with the user's text and asks
again. This time the user replies `yes`.

## 6. PR creation

The skill runs:

```
gh pr create \
  --title "feat(skills): add pr-review-packager skill" \
  --body  "<rendered body>" \
  --base  main \
  --head  issue-17-pr-review-packager
```

`gh` returns:

```
https://github.com/olivermorgan2/workflow-generator/pull/42
```

The skill reports:

> Opened PR #42:
> `https://github.com/olivermorgan2/workflow-generator/pull/42`
> (Closes #17, ADR-015; no TODOs remaining.)

## 7. What would change for other scenarios

### Multiple commits, mixed types

Branch `issue-22-workflow-guide` with three commits:

```
feat(docs): add workflow guide skeleton (ADR-020, #22)
fix(docs): fix heading levels (#22)
docs: add cross-links to ADRs (#22)
```

The change summary becomes:

```
**Features**
- add workflow guide skeleton

**Fixes**
- fix heading levels

**Docs**
- add cross-links to ADRs
```

### Branch name does not encode the issue

Branch `add-auth-middleware` with one commit
`feat(auth): add middleware (ADR-030, #31)`.

- Step 6.1 (branch-name regex) misses — no `issue-N` pattern, no
  trailing `-N`, no `#N`.
- Step 6.2 (commit subjects) matches `#31`.
- `Closes #31` is filled, no TODO needed.

### No issue number anywhere

Branch `chore-tidy-notes`, commit `chore: tidy notes`.

- Step 6.1, 6.2, 6.3 all miss.
- The skill emits
  `Closes #<!-- TODO: fill in issue number -->`
  and flags it at the approval gate. The user fills it during the
  `edit` loop.

### No ADR referenced

Same branch as above — no `ADR-NNN` tokens anywhere.

- The `## ADR` section becomes `Related ADR: none`.

### Detached HEAD

The user ran `/pr-review-packager` while rebasing.

- `git symbolic-ref --short HEAD` fails.
- The skill aborts with
  `"HEAD is detached. Check out a branch before running /pr-review-packager."`

### Branch not pushed

User forgot `git push -u origin <branch>`.

- `git rev-parse --abbrev-ref --symbolic-full-name @{u}` errors.
- The skill aborts with
  `"No upstream for <branch>. Run: git push -u origin <branch>"`.

### Optional flags forwarded

```
/pr-review-packager --draft --label feature --reviewer @alice
```

The skill passes `--draft --label feature --reviewer @alice` through
to `gh pr create` in step 14. Everything else is identical.
