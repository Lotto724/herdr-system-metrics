# herdr-system-metrics Agent Guidance

This is an independent, pre-implementation Go repository for a Herdr system-metrics popup. Keep work inside this repository and within the approved Linux and WSL scope. Do not add functional behavior, host integration assumptions, or release claims unless the task explicitly authorizes them.

The repository owns its quality interface through [`Makefile`](Makefile), pinned tools through [`go.mod`](go.mod), and CI through [`.github/workflows/quality.yml`](.github/workflows/quality.yml). Do not replace those interfaces with agent-specific commands or tooling.

## Project Skill

Load the project skill before formatting, linting, type-checking, testing, quality, verification, or CI work.

| Skill | Trigger | Path |
| --- | --- | --- |
| `project-quality` | Format, lint, typecheck, quality, tests, verification, or CI | [`skills/project-quality/SKILL.md`](skills/project-quality/SKILL.md) |
