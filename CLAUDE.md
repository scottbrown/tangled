# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**tangled** is a Go module dependency graph visualization tool that parses `go mod graph` output and generates visualizations in multiple formats (text tree, interactive HTML/D3, MermaidJS, GraphViz DOT). The tool helps developers understand complex dependency relationships in Go projects.

## Development Commands

### Primary Workflow
```bash
task          # Full build pipeline (clean, format, lint, test, build)
task dev      # Quick development workflow (format, lint, test-short)
```

### Individual Tasks
```bash
task clean       # Clean artifacts and create directories
task format      # Format Go code
task lint        # Run go vet
task test        # Run tests with HTML coverage report (.test/coverage.html)
task test-short  # Quick tests without coverage  
task install     # Install binary to GOPATH/bin
task run         # Test run with example.graph
task sast        # Security analysis with gosec
task vuln        # Vulnerability scanning
```

### Single Test Execution
```bash
go test -run TestSpecificFunction ./...
go test -run TestParseModule ./...  # Example for parser tests
```

## Architecture Overview

### Core Components
- **CLI Layer** (`cmd/tangled/`): Cobra-based command-line interface handling input/output file management
- **Parser** (`parser.go`): Parses `go mod graph` output, builds internal dependency graph with error handling
- **Renderer Interface** (`renderer.go`): Strategy pattern for multiple output formats (TextRenderer, HTMLRenderer, MermaidRenderer, DotRenderer)
- **Types** (`types.go`): Core data structures (Module, Dependency, DependencyGraph) with tree building and caching

### Data Flow
1. CLI accepts graph file path and output format selection
2. Parser validates input and constructs `DependencyGraph` with dependency relationships
3. Appropriate renderer (selected by format flag) generates output
4. Result written to specified file or stdout

### Key Design Patterns
- **Interface-based rendering**: All renderers implement `Renderer` interface for extensibility
- **Builder pattern**: `DependencyGraph` construction and cached tree structure building
- **Error wrapping**: Custom `ParseError` type with line numbers and context
- **Dependency injection**: Renderers injected based on format selection

## Code Organization

### Package Structure
- Root package (`tangled`): Core business logic, follows Go convention of package name matching directory
- `cmd/tangled/`: CLI entry point and command definitions
- Minimal external dependencies (only Cobra for CLI)

### Testing Strategy
- Comprehensive test coverage with table-driven tests
- HTML coverage reports generated in `.test/coverage.html`
- Both unit tests (`parser_test.go`, `renderer_test.go`) and integration testing via `task run`

### Security Considerations
- Uses `#nosec G304` suppressions for justified CLI file operations (input/output files from user-provided paths)
- Regular vulnerability scanning via `task vuln`
- Static analysis via `task sast`

## Development Notes

### Adding New Output Formats
1. Implement the `Renderer` interface in `renderer.go`
2. Add format selection logic in `cmd/root.go`
3. Add corresponding tests following existing patterns

### Parser Modifications
- All parsing errors should use `ParseError` type with line context
- Graph building maintains cached tree structure for performance
- Follow existing dependency validation patterns

### File Operations
File operations use CLI-provided paths and include gosec suppressions with justification comments for legitimate use cases (user-specified input/output files).