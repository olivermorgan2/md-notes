# ADR-002: Configuration surface — flag > env > file > default

**Status:** accepted
**Date:** 2026-05-02

## Context

md-notes needs to resolve at least one piece of configuration on every command: the **notes directory** (default `~/notes/`). Future configuration may include editor override, default tag list, or search-backend tuning. The PRD's "Constraints and preferences" calls the configuration mechanism unresolved (flag, env var, config file, or combination). The MVP's principles include "the tool stays out of the way" — which argues against forcing the user to learn yet another config-file format for a one-knob tool, but also argues for predictability when the user has a non-default setup.

The decision affects every command's startup path and is hard to change later without a migration of user setups.

## Options considered

### Option A: Flag only

- Pros:
  - Simplest possible implementation.
  - Explicit; no hidden state.
- Cons:
  - User must pass `--dir ~/notes` on every invocation, which contradicts "stays out of the way" for the common case.
  - Shell aliases become the de-facto config mechanism, defeating the purpose.

### Option B: Environment variable only (`MD_NOTES_DIR`)

- Pros:
  - One-time setup in the user's shell rc file; no per-invocation noise.
  - Familiar Unix idiom; no new file format.
- Cons:
  - Discoverability is poor — `--help` can document it but users forget.
  - Hard to override per-invocation without the explicit flag option as well.

### Option C: All three with documented precedence — `flag > env > file > default`

- Pros:
  - Each layer has a clear use case: flag for one-off override, env for shell integration, file for the durable user setup, default for fresh installs.
  - Conventional Unix shape; matches user expectations from `git`, `gh`, `ripgrep`, etc.
  - Per-invocation overrides remain trivial.
- Cons:
  - Three resolution paths to implement and document.
  - Subtle bugs possible if precedence is implemented wrongly; needs unit tests for each combination.

## Decision

Choose **Option C: all three layers with precedence `flag > env > file > default`.** This is the conventional shape for Unix CLIs and gives the user the right tool for each situation — `--dir` for one-offs, `MD_NOTES_DIR` for shell integration, `~/.config/md-notes/config.toml` for a persistent setup, and `~/notes/` as the fresh-install default. The implementation cost is small (three reads in priority order) and the unit-test surface is bounded.

The config file format is **TOML** at `~/.config/md-notes/config.toml` (XDG-respecting), with a single `[paths]` section in v0.1 — only the `dir` key is recognised. Adding keys later is non-breaking.

## Consequences

- **Easier:** users with a non-default notes directory configure it once and stop thinking about it; ad-hoc overrides remain trivial; future config keys (editor override, default tags) slot into the same precedence chain.
- **Harder:** initial config-loader implementation has three layers instead of one; needs explicit unit tests per layer and per combination.
- **Maintain:** a documented precedence in `README.md`; an XDG-aware config-file path resolver; unit tests covering flag-overrides-env, env-overrides-file, file-overrides-default, and default-when-nothing-set.
- **Deferred:** project-local config (`./.md-notes.toml`) for per-directory overrides — re-evaluate in v0.2 if the use case appears.
