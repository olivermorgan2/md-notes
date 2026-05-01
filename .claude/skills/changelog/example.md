# changelog — worked example

A run of `/changelog` against this kit's own recent history. The
input is the range from commit `dc2869d` (the "Accept ADRs 007-021"
commit on `main`) to `HEAD`. This is the same sort of range the
skill would see if a release tag `v0.1.0` had been cut at `dc2869d`
and the user ran `/changelog --since-last-release`.

---

## 1. Invocation

```
/changelog --from=dc2869d --to=HEAD
```

Equivalent if `dc2869d` were tagged `v0.1.0`:

```
/changelog --since-last-release
```

## 2. Raw git log in the range

With default `--no-merges`:

```
bdfb349 feat(bin): add install-workflow-kit script (ADR-009, #13)
cd8fd3b feat(skills): add /prepare-issue skill (ADR-013, #15)
f06f3cc feat(templates): expand CLAUDE.md starter template (ADR-007, #11)
43dcf65 feat(prompts): add dedicated prompts/ folder and template (ADR-008, #12)
c344cfa Update feature-ideas.md statuses to reflect accepted ADRs (#24)
```

Four of the five commits match the conventional `feat(...)` prefix
and land in **Features**. The last one (`Update feature-ideas.md ...`)
has no verb prefix and lands in **Other**. The two `Merge #N:` commits
that also sit in this range are dropped by `--no-merges`.

## 3. Parsing trace

For `cd8fd3b feat(skills): add /prepare-issue skill (ADR-013, #15)`:

- Subject regex match → verb = `feat`, scope = `skills`.
- Stripped subject: `add /prepare-issue skill (ADR-013, #15)`.
- ADR tokens: `ADR-013`.
- Issue tokens: `#15`.
- Trailing `(#15)` recognised as duplicate of the linked issue →
  stripped from the subject during render.
- Final rendered subject: `add /prepare-issue skill`.

For `c344cfa Update feature-ideas.md statuses to reflect accepted ADRs (#24)`:

- Subject regex does not match a known verb → section = Other.
- ADR tokens: none (literal `ADRs` is not `ADR-<digits>`).
- Issue tokens: `#24`.
- Trailing `(#24)` stripped from rendered subject.

## 4. Output — stdout

```markdown
# Changelog dc2869d..HEAD

## Features

- add install-workflow-kit script ([bdfb349](https://github.com/olivermorgan2/workflow-generator/commit/bdfb349...), [#13](https://github.com/olivermorgan2/workflow-generator/issues/13), ADR-009)
- add /prepare-issue skill ([cd8fd3b](https://github.com/olivermorgan2/workflow-generator/commit/cd8fd3b...), [#15](https://github.com/olivermorgan2/workflow-generator/issues/15), ADR-013)
- expand CLAUDE.md starter template ([f06f3cc](https://github.com/olivermorgan2/workflow-generator/commit/f06f3cc...), [#11](https://github.com/olivermorgan2/workflow-generator/issues/11), ADR-007)
- add dedicated prompts/ folder and template ([43dcf65](https://github.com/olivermorgan2/workflow-generator/commit/43dcf65...), [#12](https://github.com/olivermorgan2/workflow-generator/issues/12), ADR-008)

## Other

- Update feature-ideas.md statuses to reflect accepted ADRs ([c344cfa](https://github.com/olivermorgan2/workflow-generator/commit/c344cfa...), [#24](https://github.com/olivermorgan2/workflow-generator/issues/24))
```

(Full SHAs truncated with `...` in this example for readability;
the real output uses the full 40-character SHA in the URL.)

Followed by a one-line summary on stderr:

```
5 entries in 2 sections (Features: 4, Other: 1) — stdout
```

## 5. Same range with `--include-merges`

```
/changelog --from=dc2869d --to=HEAD --include-merges
```

Adds two more entries under **Other** because `Merge #N:` subjects
do not match the verb map:

```markdown
## Other

- Merge #13: install-workflow-kit script (ADR-009) ([dbe4562](...))
- Merge #15: /prepare-issue skill (ADR-013) ([779ec45](...))
- Update feature-ideas.md statuses to reflect accepted ADRs ([c344cfa](...), [#24](...))
```

In a squash-merge workflow this is usually noise — the default
`--no-merges` behaviour is what you want.

## 6. File target

```
/changelog --from=dc2869d --to=HEAD --output=CHANGELOG.md
```

Writes the exact markdown above to `CHANGELOG.md`, overwriting if
present. Stdout gets:

```
Wrote 5 entries to CHANGELOG.md
```

The user is expected to move that content under a `## [0.1.0] — 2026-04-17`
heading in an existing `CHANGELOG.md` by hand — this skill does not
do merge-into-existing-changelog. That is intentional: editing a
curated changelog file is a review act.

## 7. GitHub Release target

```
/changelog --since-last-release --github-release=v0.2.0
```

Assuming the tag `v0.2.0` exists and the release has been created
(by `/release` or by hand), the skill runs:

```
gh release edit v0.2.0 --notes-file -
```

with the rendered markdown on stdin. On success it prints the
release URL.

If the release does not exist yet, the skill stops and asks:

> Release v0.2.0 does not exist. Create it now via
> `gh release create v0.2.0 --notes-file -`? (yes/no)

Default no. The skill never creates a tag — that is `/release`'s job.

## 8. Self-check trace

- [x] Exactly one ref-selection mode (`--from/--to`).
- [x] Both refs resolved via `git rev-parse`.
- [x] Title line filled (`dc2869d..HEAD`).
- [x] Every entry has a linked short SHA.
- [x] Every `#N` is a linked issue.
- [x] Trailing `(#N)` stripped from subjects where the issue is
      already a linked ref.
- [x] Section order is Features → Other. Empty sections omitted.
- [x] No duplicates within a section.

All checks pass. The skill reports back:

> Rendered 5 entries across 2 sections for range `dc2869d..HEAD`.
> Output written to stdout.
