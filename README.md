# Kyver

> Git versions source code. Kyver versions engineering knowledge.

Kyver (Engineering Knowledge Version Control - EKVC) is a tool that runs alongside Git (`.git/`) by creating a `.kyver/` repository to store engineering knowledge.

## Features

- Requirements
- Architecture Decision Records (ADR)
- AI Prompt History
- Knowledge Commits
- Engineering Memory
- Requirement Traceability
- Knowledge Graph
- Change Impact Analysis
- AI Agent History

## Architecture

This project is built with clean architecture in Go.

- `cmd/kyver`: The CLI entry point.
- `internal/cli`: CLI commands (`kyver init`, `kyver capture`, etc.)
- `internal/core`: Business logic, Repository Engine, and Knowledge Engine.
- `internal/storage`: Abstractions for storing metadata.
- `internal/config`: Configuration schema.
- `internal/ai` & `internal/cloud`: Placeholder domains for future integrations.

## Commands (Planned)

- `kyver init`
- `kyver capture`
- `kyver commit`
- `kyver why`
- `kyver sync`
- `kyver status`
- `kyver doctor`

## Semantic Commands

- `kyver context`
- `kyver search`
- `kyver goto`
- `kyver peek`
- `kyver outline`
- `kyver hierarchy`
- `kyver callers`
- `kyver callees`
- `kyver graph`
- `kyver impacts`
- `kyver references`
- `kyver implements`
- `kyver interface`
- `kyver ask`
