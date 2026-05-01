# release — worked example

A typical second-release run for the *Pace Drift* project. The last
release was `v0.2.0`; since then one new ADR (ADR-007: rolling-window
smoothing) was accepted, two features were merged, and one bug was
fixed. No breaking changes. The user wants to publish `v0.3.0`.

---

## 1. Invocation

```
/release
```

No flags — the user wants the skill to suggest the version.

## 2. Prerequisites check

The skill runs, silently in the background:

- `gh auth status` — OK.
- `git status --porcelain` — empty, clean tree.
- `git rev-parse --abbrev-ref HEAD` — `main`.
- `git fetch origin main && git rev-list --left-right --count HEAD...origin/main` — `0 0`.
- `.claude/skills/changelog/SKILL.md` — exists.

All pass.

## 3. Last-tag detection

```
git describe --tags --abbrev=0
```

Returns `v0.2.0`.

## 4. Suggested-bump rationale

The skill scans commits and ADR diffs in `v0.2.0..HEAD`:

- Commits: 2 `feat(` + 1 `fix(` + 3 `chore(`.
- ADRs: `Design/adr/adr-007-rolling-window-smoothing.md` is new,
  status `accepted`.
- No `BREAKING CHANGE` markers, no `breaking` PR label, no ADR
  status change from `accepted` to `superseded`.

Decision tree:

- Major? No breaking signals.
- Minor? Yes — new accepted ADR **and** new `feat(` commits.
- Patch? N/A, minor wins.

Suggestion: **minor** → `v0.2.0` → `v0.3.0`.

## 5. `/changelog` invocation

```
/changelog --since-last-release --output=- --github-release=v0.3.0
```

Returns (abbreviated):

```markdown
## v0.3.0

### Added
- Rolling-window smoothing for pace series (ADR-007, #42)
- CSV export from the drift chart (#45)

### Fixed
- GPX parser tolerates missing elevation nodes (#47)

### Other
- Bump dev dependencies (#46, #48)
```

## 6. Plan presented to the user

```
Release plan

  Last tag:        v0.2.0
  Next version:    v0.3.0  (minor bump — new ADR + feat commits)
  Branch:          main (in sync with origin)
  Draft:           no
  Prerelease:      no

Release notes preview:

  ## v0.3.0
  ### Added
  - Rolling-window smoothing for pace series (ADR-007, #42)
  - CSV export from the drift chart (#45)
  ### Fixed
  - GPX parser tolerates missing elevation nodes (#47)
  ### Other
  - Bump dev dependencies (#46, #48)

Annotated tag message:

  Release 0.3.0

  Rolling-window smoothing for pace series (ADR-007, #42)

Commands to run on approval:

  /changelog --since-last-release --output=<tmp> --github-release=v0.3.0
  git tag -a v0.3.0 -m "Release 0.3.0\n\nRolling-window smoothing for pace series (ADR-007, #42)"
  git push origin v0.3.0
  gh release create v0.3.0 --title "v0.3.0" --notes-file <tmp>

Type `yes` to create and push the tag and publish the release.
Any other response cancels.
```

## 7. Approval

User types `yes`.

## 8. Execution

Commands run in order. `gh release create` returns:

```
https://github.com/pacedrift/pacedrift/releases/tag/v0.3.0
```

## 9. Report

```
Released v0.3.0.

  Tag:     v0.3.0 (pushed to origin)
  Release: https://github.com/pacedrift/pacedrift/releases/tag/v0.3.0
  Branch:  main (unchanged)
```

---

## Dry-run variant

If the user had run `/release --dry-run` instead, the output stops at
step 6 and the approval prompt is replaced with:

```
Dry-run. Would execute:

  git tag -a v0.3.0 -m "Release 0.3.0\n\nRolling-window smoothing for pace series (ADR-007, #42)"
  git push origin v0.3.0
  gh release create v0.3.0 --title "v0.3.0" --notes-file <tmp>

No changes made.
```

---

## Edge-case variant — tag already exists

If the user had passed `--version=0.2.0` (the existing tag), the skill
stops at step 4:

```
Tag v0.2.0 already exists. Releases do not modify existing tags.
Suggested next version: v0.2.1 (patch). Re-run with --version=0.2.1
or --bump=patch if that's what you want.
```

No further steps run.
