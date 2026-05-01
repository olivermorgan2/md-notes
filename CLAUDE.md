<!--
  Template: CLAUDE.md — project rules for Claude Code
  Filled by:  the workflow-docs skill (later), or a human at project bootstrap
  Output in a target project: CLAUDE.md at the repo root
  Source:     templates/claude-md-template.md in the Claude Code Workflow Kit

  CLAUDE.md is the single most important file for AI-assisted work in a
  project. Claude Code reads it on every session. Keep it current — stale
  rules create worse output than no rules.

  ------------------------------------------------------------------------
  PLACEHOLDERS (all use {{UPPER_SNAKE}} syntax; grep them with:
      grep -E '\{\{[A-Z_]+\}\}' CLAUDE.md
  ------------------------------------------------------------------------

  Required — the template is not useful until these are filled:
    md-notes          Human-readable project name. Example: "Acme Notes"
    {{PROJECT_TAGLINE}}       One-line elevator pitch. Example: "A local-first
                              Markdown note-taking app for researchers."
    {{PRIMARY_LANGUAGE}}      Main runtime / language. Example: "Node.js 22 + TypeScript"
    {{FRAMEWORK}}             Primary framework, or "none". Example: "Next.js 15"
    {{DATA_LAYER}}            Database / storage, or "none yet". Example: "Postgres 16"
    {{KEY_LIBRARIES}}         Comma-separated top 3–5 libraries a reviewer
                              must know. Example: "Prisma, tRPC, Tailwind"
    {{PACKAGE_MANAGER}}       Example: "pnpm", "npm", "uv", "cargo"
    {{INSTALL_COMMAND}}       Example: "pnpm install"
    {{DEV_COMMAND}}           Example: "pnpm dev"
    {{BUILD_COMMAND}}         Example: "pnpm build"
    {{TEST_COMMAND}}          Example: "pnpm test"
    {{LINT_COMMAND}}          Example: "pnpm lint", or "none" if you haven't
                              picked a linter yet
    {{TEST_FRAMEWORK}}        Example: "vitest", "pytest", "go test"
    {{TEST_LOCATION}}         Example: "test/ mirroring src/"
    {{FORMATTER}}             Example: "prettier", "black", "rustfmt", "none"
    {{MODULE_SYSTEM}}         Example: "ESM", "CommonJS", "N/A"
    {{GITHUB_OWNER}}          GitHub user or org. Example: "acme-labs"
    {{GITHUB_REPO}}           Repo slug. Example: "acme-notes"
    {{DEFAULT_BRANCH}}        Example: "main"
    {{CURRENT_MILESTONE}}     Example: "Foundation", "MVP", "v0.2"
    {{CURRENT_PHASE}}         One short line describing what the project is
                              focused on right now. Update as work moves.

  Optional — safe to leave as-is early on, revisit as the project matures:
    {{DEPLOY_TARGET}}         Example: "Vercel", "Fly.io", "n/a (local only)"
    {{SECRETS_LOCATION}}      Example: ".env.local (gitignored), 1Password vault
                              'Acme Notes'"
    {{COMMIT_STYLE}}          Example: "conventional commits", "plain prose"
    {{BRANCH_NAMING}}         Example: "kebab-case, one branch per issue"

  Free-form slots use {{prose hints}} or _italic hints_ — fill them with
  plain English, not a single token.
-->

# md-notes

> {{PROJECT_TAGLINE}}

This file is the project rules for [Claude Code](https://claude.com/product/claude-code).
Claude reads it on every session. Keep it current — stale rules produce
worse output than no rules.

## What this is

{{One paragraph: what the project does, for whom, and roughly how. Link to
the PRD, MVP doc, or AI summary if they exist — `Design/mvp.md` and
`Design/ai-summary.md` are the standard locations in a kit-installed project.}}

## Current phase

{{CURRENT_PHASE}}

Active milestone: `{{CURRENT_MILESTONE}}`.

## Technology stack

- Runtime / language: {{PRIMARY_LANGUAGE}}
- Framework: {{FRAMEWORK}}
- Database / storage: {{DATA_LAYER}}
- Key libraries: {{KEY_LIBRARIES}}
- Package manager: {{PACKAGE_MANAGER}}
- Module system: {{MODULE_SYSTEM}}
- Deployment target: {{DEPLOY_TARGET}}

## How to run

```bash
{{INSTALL_COMMAND}}
{{DEV_COMMAND}}
{{BUILD_COMMAND}}
```

## Testing

- Framework: {{TEST_FRAMEWORK}}
- Location: {{TEST_LOCATION}}
- Run: `{{TEST_COMMAND}}`
- Coverage expectations: new modules include unit tests for the happy
  path, edge cases, and error handling. Existing tests must continue to
  pass on every PR.

## Code style

- Formatter: {{FORMATTER}}
- Linter: {{LINT_COMMAND}}
- Commit style: {{COMMIT_STYLE}}
- Secret management: {{SECRETS_LOCATION}}

## Project structure

```
md-notes/
  {{top-level directory tree with one-line descriptions. Keep it short —
    three or four entries is enough. Example:
    src/            application code
    test/           unit and integration tests
    Design/         ADRs, MVP doc, AI summary
    notes/          per-issue prompts and process notes
    .claude/skills/ installed workflow skills
  }}
```

See also:

- `Design/` — ADRs (`Design/adr/`), MVP (`Design/mvp.md`), AI summary (`Design/ai-summary.md`)
- `notes/` — per-issue prompts and working notes
- `.claude/skills/` — installed workflow skills (do not edit by hand)

## Workflow rules

This project follows the [Claude Code Workflow Kit](https://github.com/olivermorgan2/workflow-generator)
model. The rules below are load-bearing — Claude Code should treat them
as hard requirements unless a human overrides them in the session.

- **Plan-first execution.** When given a non-trivial task, propose a short
  step-by-step plan and wait for explicit approval before editing files.
- **Issue-by-issue.** Work is scoped to one GitHub issue at a time.
  Per-issue prompts live in `notes/issueNN-prompt.md`.
- **Consult ADRs before changing load-bearing behaviour.** If a change
  touches architecture, installation, or conventions, check
  `Design/adr/` first. Never edit an accepted ADR in place — supersede it
  with a new ADR.
- **Stay in scope.** Do not refactor unrelated code, rename files, or add
  speculative abstractions while working on an issue. If something is
  out of scope, note it for a follow-up issue.
- **Tests stay green.** Run `{{TEST_COMMAND}}` before declaring a task
  done. Fix regressions in the same PR that caused them.
- **Ask when ambiguous.** Prefer a clarifying question over a plausible
  guess on anything that affects scope, public API, or data shape.

## GitHub conventions

This project uses GitHub Flow — `{{DEFAULT_BRANCH}}` is always
deployable; all work lands via pull requests.

- **Repository:** `{{GITHUB_OWNER}}/{{GITHUB_REPO}}`
- **Default branch:** `{{DEFAULT_BRANCH}}`
- **Branch naming:** {{BRANCH_NAMING}}. Branch from `{{DEFAULT_BRANCH}}`,
  never commit to it directly, delete branches after merge.
- **One issue per branch / PR.** Split PRs that grow to cover multiple
  issues.
- **Labels:** `feature`, `bug`, `design`, `infra`, `security`, `docs`.
  Every issue gets exactly one primary label.
- **Milestones:** `Foundation`, `MVP`, `Post-MVP` (add more only when the
  list gets crowded).
- **Pull requests** include the sections defined in
  [`.github/pull_request_template.md`](.github/pull_request_template.md):
  Summary, `Closes #N`, ADR reference, Changes, Test results, Manual
  verification.
- **Commit messages** reference the ADR and issue when the change is
  driven by them. Example: `feat(auth): add session middleware (ADR-003, #15)`.
  Commit style: {{COMMIT_STYLE}}.

## Review expectations

- Every change lands via PR linked to a GitHub issue with `Closes #N`.
- Plan-first: propose a plan, wait for approval, then implement.
- Keep PRs small and focused. A PR should be reviewable in ~15 minutes.
- No direct commits to `{{DEFAULT_BRANCH}}`.
- Existing tests must pass; new behaviour has new tests.

## Definitions of done

A task is **done** when:

1. Code compiles / type-checks cleanly.
2. `{{TEST_COMMAND}}` passes locally.
3. `{{LINT_COMMAND}}` passes locally (skip if `none`).
4. The PR body fills in every required section.
5. The ADR (if any) and the issue number are referenced in the commit.
6. A human has approved the PR or explicitly waived review.

## What this file is NOT

- Not a spec — architectural decisions live in `Design/adr/`.
- Not a roadmap — phased plans live in `Design/build-out-plan.md`.
- Not an AI-readable summary — that is `Design/ai-summary.md`.

Keep this file focused on rules and conventions Claude Code needs to do
its job. When in doubt, link out.
