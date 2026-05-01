# Kit Documentation

This directory holds the documentation for the Claude Code Workflow Kit itself:
how to install it into a new target project, how to set up GitHub for the
generated workflow, and how each skill is meant to be used.

These docs live in the **kit repository only**. They are not copied into
target projects. A target project gets its own, project-specific docs generated
by the kit's skills (see `docs/repo-structure.md`).

## Planned documents

| File | Purpose | Added in |
|---|---|---|
| [`repo-structure.md`](repo-structure.md) | How the kit is laid out and what gets copied into target projects | Issue #1 |
| [`install.md`](install.md) | How to install the kit into a new target project | Issue #2 |
| [`github-setup.md`](github-setup.md) | GitHub repo / labels / branches / PR setup for generated projects | Issue #3 |
| [`issue-prompt-guide.md`](issue-prompt-guide.md) | How to fill the reusable Claude Code session prompt and what the evaluation summary must contain | Issue #8 |
| `workflow-guide.md` | End-to-end flow from idea or PRD to issue execution | later |
| `adr-guide.md` | When and how to write ADRs in a generated project | later |
| `claude-code-guide.md` | How to use the installed skills from inside a target project | later |
