# Install Guide

How to install the Claude Code Workflow Kit into a **new** project. The
kit is project-agnostic (per [ADR-028](../Design/adr/adr-028-workflow-agnostic-framing.md))
— software, research, content, curriculum, or any other structured
project that fits the kit's three assumptions: a git repo, GitHub, and
Claude Code.

> **Before you start:** the kit is for **new projects only**. It does
> not adapt or migrate existing repositories. If you want to add the
> workflow to an existing codebase, this kit is not the right fit —
> see [What this kit does not support](#what-this-kit-does-not-support) below and
> [ADR-002](../Design/adr/adr-002-new-project-only-scope.md).

The kit uses a **project-local installation model**: each new target
project gets its own copy of the skills under `.claude/skills/`. There is
no global install. See [ADR-001](../Design/adr/adr-001-project-local-installation-model.md)
for the reasoning.

---

## 1. Prerequisites

Install these once per machine. The kit assumes macOS or Linux; commands
are cross-platform where possible.

| Tool | Purpose | Install |
|---|---|---|
| **Git** | Version control | Ships with macOS (Xcode CLT) or `brew install git` |
| **GitHub CLI (`gh`)** | Repo, issue, and PR operations from the terminal | `brew install gh` then `gh auth login` |
| **Claude Code** | AI-assisted development | `npm install -g @anthropic-ai/claude-code` |
| **SSH key** | Passwordless git over SSH | `ssh-keygen -t ed25519` then add the public key to GitHub |

### Verify the setup

```bash
git --version          # any recent 2.x is fine
gh --version           # any recent 2.x is fine
gh auth status         # should report "Logged in to github.com"
claude --version       # confirms Claude Code is on your PATH
ssh -T git@github.com  # "Hi <user>! You've successfully authenticated"
```

If any of these fail, fix the failing tool before continuing — the rest of
the guide assumes they all pass.

---

## 2. Choose your target project

Pick **one** of these two paths. The rest of the guide assumes you end up
inside the target project directory.

**A. Create the repo on GitHub first** (recommended):

```bash
gh repo create my-project --public --clone   # use --private if appropriate
cd my-project
```

**B. Start locally, add the remote later:**

```bash
mkdir my-project && cd my-project
git init
```

---

## 3. Install the kit into the target project

You have three options, in order of preference:

- **Bootstrap install** (recommended for users) — one command pulls the
  kit at a pinned version from GitHub, runs the installer, and discards
  the temporary kit copy. See [3A](#3a-bootstrap-install-recommended)
  below.
- **Explicit fetch** — three explicit lines that do the same thing as
  the bootstrap script, more transparently. See
  [3B](#3b-explicit-fetch-alternative) below.
- **Manual install** — full do-it-yourself copy flow. See
  [3C](#3c-manual-install) below.

All three produce the same target-project layout described in
[`docs/repo-structure.md`](repo-structure.md). Per ADR-029, the kit is
designed for **per-project remote install** — there is no long-lived
local kit clone to maintain. Contributors who edit the kit itself want
a different setup; see
[Contributor / kit-developer setup](#contributor--kit-developer-setup)
near the end of this doc.

### 3A. Bootstrap install (recommended)

The bootstrap script (per ADR-029) ships as an asset on every tagged
release. It fetches the kit at the pinned version, runs the installer
against the current directory (or `--target=PATH`), and cleans up the
temporary kit copy on exit.

```bash
cd my-project                                              # target project root
bash <(curl -fsSL https://github.com/olivermorgan2/workflow-generator/releases/download/v3.2.0/bootstrap-workflow-kit) \
  --project-name=my-project
```

Prefer to inspect the script before running, or pipe-to-bash makes
you nervous? Download it first:

```bash
gh release download v3.2.0 -p bootstrap-workflow-kit \
  -R olivermorgan2/workflow-generator
chmod +x bootstrap-workflow-kit
./bootstrap-workflow-kit --project-name=my-project
```

What the bootstrap does, in order:

1. Resolves the target version (env var `WORKFLOW_KIT_VERSION`, then
   `gh release view`, then `git ls-remote --tags`).
2. Creates a temporary directory, traps cleanup on exit.
3. Fetches the kit (`gh repo clone --depth=1 --branch=vX.Y.Z` if `gh`
   is available; falls back to `git clone --depth=1` over HTTPS).
4. Verifies the fetched tree contains `bin/install-workflow-kit`.
5. Forwards all CLI args to that installer.

Environment variables:

| Variable | Default | Purpose |
|---|---|---|
| `WORKFLOW_KIT_VERSION` | latest tag | Pin to a specific tag, e.g. `v3.2.0`. |
| `WORKFLOW_KIT_REPO` | `olivermorgan2/workflow-generator` | Override for forks. |

The installer's behaviour itself is unchanged from ADR-009:

1. Creates `Design/adr/`, `prompts/`, `notes/`, and `.claude/skills/`
   in the target project.
2. Copies every skill from the kit's `skills/` into `.claude/skills/`.
3. Copies `bin/sync-adr-index` into `.claude/bin/` (ADR-023).
4. Seeds `prompts/_template.md` so new per-issue prompts stay
   consistent (ADR-008).
5. Renders `CLAUDE.md` from `templates/claude-md-template.md`
   (ADR-007). You are prompted for any `{{UPPER_SNAKE}}` placeholder
   not provided on the command line, unless `--non-interactive` is set.
6. `git init`s the target if it is not already a git repo, then
   creates an initial commit `chore: install workflow kit
   (project-local)`.

Useful installer flags (forwarded by the bootstrap):

| Flag | What it does |
|---|---|
| `--target=PATH` | Target project directory (default: current dir) |
| `--project-name=NAME` | Value for `{{PROJECT_NAME}}` in `CLAUDE.md` |
| `--set KEY=VALUE` | Provide a value for any other `{{UPPER_SNAKE}}` placeholder. Repeatable. |
| `--with-docs` | Also copy kit docs to `docs/workflow-kit/` in the target (ADR-010) |
| `--force` | Overwrite existing `CLAUDE.md` and re-copy skills |
| `--no-commit` | Skip the initial commit |
| `--non-interactive` | Never prompt; fall back to defaults |
| `--license=ID` | Scaffold a starter `LICENSE` in the target. Supported: `mit`. Default: no `LICENSE` is written (license choice is the project author's call, per ADR-025). |
| `--license-holder=NAME` | Copyright holder for the rendered `LICENSE` (only used when `--license` is set). Falls back to `--project-name`, then to the target's basename. |
| `-h`, `--help` | Show full usage |

> **License scaffolding example.** Add an MIT LICENSE attributed to
> "Jane Doe" alongside the rest of the install:
>
> ```bash
> bash <(curl -fsSL https://github.com/olivermorgan2/workflow-generator/releases/download/v3.2.0/bootstrap-workflow-kit) \
>   --project-name=my-project \
>   --license=mit \
>   --license-holder="Jane Doe"
> ```

The installer is idempotent: re-running on an already-installed target
skips files that already exist and makes no commit if nothing changed.
Pass `--force` to re-render `CLAUDE.md` or re-copy the skills.

### 3B. Explicit fetch alternative

The bootstrap script is convenient, but if you want to see exactly
what is happening, this three-line form does the same thing — fetch
the kit at a pinned version into a temp dir, run the installer,
clean up:

```bash
TMPKIT="$(mktemp -d)" && \
  gh repo clone olivermorgan2/workflow-generator "$TMPKIT" -- \
    --depth=1 --branch=v3.2.0 && \
  "$TMPKIT/bin/install-workflow-kit" --project-name=my-project && \
  rm -rf "$TMPKIT"
```

Replace `v3.2.0` with whichever release you want to pin. Replace the
installer flags with whatever your project needs (see the table in
[3A](#3a-bootstrap-install-recommended)).

If `gh` is not installed, swap the clone line for plain git over
HTTPS:

```bash
git clone --depth=1 --branch=v3.2.0 \
  https://github.com/olivermorgan2/workflow-generator.git "$TMPKIT"
```

### 3C. Manual install

The manual flow is the explicit, do-it-yourself version of what the
installer does. Use it when you want full visibility, or as a
diagnostic reference if the installer misbehaves.

#### 3C.1 Fetch the kit at a pinned version

```bash
TMPKIT="$(mktemp -d)"
gh repo clone olivermorgan2/workflow-generator "$TMPKIT" -- --depth=1 --branch=v3.2.0
# or:  git clone --depth=1 --branch=v3.2.0 https://github.com/olivermorgan2/workflow-generator.git "$TMPKIT"
```

#### 3C.2 From inside the target project, copy the skills

```bash
cd my-project                                   # target project root
mkdir -p .claude/skills .claude/bin Design/adr prompts notes
cp -R "$TMPKIT/skills/"* .claude/skills/
cp "$TMPKIT/bin/sync-adr-index" .claude/bin/
chmod +x .claude/bin/sync-adr-index
cp "$TMPKIT/prompts/_template.md" prompts/_template.md
```

The `prompts/` folder holds one filled session brief per GitHub
issue (`issue-NNN-short-title.md`), copied from `_template.md`.
Freeform working notes stay in `notes/`. See
[ADR-008](../Design/adr/adr-008-dedicated-prompts-folder.md).

This is the step that makes the install **project-local**: the
skills now live under the target project's own `.claude/skills/`,
tracked in the target project's git history. The temporary kit
clone is no longer required for those skills to work.

#### 3C.3 Render the starter `CLAUDE.md`

`CLAUDE.md` at the project root is Claude Code's primary rules file
for the project. Copy the template and fill in the project-specific
fields by hand:

```bash
cp "$TMPKIT/templates/claude-md-template.md" CLAUDE.md
```

The template uses `{{PLACEHOLDER}}` tokens (ADR-007). Open
`CLAUDE.md` and replace each one with real values, starting with
`{{PROJECT_NAME}}`. The automated installer above does this
substitution for you.

#### 3C.4 Commit the install and clean up

```bash
git add .claude CLAUDE.md Design prompts notes
git commit -m "chore: install workflow kit (project-local)"
rm -rf "$TMPKIT"
```

---

## 4. Verify the install

After step 3 the target project should look like this (trimmed):

```
my-project/
  CLAUDE.md
  .claude/
    skills/
      idea-to-prd/          # plus the other skills shipped by the kit
  Design/
    adr/
  prompts/
    _template.md            # blank template; copy to start an issue session
  notes/
```

A quick check:

```bash
ls .claude/skills          # should list the skill directories
ls prompts                 # should list _template.md
cat CLAUDE.md | head -20   # should show your project rules, not this kit's
```

For the full expected layout and which artifacts originate where, see
[`docs/repo-structure.md`](repo-structure.md).

---

## 5. Pick a starting path

The kit supports three ways to start planning, defined in
[ADR-003](../Design/adr/adr-003-prd-intake-model.md). Pick the one that
matches what you have in hand:

| You have… | Use this skill | Status |
|---|---|---|
| A rough idea, no PRD | `idea-to-prd` | Issue #5 |
| A standard PRD | `prd-normalizer` → `prd-to-mvp` | Issue #6, #7 |
| Mixed notes / custom PRD format | `prd-normalizer` → `prd-to-mvp` | Issue #6, #7 |

Run the chosen skill inside the target project using Claude Code. For
how to invoke skills, use plan mode, and follow the approve-then-implement
loop, see [`claude-code-guide.md`](claude-code-guide.md). Each skill's
own `SKILL.md` remains the authoritative spec for its inputs and outputs.

---

## Contributor / kit-developer setup

The flows above are for **users** of the kit — people scaffolding
their own projects. If you are working on the kit itself
(adding skills, fixing templates, drafting ADRs), you want a
long-lived local clone of the kit's source, not the bootstrap flow.

```bash
gh repo clone olivermorgan2/workflow-generator ~/src/workflow-generator
cd ~/src/workflow-generator
~/dotfiles/claude-config/bin/link-skills           # one-time, dogfooding
```

`link-skills` symlinks the kit's `skills/<name>/` directories into
its own `.claude/skills/<name>/`, so the kit can use its own skills
while you develop them. This convention is documented in `CLAUDE.md`
under "Developing the kit on itself (dogfooding)" and in the
dogfooding playbook at
`~/dotfiles/claude-config/docs/dogfooding-playbook.md`.

The contributor clone is **not** the same as the legacy "clone once
and use it forever" install model. Per ADR-029, end-user installs
go through the bootstrap flow and never need a long-lived kit
clone — the contributor clone exists only because the kit's source
is what contributors edit.

---

## What this kit does not support

The kit intentionally does **not** include:

- **Retrofitting existing repositories.** The kit assumes a clean slate.
  Pointing it at an established codebase will not produce useful output.
- **A global install.** Every target project gets its own copy of the
  skills. This is a design choice, not a limitation (see ADR-001).
- **Team/multi-repo features.** The kit is scoped to a solo author or
  small team working on a single project at a time.
- **Non-GitHub providers.** The workflow is GitHub-first
  ([ADR-004](../Design/adr/adr-004-github-first-workflow-model.md)).
  GitLab, Bitbucket, and others are not supported.
- **Non-Claude AI tools.** The skills target Claude Code specifically
  ([ADR-006](../Design/adr/adr-006-claude-code-execution-model.md)).

If any of the above is a hard requirement for your project, this release
is not the right fit — revisit after Phase 2.

---

## Troubleshooting

**`gh: command not found`**
Install the GitHub CLI (`brew install gh` on macOS) and run `gh auth login`.

**`gh auth status` reports you are logged out**
Run `gh auth login`. Choose SSH as the git protocol to match the clone
command in step 3B.1.

**`cp: ~/src/workflow-generator/skills/*: No such file or directory`**
The kit was cloned to a different path than the guide assumes. Either
re-clone to `~/src/workflow-generator` or substitute your actual path in
step 3B.2 (or pass the correct path to `bin/install-workflow-kit`).

**`install-workflow-kit: error: kit skills/ not found`**
The installer resolves its sources relative to its own location. Run it
directly from the kit clone (e.g. `~/src/workflow-generator/bin/install-workflow-kit`),
not via a copy detached from the kit repo.

**Claude Code does not see the skills**
Confirm you are running Claude Code from the **target project's root**.
Project-local skills are discovered relative to the current working
directory. If you `cd` into a subdirectory, skills still resolve, but if
you run Claude Code from an unrelated directory they will not.

**I already have a `Design/` or `.claude/` in the target project**
The copy commands in step 3 use `mkdir -p` and directory-level `cp -R`,
so they will not overwrite files that already exist with different
names. If you want to replace them wholesale, remove them first.

---

## What's next

- [`docs/workflow-guide.md`](workflow-guide.md) — end-to-end flow from idea to shipped release
- [`docs/repo-structure.md`](repo-structure.md) — how the kit is laid out and what it generates
- [`docs/github-setup.md`](github-setup.md) — GitHub labels, milestones, and branch conventions
