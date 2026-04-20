# Redis-Like Golang

A lightweight Redis-like key-value server built in Go. It exposes a simple TCP interface, supports TTL-based expiration, and can optionally persist write operations using an append-only file (AOF).

## Overview

This project is a learning-oriented Redis-inspired server that focuses on core data operations and clean architecture principles. It includes a server entrypoint, a minimal CLI client, in-memory storage with expiration handling, and optional AOF replay on startup.

## Features

- In-memory key-value store with thread-safe access
- Core commands: `SET`, `GET`, `DEL`
- Expiration commands: `EXPIRE`, `TTL`, `PERSIST`
- Utility commands: `KEYS`, `EXISTS`, `PING`, `INFO`, `QUIT`
- Wildcard key matching for `KEYS` (`*` and `?` patterns)
- Optional AOF persistence and replay (`-aof` flag)
- Graceful server shutdown support
- Unit and integration test coverage

## Tech Stack

- **Language:** Go (`go 1.26.1` in `go.mod`)
- **Networking:** Go standard library (`net`, `bufio`)
- **Concurrency:** Goroutines, `sync.RWMutex`, `sync/atomic`
- **Dependency Injection:** [Google Wire](https://github.com/google/wire)
- **Testing:** Go `testing` package (unit + integration tests)

## Architecture

The codebase follows a layered structure inspired by clean architecture:

- `internal/domain`: Core entities, command types, and repository contracts
- `internal/usecase`: Command execution and stats logic
- `internal/adapter`: TCP handler and protocol parser
- `internal/infrastructure`: Storage and persistence implementations
- `internal/container`: Dependency wiring (Wire)
- `cmd/server`: Server application entrypoint
- `cmd/client`: Simple CLI client for manual testing

## Prerequisites

- Go 1.26+ installed

## Getting Started

### 1. Run the server

```bash
go run ./cmd/server
```

Server flags:

- `-port` (default: `6379`)
- `-aof` (default: `false`)

Example with AOF enabled:

```bash
go run ./cmd/server -port 6379 -aof
```

### 2. Connect with the CLI client

In another terminal:

```bash
go run ./cmd/client localhost:6379
```

## Quick Command Demo

After connecting with the client:

```text
SET user:1 Alice
GET user:1
EXPIRE user:1 30
TTL user:1
EXISTS user:1 user:2
KEYS user:*
PING hello
INFO
QUIT
```

## Supported Commands

| Command | Example | Behavior |
| --- | --- | --- |
| `SET` | `SET key value` | Sets a key to a value |
| `GET` | `GET key` | Returns the value or `nil` |
| `DEL` | `DEL key1 key2` | Deletes keys and returns removed count |
| `EXPIRE` | `EXPIRE key 60` | Sets expiration in seconds |
| `TTL` | `TTL key` | Returns remaining seconds or `-1` |
| `PERSIST` | `PERSIST key` | Removes key expiration |
| `KEYS` | `KEYS user:*` | Returns matching keys |
| `EXISTS` | `EXISTS key1 key2` | Returns how many keys exist |
| `PING` | `PING` or `PING hello` | Returns `PONG` or custom message |
| `INFO` | `INFO` | Returns server stats summary |
| `QUIT` | `QUIT` | Closes the client connection |

## Persistence (AOF)

When `-aof` is enabled:

- Write commands are appended to an AOF file (`data.aof`)
- The server replays the file on startup to recover state

This persistence model is intentionally simple and focused on core recovery behavior.

## Running Tests

Run all tests:

```bash
go test ./...
```

The repository includes:

- Unit tests for protocol parsing and storage behavior
- Integration tests for end-to-end TCP command flows

## Current Scope and Notes

- Uses a plain text, line-based protocol (not full RESP compatibility)
- Intended as a Redis-like educational project, not a production Redis replacement
- Focused on string values and core command semantics

## Possible Next Improvements

- RESP protocol support for better client compatibility
- Additional data types (lists, sets, hashes)
- Authentication and access control
- Replication and clustering
- Richer metrics and observability
