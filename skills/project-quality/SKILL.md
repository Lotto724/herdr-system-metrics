---
name: project-quality
description: "Trigger: format, lint, typecheck, quality, tests, CI. Apply repository quality gates without changing the toolchain."
license: Apache-2.0
metadata:
  author: "Lotto724"
  version: "1.0"
---

## Activation Contract

Load this skill before formatting, linting, type-checking, testing, running quality gates, changing CI, or preparing review evidence.

## Hard Rules

- Treat `Makefile`, `go.mod`, and `.github/workflows/quality.yml` as authoritative.
- Run `make format` only before review START. After snapshot identity freezes, use check-only commands; stop for authorization if mutation is required.
- Never adopt, replace, configure, or upgrade quality tools, hooks, or CI independently.
- Do not hide failures or add tests for behavior that does not exist.

## Decision Gates

| Changed path | Required commands |
| --- | --- |
| `*.go`, `go.mod`, `go.sum` | `make quality`, then `make verify` |
| `Makefile`, `.github/workflows/quality.yml` | Relevant target, then `make quality` and `make verify` |
| Documentation or agent guidance only | `git diff --check`; run broader gates when requested |
| Any Go change reviewed on Linux/amd64 | Add `make test-race` |

## Execution Steps

1. Inspect changed paths and select gates above.
2. Before review START, run `make format` if needed; rerun it and require an empty second formatting diff.
3. Run selected check-only gates and `git diff --check`.
4. Record each command, exit status, and meaningful output. Record `git status --short` and the exact diff.
5. For a demonstrably pre-existing failure, reproduce it on the unchanged baseline or cite prior captured evidence; report it separately and do not modify unrelated files.
6. Stop for authorization before changing quality tooling, CI policy, hooks, frozen identity, or task scope.

## Output Contract

Return changed paths, commands and results, pre-existing failures, skipped gates with reasons, and stop conditions.

## References

- [`../../Makefile`](../../Makefile)
- [`../../go.mod`](../../go.mod)
- [`../../.github/workflows/quality.yml`](../../.github/workflows/quality.yml)
