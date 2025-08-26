package tangled

import (
	"fmt"
	"sort"
)

// Module represents a Go module with its path and version
type Module struct {
	Path    string
	Version string
}

// String returns the string representation of a module
func (m Module) String() string {
	if m.Version == "" {
		return m.Path
	}
	return fmt.Sprintf("%s@%s", m.Path, m.Version)
}

// Dependency represents a dependency relationship between two modules
type Dependency struct {
	From Module
	To   Module
}

// DependencyGraph represents the complete dependency graph
type DependencyGraph struct {
	MainModule   Module
	Dependencies []Dependency
	tree         map[string][]string // cached tree structure for visualization
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph(mainModule Module) *DependencyGraph {
	return &DependencyGraph{
		MainModule:   mainModule,
		Dependencies: make([]Dependency, 0),
		tree:         make(map[string][]string),
	}
}

// AddDependency adds a dependency to the graph
func (dg *DependencyGraph) AddDependency(from, to Module) {
	dg.Dependencies = append(dg.Dependencies, Dependency{From: from, To: to})
	// Invalidate cached tree
	dg.tree = make(map[string][]string)
}

// GetDirectDependencies returns all direct dependencies of a module
func (dg *DependencyGraph) GetDirectDependencies(module Module) []Module {
	var deps []Module
	moduleStr := module.String()

	for _, dep := range dg.Dependencies {
		if dep.From.String() == moduleStr {
			deps = append(deps, dep.To)
		}
	}
	return deps
}

// GetAllModules returns all unique modules in the graph
func (dg *DependencyGraph) GetAllModules() []Module {
	moduleSet := make(map[string]Module)
	moduleSet[dg.MainModule.String()] = dg.MainModule

	for _, dep := range dg.Dependencies {
		moduleSet[dep.From.String()] = dep.From
		moduleSet[dep.To.String()] = dep.To
	}

	var modules []Module
	for _, module := range moduleSet {
		modules = append(modules, module)
	}

	// Sort modules for deterministic ordering to ensure consistent node IDs
	// between generateNodes() and generateLinks() in HTML renderer
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].String() < modules[j].String()
	})

	return modules
}

// buildTree builds the tree structure for visualization
func (dg *DependencyGraph) buildTree() {
	if len(dg.tree) > 0 {
		return // Already built
	}

	for _, dep := range dg.Dependencies {
		fromStr := dep.From.String()
		toStr := dep.To.String()

		if dg.tree[fromStr] == nil {
			dg.tree[fromStr] = make([]string, 0)
		}
		dg.tree[fromStr] = append(dg.tree[fromStr], toStr)
	}
}

// GetTree returns the tree structure starting from the main module
func (dg *DependencyGraph) GetTree() map[string][]string {
	dg.buildTree()
	return dg.tree
}
