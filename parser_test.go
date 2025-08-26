package godepviewer

import (
	"strings"
	"testing"
)

func TestParseModule(t *testing.T) {
	tests := []struct {
		input    string
		expected Module
		wantErr  bool
	}{
		{
			input:    "github.com/example/module@v1.2.3",
			expected: Module{Path: "github.com/example/module", Version: "v1.2.3"},
			wantErr:  false,
		},
		{
			input:    "github.com/example/module",
			expected: Module{Path: "github.com/example/module", Version: ""},
			wantErr:  false,
		},
		{
			input:    "github.com/example@test/module@v1.2.3",
			expected: Module{Path: "github.com/example@test/module", Version: "v1.2.3"},
			wantErr:  false,
		},
		{
			input:   "",
			wantErr: true,
		},
		{
			input:   "@v1.2.3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseModule(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseModule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("parseModule() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseGraph(t *testing.T) {
	input := `github.com/example/main github.com/dep1@v1.0.0
github.com/example/main github.com/dep2@v2.0.0
github.com/dep1@v1.0.0 github.com/subdep@v1.0.0`

	reader := strings.NewReader(input)
	graph, err := ParseGraph(reader)
	if err != nil {
		t.Fatalf("ParseGraph() error = %v", err)
	}

	if graph.MainModule.Path != "github.com/example/main" {
		t.Errorf("MainModule.Path = %v, want %v", graph.MainModule.Path, "github.com/example/main")
	}

	if len(graph.Dependencies) != 3 {
		t.Errorf("Dependencies length = %v, want %v", len(graph.Dependencies), 3)
	}

	// Check first dependency
	dep := graph.Dependencies[0]
	if dep.From.Path != "github.com/example/main" {
		t.Errorf("First dependency From.Path = %v, want %v", dep.From.Path, "github.com/example/main")
	}
	if dep.To.Path != "github.com/dep1" || dep.To.Version != "v1.0.0" {
		t.Errorf("First dependency To = %v, want github.com/dep1@v1.0.0", dep.To)
	}
}

func TestParseGraphErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "invalid line format",
			input:   "github.com/example/main",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "empty module",
			input:   " github.com/dep@v1.0.0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			_, err := ParseGraph(reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGraph() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDependencyGraph_GetDirectDependencies(t *testing.T) {
	mainModule := Module{Path: "github.com/example/main", Version: ""}
	graph := NewDependencyGraph(mainModule)

	dep1 := Module{Path: "github.com/dep1", Version: "v1.0.0"}
	dep2 := Module{Path: "github.com/dep2", Version: "v2.0.0"}

	graph.AddDependency(mainModule, dep1)
	graph.AddDependency(mainModule, dep2)

	deps := graph.GetDirectDependencies(mainModule)
	if len(deps) != 2 {
		t.Errorf("GetDirectDependencies() returned %d dependencies, want 2", len(deps))
	}
}

func TestDependencyGraph_GetAllModules(t *testing.T) {
	mainModule := Module{Path: "github.com/example/main", Version: ""}
	graph := NewDependencyGraph(mainModule)

	dep1 := Module{Path: "github.com/dep1", Version: "v1.0.0"}
	dep2 := Module{Path: "github.com/dep2", Version: "v2.0.0"}

	graph.AddDependency(mainModule, dep1)
	graph.AddDependency(dep1, dep2)

	modules := graph.GetAllModules()
	if len(modules) != 3 {
		t.Errorf("GetAllModules() returned %d modules, want 3", len(modules))
	}
}

func TestModule_String(t *testing.T) {
	tests := []struct {
		module   Module
		expected string
	}{
		{
			module:   Module{Path: "github.com/example/module", Version: "v1.2.3"},
			expected: "github.com/example/module@v1.2.3",
		},
		{
			module:   Module{Path: "github.com/example/module", Version: ""},
			expected: "github.com/example/module",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.module.String(); got != tt.expected {
				t.Errorf("Module.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
