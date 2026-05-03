# Architecture Decision Records

This directory holds the ADRs for md-notes. Each ADR records one
architectural decision: its context, the options considered, the
choice made, and the consequences.

ADRs are immutable once `accepted`. To change an accepted decision,
draft a new ADR that supersedes the old one — never edit the old
file in place.

## Index

<!-- adr-index:start -->

| ADR | Title | Status |
|---|---|---|
| [ADR-001](adr-001-language-choice.md) | Language choice — Go | accepted |
| [ADR-002](adr-002-configuration-surface.md) | Configuration surface — flag > env > file > default | accepted |
| [ADR-003](adr-003-filename-convention.md) | Filename convention — `YYYY-MM-DD-slug.md` with numeric collision suffix | accepted |
| [ADR-004](adr-004-frontmatter-shape.md) | Frontmatter shape — minimal YAML (`title`, `date`, `tags: []`) | accepted |
| [ADR-005](adr-005-search-backend.md) | Search backend — in-process scan in Go | accepted |

<!-- adr-index:end -->
