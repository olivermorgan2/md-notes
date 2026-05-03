# ADR-004: Frontmatter shape — minimal YAML (`title`, `date`, `tags: []`)

**Status:** accepted
**Date:** 2026-05-02

## Context

Every note created by `notes new` carries YAML frontmatter at the top. The MVP explicitly requires `tags: []` to be written even though tag filtering is deferred to v0.2 — Risk 2 in `Design/build-out-plan.md` is "frontmatter shape in v0.1 forces migration in v0.2", and the mitigation is to lock the shape now so v0.2 adds keys without rewriting older notes.

Constraints from the MVP and product principles:

- Notes must remain editable in any external editor — frontmatter must be standard YAML, not a custom dialect.
- The tool must "stay out of the way" — minimal noise at the top of every file.
- v0.2 will add tag-filtering and possibly daily notes; the v0.1 shape must accommodate those without migration.
- Many tools in the markdown ecosystem (Obsidian, Hugo, Jekyll, MkDocs) read YAML frontmatter — staying close to the common subset preserves the "any tool can edit" principle.

## Options considered

### Option A: Bare minimum — `date` and `tags: []` only

- Pros:
  - Minimal noise; smallest frontmatter block.
  - Less to type by hand if a user creates a note outside `notes new`.
- Cons:
  - Title is in the filename only; `notes find` and `notes ls` cannot show a clean title without parsing the slug.
  - Slug-derived titles are lossy (case folded, punctuation stripped) — losing the original title is irreversible without scanning the body for an `# H1`.

### Option B: Minimal — `title`, `date`, `tags: []`

- Pros:
  - Original (un-slugified) title is preserved verbatim — `notes ls` and `notes find` can show clean human-readable titles without parsing the body.
  - Three keys is still terse.
  - Stable shape: v0.2 adds keys (e.g. `daily: true` for `notes today`) without touching existing notes.
- Cons:
  - Slight redundancy with the filename slug — but the filename is lossy and the title is the user's verbatim string, so this is real information not just duplication.

### Option C: Rich — `title`, `date`, `tags: []`, `id`, `modified`, `aliases`

- Pros:
  - Sets the project up for cross-note linking (`id`, `aliases`) and modification tracking.
- Cons:
  - Cross-linking is an explicit MVP non-goal ("Cross-note linking, backlinks, or graph view" in `Design/mvp.md`).
  - `modified` duplicates `mtime` on the filesystem and risks drift if the user edits the file outside the tool.
  - Adding fields the MVP does not use creates noise the user must learn to ignore; once written, fields are sticky.

## Decision

Choose **Option B: `title`, `date`, `tags: []`.** These three keys cover everything v0.1 needs (preserved title for display, date for sort/scan, tags for v0.2 forward compatibility) without paying for non-goal capabilities. The shape is stable: v0.2's tag filtering reads existing `tags: []` directly with no migration; v0.2's `notes today` adds a `daily: true` key without disturbing existing notes.

Concrete shape:

```yaml
---
title: "Design ideas"
date: 2026-05-02
tags: []
---
```

`title` is always quoted to handle colons, quotes, and Unicode safely. `date` is `YYYY-MM-DD` (matches the filename prefix). `tags` is always emitted, even when empty.

## Consequences

- **Easier:** v0.2 tag filtering reads existing frontmatter unchanged; `notes ls` and `notes find` can display the verbatim title without slug-reversing; the shape is the common subset of Obsidian / Hugo / Jekyll, so users importing or exporting notes are unsurprised.
- **Harder:** `notes new` must safely YAML-encode the title (quoting, escaping); a unit test covers titles containing quotes, colons, and Unicode.
- **Maintain:** the frontmatter shape is documented in `README.md` and locked by a unit test. Adding a new key in v0.2 is allowed; removing or renaming an existing key triggers a superseding ADR.
- **Deferred:** richer fields (`id`, `aliases`, `modified`) — not added in v0.1; revisit only when a feature actually needs them.
