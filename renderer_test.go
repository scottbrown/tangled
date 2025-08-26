package tangled

import (
	"bytes"
	"strings"
	"testing"
)

func createTestGraph() *DependencyGraph {
	mainModule := Module{Path: "github.com/example/main", Version: ""}
	graph := NewDependencyGraph(mainModule)

	dep1 := Module{Path: "github.com/dep1", Version: "v1.0.0"}
	dep2 := Module{Path: "github.com/dep2", Version: "v2.0.0"}
	subdep := Module{Path: "github.com/subdep", Version: "v1.0.0"}

	graph.AddDependency(mainModule, dep1)
	graph.AddDependency(mainModule, dep2)
	graph.AddDependency(dep1, subdep)

	return graph
}

func TestPlaintextRenderer_Render(t *testing.T) {
	graph := createTestGraph()
	renderer := NewPlaintextRenderer()

	var buf bytes.Buffer
	err := renderer.Render(graph, &buf)
	if err != nil {
		t.Fatalf("PlaintextRenderer.Render() error = %v", err)
	}

	output := buf.String()

	// Check that main module is at the root
	if !strings.Contains(output, "github.com/example/main") {
		t.Error("Output should contain main module")
	}

	// Check that dependencies are included
	if !strings.Contains(output, "github.com/dep1@v1.0.0") {
		t.Error("Output should contain dep1")
	}

	if !strings.Contains(output, "github.com/dep2@v2.0.0") {
		t.Error("Output should contain dep2")
	}

	// Check tree structure characters
	if !strings.Contains(output, "├──") || !strings.Contains(output, "└──") {
		t.Error("Output should contain tree structure characters")
	}
}

func TestMermaidRenderer_Render(t *testing.T) {
	graph := createTestGraph()
	renderer := NewMermaidRenderer()

	var buf bytes.Buffer
	err := renderer.Render(graph, &buf)
	if err != nil {
		t.Fatalf("MermaidRenderer.Render() error = %v", err)
	}

	output := buf.String()

	// Check that it starts with the correct Mermaid syntax
	if !strings.HasPrefix(output, "graph TD") {
		t.Error("Output should start with 'graph TD'")
	}

	// Check that nodes are defined
	if !strings.Contains(output, `["github.com/example/main"]`) {
		t.Error("Output should contain main module node")
	}

	// Check that edges are defined
	if !strings.Contains(output, "-->") {
		t.Error("Output should contain arrows")
	}
}

func TestGraphvizRenderer_Render(t *testing.T) {
	graph := createTestGraph()
	renderer := NewGraphvizRenderer()

	var buf bytes.Buffer
	err := renderer.Render(graph, &buf)
	if err != nil {
		t.Fatalf("GraphvizRenderer.Render() error = %v", err)
	}

	output := buf.String()

	// Check that it starts and ends correctly
	if !strings.HasPrefix(output, "digraph dependencies {") {
		t.Error("Output should start with 'digraph dependencies {'")
	}

	if !strings.HasSuffix(strings.TrimSpace(output), "}") {
		t.Error("Output should end with '}'")
	}

	// Check that main module is highlighted
	if !strings.Contains(output, "fillcolor=lightblue") {
		t.Error("Output should highlight main module")
	}

	// Check that edges are defined
	if !strings.Contains(output, "->") {
		t.Error("Output should contain arrows")
	}
}

func TestHTMLRenderer_Render(t *testing.T) {
	graph := createTestGraph()
	renderer := NewHTMLRenderer()

	var buf bytes.Buffer
	err := renderer.Render(graph, &buf)
	if err != nil {
		t.Fatalf("HTMLRenderer.Render() error = %v", err)
	}

	output := buf.String()

	// Check that it's valid HTML
	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("Output should be valid HTML")
	}

	if !strings.Contains(output, "<html>") {
		t.Error("Output should contain html tag")
	}

	// Check that D3.js is included
	if !strings.Contains(output, "d3js.org/d3") {
		t.Error("Output should include D3.js")
	}

	// Check that nodes are included in the JavaScript
	if !strings.Contains(output, "github.com/example/main") {
		t.Error("Output should contain module names in JavaScript")
	}
}

func TestGraphvizRenderer_sanitizeNodeID(t *testing.T) {
	renderer := NewGraphvizRenderer()

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "github.com/example/module@v1.2.3",
			expected: "github_com_example_module_v1_2_3",
		},
		{
			input:    "simple-module",
			expected: "simple_module",
		},
		{
			input:    "module.with.dots",
			expected: "module_with_dots",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := renderer.sanitizeNodeID(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeNodeID() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHTMLRenderer_generateNodes(t *testing.T) {
	graph := createTestGraph()
	renderer := NewHTMLRenderer()

	nodes := renderer.generateNodes(graph)

	// Check that nodes are in JSON format
	if !strings.HasPrefix(nodes, "[") || !strings.HasSuffix(nodes, "]") {
		t.Error("Nodes should be in JSON array format")
	}

	// Check that main module has different group
	if !strings.Contains(nodes, `"group": 2`) {
		t.Error("Main module should have group 2")
	}

	// Check that regular modules have group 1
	if !strings.Contains(nodes, `"group": 1`) {
		t.Error("Regular modules should have group 1")
	}
}

func TestHTMLRenderer_generateLinks(t *testing.T) {
	graph := createTestGraph()
	renderer := NewHTMLRenderer()

	links := renderer.generateLinks(graph)

	// Check that links are in JSON format
	if !strings.HasPrefix(links, "[") || !strings.HasSuffix(links, "]") {
		t.Error("Links should be in JSON array format")
	}

	// Check that source and target indices are present
	if !strings.Contains(links, `"source":`) || !strings.Contains(links, `"target":`) {
		t.Error("Links should contain source and target indices")
	}
}

func TestRendererInterfaces(t *testing.T) {
	// Test that all renderers implement the Renderer interface
	var _ Renderer = &PlaintextRenderer{}
	var _ Renderer = &MermaidRenderer{}
	var _ Renderer = &GraphvizRenderer{}
	var _ Renderer = &HTMLRenderer{}
}
