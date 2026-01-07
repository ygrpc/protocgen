# Project Context

## Purpose

`protocgen` is a collection of `protoc` compiler plugins for the `ygrpc` ecosystem. Its goal is to enable a low-code development workflow where Protobuf IDL serves as the single source of truth. It automates the generation of:

- SQL initialization scripts (`protoc-gen-ygrpc-sql`)
- Auxiliary Protobuf messages like lists and CRUD RPC definitions (`protoc-gen-ygrpc-msglist`)
- Database interaction code and other boilerplate (`protoc-gen-ygrpc-protodb`)

The philosophy is to reduce repetitive work while keeping the underlying process transparent and understandable.

## Tech Stack

- **Languages**: Go (1.23+), Protobuf
- **Core Dependencies**:
  - `google.golang.org/protobuf`
  - `github.com/ygrpc/protodb`
- **Generators**: Custom `protoc` plugins
- **Target Outputs**: SQL, Go, TypeScript (for frontend integration)

## Project Conventions

### Code Style

- **Go**: Follows standard Go idioms and formatting (`gofmt`, `goimports`).
- **Protobuf**: Snake_case for file names, PascalCase for message names.

### Architecture Patterns

- **Plugin Architecture**: Each tool in `cmd/` acts as a standalone `protoc` plugin, reading `CodeGeneratorRequest` from stdin and writing `CodeGeneratorResponse` to stdout.
- **Single Source of Truth**: All business logic starts from definitions in `.proto` files.

### Testing Strategy

- Manual verification of generated output.
- Integration tests by running plugins against sample `.proto` files and verifying the artifacts.

### Git Workflow

- Standard feature branches.
- Commit messages should be descriptive.

## Domain Context

- **ygrpc**: A low-code system/framework.
- **Flow**:
  1. Define DB schema and services in `.proto`.
  2. Generate SQL to init DB.
  3. Generate Message Lists/RPC helpers.
  4. Generate CRUD code (Go backend, TS frontend).

## Important Constraints

- **Protoc Compatibility**: Must work with standard `protoc` compiler.
- **Idempotency**: Generated code should be deterministic.

## External Dependencies

- **Protobuf Compiler (`protoc`)**: Required to run the plugins.
