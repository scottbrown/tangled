package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/scottbrown/tangled"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	outputFile   string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tangled [graph-file]",
	Short: "Visualize Go module dependency graphs",
	Long: `tangled parses the output from 'go mod graph' and generates
various visualization formats including plaintext tree, HTML/D3, MermaidJS, and GraphViz DOT.

Example usage:
  go mod graph > deps.graph
  tangled deps.graph
  tangled -f html -o deps.html deps.graph
  tangled -f mermaid -o deps.mmd deps.graph`,
	Args:    cobra.ExactArgs(1),
	Version: tangled.Version(),
	RunE:    runRoot,
}

func runRoot(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	// Parse the dependency graph
	graph, err := tangled.ParseGraphFromFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to parse graph file: %w", err)
	}

	// Create the appropriate renderer
	var renderer tangled.Renderer
	switch strings.ToLower(outputFormat) {
	case "text", "plaintext", "tree":
		renderer = tangled.NewPlaintextRenderer()
	case "html", "d3":
		renderer = tangled.NewHTMLRenderer()
	case "mermaid", "mmd":
		renderer = tangled.NewMermaidRenderer()
	case "dot", "graphviz":
		renderer = tangled.NewGraphvizRenderer()
	default:
		return fmt.Errorf("unsupported output format: %s (supported: text, html, mermaid, dot)", outputFormat)
	}

	// Determine output destination
	var writer *os.File
	if outputFile == "" || outputFile == "-" {
		writer = os.Stdout
	} else {
		file, err := os.Create(outputFile) // #nosec G304 -- CLI tool, output file from user-provided command line flag
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		writer = file
	}

	// Render the graph
	if htmlRenderer, ok := renderer.(*tangled.HTMLRenderer); ok {
		// For HTML renderer, pass the filename for dynamic title
		filename := filepath.Base(inputFile)
		if err := htmlRenderer.RenderWithFilename(graph, writer, filename); err != nil {
			return fmt.Errorf("failed to render graph: %w", err)
		}
	} else {
		// For other renderers, use the standard render method
		if err := renderer.Render(graph, writer); err != nil {
			return fmt.Errorf("failed to render graph: %w", err)
		}
	}

	// Print success message to stderr if outputting to file
	if outputFile != "" && outputFile != "-" {
		fmt.Fprintf(os.Stderr, "Successfully generated %s output in %s\n", outputFormat, outputFile)
	}

	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "Output format (text, html, mermaid, dot)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
}
