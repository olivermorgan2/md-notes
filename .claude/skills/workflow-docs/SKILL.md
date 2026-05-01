---
name: workflow-docs
description: Generate README.md and Design/ai-summary.md for a target project from PRD, MVP, ADRs, and CLAUDE.md — re-runnable with marker-fenced sections that preserve manual edits
---

# workflow-docs

Generate two project-level documents in a target project by reading the
artifacts the kit already produces and filling templates:

- `README.md` (public entry point for humans)
- `Design/ai-summary.md` (AI-readable project summary)

The skill is **re-runnable**. Generated regions are wrapped in comment
markers so a second run updates those regions without clobbering manual
edits outside them. Sections with no source data are **omitted**
entirely (heading + body) rather than left blank or filled with
placeholders.

See [ADR-018](../../Design/adr/adr-018-workflow-docs-skill.md) for the
decision to consolidate both outputs into one skill instead of two.

## When to use this skill

- After `prd-to-mvp` has produced `Design/mvp.md`, to generate a real
  README and AI summary.
- After landing one or more ADRs, to refresh the "Key decisions" section
  in both outputs.
- Whenever `CLAUDE.md`, the PRD, or the MVP statement has materially
  changed and the generated docs should reflect that change.

Typical invocation:

```
/workflow-docs
```

No arguments. The skill infers everything from files in the current
working directory (the target project root).

## What this skill does not do

- Does not modify the PRD, MVP statement, ADRs, or `CLAUDE.md`. These
  are **read-only inputs**.
- Does not invent content. A missing section source → the section is
  omitted, not filled with a placeholder or TODO.
- Does not overwrite files without explicit user approval. Both
  generated files are shown in full before any write.
- Does not publish, deploy, or push. Output is local files only.
- Does not touch content outside its `<!-- workflow-docs:start:... -->`
  / `<!-- workflow-docs:end:... -->` markers on re-run.
- Does not manage CLAUDE.md — that is rendered from
  `templates/claude-md-template.md` at install time or by a human.

## Inputs (read-only)

All inputs live in the **target project** (the current working
directory when the skill runs). Every input is optional — the skill
degrades gracefully. Missing inputs cause their derived sections to be
omitted.

| Source                | What it contributes                                          |
|-----------------------|--------------------------------------------------------------|
| `Design/prd.md` or `Design/prd-normalized.md` | Project description, primary user, problem statement |
| `Design/mvp.md`       | Product name, tagline, principles, In-scope, Out-of-scope, success criteria |
| `Design/adr/*.md`     | "Key decisions" bullets — one line per ADR, newest first     |
| `CLAUDE.md`           | Tech stack, commands (install, dev, test), current phase, milestone |

If **none** of these exist, stop with a clear message pointing the user
at `prd-normalizer` / `prd-to-mvp` first. There is nothing useful to
generate from an empty kit.

## Outputs

Both outputs are written to the **target project** root, next to the
inputs:

- `README.md` — rendered from
  [`templates/readme-template.md`](../../templates/readme-template.md).
- `Design/ai-summary.md` — rendered from
  [`templates/ai-summary-template.md`](../../templates/ai-summary-template.md).

On a first run, whole files are written. On a re-run, only the content
inside `<!-- workflow-docs:start:<section> -->` / `<!-- workflow-docs:end:<section> -->`
markers is regenerated; any other lines the user has added are left
untouched.

## Marker format (exact)

Every generated section is wrapped by two HTML comment markers on their
own lines:

```markdown
<!-- workflow-docs:start:overview -->
## Overview

One-paragraph description of the project, derived from the PRD and MVP.
<!-- workflow-docs:end:overview -->
```

Rules:

- Markers are always on their **own line**, with no leading whitespace.
- The section id (here `overview`) is kebab-case and stable across
  releases of this skill.
- The `start` marker precedes the section heading (not inside it), so
  the heading itself is regeneratable — this is what lets the skill
  cleanly **omit** a section on re-run if its source disappears.
- Content outside markers is never touched on re-run.
- If the user has moved markers or deleted the closing `end` marker,
  the skill flags the file as "manual edits — cannot safely update"
  and asks the user to either restore the markers or delete the
  generated file and regenerate.

## Section omission logic

Each generated section has a deterministic rule for when it appears:

### `README.md`

| Section         | Appears when                                                |
|-----------------|-------------------------------------------------------------|
| `tagline`       | `Design/mvp.md` has a one-line description, **or** PRD has a tagline |
| `overview`      | `Design/prd.md` / `prd-normalized.md` or `Design/mvp.md` exists |
| `who-for`       | PRD / MVP identifies a primary user                         |
| `status`        | `CLAUDE.md` has a non-placeholder `{{CURRENT_PHASE}}` or milestone |
| `scope`         | `Design/mvp.md` has an "In scope" and/or "Out of scope" section |
| `how-to-run`    | `CLAUDE.md` has non-placeholder commands                    |
| `key-decisions` | `Design/adr/` contains at least one ADR                     |
| `roadmap`       | `Design/build-out-plan.md` has 2+ `### Phase N` blocks (per ADR-032). Single-phase / no-phase plans omit this section. |
| `more`          | Always — static pointers to `CLAUDE.md`, `Design/`          |

### `Design/ai-summary.md`

| Section          | Appears when                                              |
|------------------|-----------------------------------------------------------|
| `objectives`     | `Design/mvp.md` has a product goal or success criteria    |
| `architecture`   | `CLAUDE.md` has a "Project structure" or tech stack block |
| `tech-stack`     | `CLAUDE.md` has any of runtime / framework / data layer   |
| `constraints`    | `Design/mvp.md` has principles, or PRD has constraints    |
| `extension-points` | `CLAUDE.md` or PRD notes extension points (rare on first run — usually omitted) |
| `current-status` | `CLAUDE.md` has a current phase, or ADRs are dated        |
| `key-decisions`  | `Design/adr/` contains at least one ADR                   |

A section is **omitted** by removing its heading, body, and both
markers — the file skips straight from the previous section to the
next. The template is written so adjacent omitted sections leave no
stray blank lines.

### Roadmap section content (per ADR-032)

When the `roadmap` section appears, render a compact table sourced
from `Design/build-out-plan.md`'s `### Phase N: <name>` blocks:

| # | Phase | Goal | Exit criterion |
|---|-------|------|----------------|
| 1 | Foundation | one-line goal | observable exit |
| 2 | Core feature | one-line goal | observable exit |
| 3 | Polish | one-line goal | observable exit |

Single-phase projects omit the roadmap section entirely — phase
metadata is implicit and adds no information to the README.

## Execution protocol

Run these steps in order. Stop on the first failure unless noted.

1. **Detect target project.** The current working directory must be the
   target project root. Confirm by checking for a `.git/` directory.
   If absent, ask the user to `cd` into the project root and stop.
2. **Collect inputs.** Read these files if they exist:
   - `Design/prd.md` (fall back to `Design/prd-normalized.md`)
   - `Design/mvp.md`
   - Every file matching `Design/adr/adr-*.md`
   - `CLAUDE.md`

   If none exist, stop with the message "No PRD, MVP, ADRs, or
   CLAUDE.md found — run `prd-normalizer` / `prd-to-mvp` first."
3. **Parse inputs into a single context dict.** Extract the fields
   listed under "Template variables" below. Any field whose source is
   missing is left as `None`. Do **not** substitute placeholder text.
4. **Determine which sections to include.** Apply the omission rules
   above to both outputs. Produce an ordered list of `(section_id,
   rendered_content)` tuples for each file.
5. **Render both files in memory.**
   - Read `templates/readme-template.md` and
     `templates/ai-summary-template.md` from the kit's
     `.claude/skills/workflow-docs/` copy or from the templates
     directory if running from the kit repo.
   - For each included section, fill `{{PLACEHOLDERS}}` from the
     context dict and wrap in its markers.
   - Drop the heading, body, and both markers for any omitted section.
   - Collapse any resulting run of more than one blank line to exactly
     one blank line.
6. **Check for existing files and merge markers.**
   - If `README.md` or `Design/ai-summary.md` already exists **and
     contains** the skill's markers: parse the existing file, replace
     only the content between each `start:/end:` pair, leave
     everything else alone.
   - If the file exists but has **no markers** (e.g. a human-written
     README from before the skill was run): treat it as a first run —
     show a diff against the generated content, warn the user that
     accepting will replace the existing file, and require an explicit
     "yes, replace".
   - If a section is in the existing file but the new run **omits** it
     (source disappeared), delete the block between markers including
     the markers themselves.
7. **Show both full file contents in chat** as fenced markdown blocks
   — or show a diff if the file already exists. Ask explicitly:
   "Write these files? (yes / edit / cancel)". Default: no.
8. **Write only after confirmation.** Report the absolute paths and a
   one-line summary of what was written, regenerated, or omitted per
   file.

## Template variables

Filled into `{{PLACEHOLDER}}` slots from the sources below. Every
field is optional; missing fields drive section omission (see above).

### README (`templates/readme-template.md`)

| Placeholder            | Source                                                    |
|------------------------|-----------------------------------------------------------|
| `{{PROJECT_NAME}}`     | `Design/mvp.md` "Product name", else `CLAUDE.md` first `#` heading, else repo basename |
| `{{PROJECT_TAGLINE}}`  | `Design/mvp.md` "One-line description", else PRD tagline  |
| `{{PROJECT_DESCRIPTION}}` | `Design/mvp.md` "Product goal", else PRD "Goal" / "Problem" paragraph |
| `{{PRIMARY_USER}}`     | `Design/mvp.md` "Primary user", else PRD "Target user"    |
| `{{CURRENT_PHASE}}`    | `CLAUDE.md` "Current phase" section                       |
| `{{CURRENT_MILESTONE}}`| `CLAUDE.md` active milestone                              |
| `{{IN_SCOPE_BULLETS}}`     | `Design/mvp.md` "In scope" bullets                    |
| `{{OUT_OF_SCOPE_BULLETS}}` | `Design/mvp.md` "Out of scope" bullets                |
| `{{INSTALL_COMMAND}}`  | `CLAUDE.md` install command                               |
| `{{DEV_COMMAND}}`      | `CLAUDE.md` dev command                                   |
| `{{TEST_COMMAND}}`     | `CLAUDE.md` test command                                  |
| `{{KEY_ADR_BULLETS}}`  | One bullet per `Design/adr/adr-NNN-*.md`, newest first: `- ADR-NNN: <one-line decision>` |

### AI summary (`templates/ai-summary-template.md`)

| Placeholder            | Source                                                    |
|------------------------|-----------------------------------------------------------|
| `{{PROJECT_NAME}}`     | Same as README                                            |
| `{{YYYY-MM-DD}}`       | Today's date                                              |
| "Objectives" bullets   | `Design/mvp.md` "Product goal" + "Success criteria"       |
| "Architecture" paragraph | `CLAUDE.md` "Project structure" + "Technology stack"    |
| "Tech stack" bullets   | `CLAUDE.md` runtime / framework / data layer / libraries / deploy target |
| "Constraints" bullets  | `Design/mvp.md` "Product principles", PRD "Constraints"   |
| "Extension points" bullets | PRD / `CLAUDE.md` extension-point notes (usually omitted on first run) |
| "Current status" bullets | `CLAUDE.md` "Current phase", ADR dates, issue activity if available |
| "Key decisions" bullets | One bullet per ADR, newest first                         |

For the AI summary, also wrap each section of
`templates/ai-summary-template.md` in the same marker format before
rendering. Section ids mirror the template headings in kebab-case.

## Edge cases

- **No PRD, no MVP, no ADRs, no CLAUDE.md** → stop; tell the user to
  run `prd-normalizer` / `prd-to-mvp` first. Do not write empty docs.
- **Only `CLAUDE.md` exists** → generate a minimal README (tagline +
  how-to-run + more) and a minimal AI summary (tech stack + current
  status). Omit all other sections.
- **Only ADRs exist** → generate only the `key-decisions` section in
  both outputs. Ask the user whether that is useful before writing.
- **Existing README has no markers** → treat as first run; show diff;
  require explicit confirmation; do not silently overwrite.
- **Existing README has markers but they are malformed** (missing
  `end`, duplicated `start`, moved out of order) → stop and ask the
  user to restore or delete the file.
- **Existing file has manual content inside a marker block** → that
  content is **regenerated** (markers are the contract for "this is
  regenerated"). Manual edits belong **outside** markers. Warn the user
  if the diff shows any non-trivial change inside a marker block.
- **ADR files missing a `# ADR-NNN:` heading** → skip that ADR with a
  warning; do not stop.
- **`CLAUDE.md` still has `{{PLACEHOLDER}}` tokens** → treat those
  fields as missing; drive section omission accordingly.
- **Multiple PRD files (`prd.md` + `prd-normalized.md`)** → prefer
  `prd-normalized.md` (the kit's canonical form).

## Review-before-write checkpoint

The user always sees both generated files (as full content on first
run, or as a diff against the existing file on re-run) before any
write. This is the approval gate for ADR-018.

Never skip this step, even if the content looks clean. Never write one
file while deferring the other — both go out together after one
confirmation, or neither.

## Self-check before writing

For each output file:

- [ ] Every `{{PLACEHOLDER}}` is either filled or its enclosing section
      has been omitted.
- [ ] Every included section is wrapped in matching `start:/end:`
      markers with the same section id.
- [ ] No omitted section has left a stray heading, body, or marker
      behind.
- [ ] On re-run: content outside markers in the existing file is
      byte-identical to what it was before.
- [ ] The user explicitly confirmed the write.

If any fail, fix and re-show before writing.

## Running against the kit repo itself

This repo (`workflow-generator`) has `CLAUDE.md` and `Design/adr/*.md`
but no `Design/prd.md` or `Design/mvp.md`. Running the skill here would
produce:

- **README.md**: the `scope`, `overview`, `who-for`, `tagline`, and
  `how-to-run` sections would be **omitted** (no PRD, no MVP, and
  `CLAUDE.md` does not carry commands in this repo). The `status`
  section may also be omitted if `CLAUDE.md` has no current-phase
  line. Only `key-decisions` (from ADRs) and `more` (static pointers)
  would be included — and because the kit repo already has a rich
  hand-written README, the skill would detect the missing markers and
  refuse to overwrite without an explicit "yes, replace".
- **Design/ai-summary.md**: only `key-decisions` (from ADRs) and a
  minimal `current-status` section would be included. Objectives,
  architecture, tech stack, constraints, and extension points would
  all be omitted.

In practice this means the kit repo should **not** run `/workflow-docs`
on itself. The skill is aimed at target projects that have already run
`prd-to-mvp`.

## Handoff

Once the two files are written, the user can edit outside markers
freely. Re-running the skill will only touch marked regions. If a
future run needs to add a new section, this skill's next release will
add a new marker id — unrecognised existing markers are left
untouched.

See [`example.md`](example.md) for a worked run on a small sample
project.
