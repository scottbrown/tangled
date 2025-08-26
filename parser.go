package tangled

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// ParseError represents an error that occurred during parsing
type ParseError struct {
	Line    int
	Content string
	Err     error
}

func (pe ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d (%q): %v", pe.Line, pe.Content, pe.Err)
}

// parseModule parses a module string into a Module struct
func parseModule(moduleStr string) (Module, error) {
	if moduleStr == "" {
		return Module{}, fmt.Errorf("empty module string")
	}

	// Find the last @ symbol to separate path from version
	lastAt := strings.LastIndex(moduleStr, "@")
	if lastAt == -1 {
		// No version (main module)
		return Module{Path: moduleStr, Version: ""}, nil
	}

	path := moduleStr[:lastAt]
	version := moduleStr[lastAt+1:]

	if path == "" {
		return Module{}, fmt.Errorf("empty module path")
	}

	return Module{Path: path, Version: version}, nil
}

// ParseGraphFromFile parses a go mod graph file and returns a DependencyGraph
func ParseGraphFromFile(filename string) (*DependencyGraph, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return ParseGraph(file)
}

// ParseGraph parses go mod graph output from a reader and returns a DependencyGraph
func ParseGraph(reader io.Reader) (*DependencyGraph, error) {
	scanner := bufio.NewScanner(reader)
	var dependencies []Dependency
	lineNum := 0

	// First pass: collect all dependencies
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse the line: "from_module to_module"
		parts := strings.Fields(line)
		if len(parts) != 2 {
			return nil, ParseError{
				Line:    lineNum,
				Content: line,
				Err:     fmt.Errorf("expected 2 fields, got %d", len(parts)),
			}
		}

		fromModule, err := parseModule(parts[0])
		if err != nil {
			return nil, ParseError{
				Line:    lineNum,
				Content: line,
				Err:     fmt.Errorf("failed to parse from module: %w", err),
			}
		}

		toModule, err := parseModule(parts[1])
		if err != nil {
			return nil, ParseError{
				Line:    lineNum,
				Content: line,
				Err:     fmt.Errorf("failed to parse to module: %w", err),
			}
		}

		dependencies = append(dependencies, Dependency{From: fromModule, To: toModule})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	if len(dependencies) == 0 {
		return nil, fmt.Errorf("no dependencies found in input")
	}

	// Second pass: identify the main module
	// The main module is the one without a version that appears as a "from" dependency
	mainModule := identifyMainModule(dependencies)

	// Create graph with correct main module
	graph := NewDependencyGraph(mainModule)

	// Add all dependencies
	for _, dep := range dependencies {
		graph.AddDependency(dep.From, dep.To)
	}

	return graph, nil
}

// identifyMainModule identifies the main module from the dependencies
// The main module is typically the one without a version that appears as a "from" dependency
func identifyMainModule(dependencies []Dependency) Module {
	// Count frequency of modules as "from" dependencies and track modules without versions
	fromCounts := make(map[string]int)
	var modulesWithoutVersion []Module

	for _, dep := range dependencies {
		fromStr := dep.From.String()
		fromCounts[fromStr]++

		// Track modules without versions (potential main modules)
		if dep.From.Version == "" {
			// Check if we already have this module
			found := false
			for _, m := range modulesWithoutVersion {
				if m.Path == dep.From.Path {
					found = true
					break
				}
			}
			if !found {
				modulesWithoutVersion = append(modulesWithoutVersion, dep.From)
			}
		}
	}

	// If there's exactly one module without a version, that's likely the main module
	if len(modulesWithoutVersion) == 1 {
		return modulesWithoutVersion[0]
	}

	// If there are multiple or no modules without versions,
	// fallback to the most frequent "from" module
	var mainModule Module
	maxCount := 0

	for _, dep := range dependencies {
		if count := fromCounts[dep.From.String()]; count > maxCount {
			maxCount = count
			mainModule = dep.From
		}
	}

	return mainModule
}
