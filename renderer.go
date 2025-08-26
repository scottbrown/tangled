package godepviewer

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// Renderer interface for different output formats
type Renderer interface {
	Render(graph *DependencyGraph, writer io.Writer) error
}

// PlaintextRenderer renders the dependency graph as plaintext tree
type PlaintextRenderer struct{}

// NewPlaintextRenderer creates a new plaintext renderer
func NewPlaintextRenderer() *PlaintextRenderer {
	return &PlaintextRenderer{}
}

// Render renders the dependency graph as a plaintext tree
func (r *PlaintextRenderer) Render(graph *DependencyGraph, writer io.Writer) error {
	visited := make(map[string]bool)
	return r.renderNode(graph, graph.MainModule.String(), "", true, visited, writer)
}

func (r *PlaintextRenderer) renderNode(graph *DependencyGraph, nodeKey string, prefix string, isLast bool, visited map[string]bool, writer io.Writer) error {
	// Print current node
	var connector string
	if prefix == "" {
		connector = ""
	} else if isLast {
		connector = "└── "
	} else {
		connector = "├── "
	}

	_, err := fmt.Fprintf(writer, "%s%s%s\n", prefix, connector, nodeKey)
	if err != nil {
		return err
	}

	// Avoid infinite recursion by tracking visited nodes
	if visited[nodeKey] {
		return nil
	}
	visited[nodeKey] = true

	// Get dependencies for this node
	tree := graph.GetTree()
	dependencies := tree[nodeKey]

	// Sort dependencies for consistent output
	sort.Strings(dependencies)

	// Calculate new prefix for children
	var newPrefix string
	if prefix == "" {
		newPrefix = "  " // Start indentation for children of root
	} else if isLast {
		newPrefix = prefix + "    "
	} else {
		newPrefix = prefix + "│   "
	}

	// Render children
	for i, dep := range dependencies {
		isLastChild := i == len(dependencies)-1
		err := r.renderNode(graph, dep, newPrefix, isLastChild, visited, writer)
		if err != nil {
			return err
		}
	}

	return nil
}

// MermaidRenderer renders the dependency graph as MermaidJS format
type MermaidRenderer struct{}

// NewMermaidRenderer creates a new MermaidJS renderer
func NewMermaidRenderer() *MermaidRenderer {
	return &MermaidRenderer{}
}

// Render renders the dependency graph as MermaidJS format
func (r *MermaidRenderer) Render(graph *DependencyGraph, writer io.Writer) error {
	_, err := fmt.Fprintln(writer, "graph TD")
	if err != nil {
		return err
	}

	// Generate unique IDs for nodes
	nodeIDs := make(map[string]string)
	idCounter := 1

	for _, module := range graph.GetAllModules() {
		moduleStr := module.String()
		nodeIDs[moduleStr] = fmt.Sprintf("N%d", idCounter)
		idCounter++
	}

	// Render node definitions
	for moduleStr, nodeID := range nodeIDs {
		escapedLabel := strings.ReplaceAll(moduleStr, `"`, `\"`)
		_, err := fmt.Fprintf(writer, "    %s[\"%s\"]\n", nodeID, escapedLabel)
		if err != nil {
			return err
		}
	}

	// Render edges
	for _, dep := range graph.Dependencies {
		fromID := nodeIDs[dep.From.String()]
		toID := nodeIDs[dep.To.String()]
		_, err := fmt.Fprintf(writer, "    %s --> %s\n", fromID, toID)
		if err != nil {
			return err
		}
	}

	return nil
}

// GraphvizRenderer renders the dependency graph as GraphViz DOT format
type GraphvizRenderer struct{}

// NewGraphvizRenderer creates a new GraphViz renderer
func NewGraphvizRenderer() *GraphvizRenderer {
	return &GraphvizRenderer{}
}

// Render renders the dependency graph as GraphViz DOT format
func (r *GraphvizRenderer) Render(graph *DependencyGraph, writer io.Writer) error {
	_, err := fmt.Fprintln(writer, "digraph dependencies {")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(writer, "    rankdir=LR;")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(writer, "    node [shape=box, style=rounded];")
	if err != nil {
		return err
	}

	// Render nodes
	modules := graph.GetAllModules()
	for _, module := range modules {
		moduleStr := module.String()
		escapedLabel := strings.ReplaceAll(moduleStr, `"`, `\"`)
		nodeID := r.sanitizeNodeID(moduleStr)

		// Highlight main module
		if moduleStr == graph.MainModule.String() {
			_, err = fmt.Fprintf(writer, "    \"%s\" [label=\"%s\", fillcolor=lightblue, style=\"rounded,filled\"];\n", nodeID, escapedLabel)
		} else {
			_, err = fmt.Fprintf(writer, "    \"%s\" [label=\"%s\"];\n", nodeID, escapedLabel)
		}
		if err != nil {
			return err
		}
	}

	// Render edges
	for _, dep := range graph.Dependencies {
		fromID := r.sanitizeNodeID(dep.From.String())
		toID := r.sanitizeNodeID(dep.To.String())
		_, err := fmt.Fprintf(writer, "    \"%s\" -> \"%s\";\n", fromID, toID)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintln(writer, "}")
	return err
}

func (r *GraphvizRenderer) sanitizeNodeID(nodeID string) string {
	// Replace problematic characters for DOT format
	sanitized := strings.ReplaceAll(nodeID, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, ".", "_")
	sanitized = strings.ReplaceAll(sanitized, "@", "_")
	sanitized = strings.ReplaceAll(sanitized, "-", "_")
	return sanitized
}

// HTMLRenderer renders the dependency graph as HTML with D3.js
type HTMLRenderer struct{}

// NewHTMLRenderer creates a new HTML renderer
func NewHTMLRenderer() *HTMLRenderer {
	return &HTMLRenderer{}
}

// Render renders the dependency graph as HTML with D3.js visualization
func (r *HTMLRenderer) Render(graph *DependencyGraph, writer io.Writer) error {
	template := r.getHTMLTemplate()

	// Generate nodes and links for D3
	nodes := r.generateNodes(graph)
	links := r.generateLinks(graph)

	// Replace placeholders in template
	html := strings.ReplaceAll(template, "{{NODES}}", nodes)
	html = strings.ReplaceAll(html, "{{LINKS}}", links)

	_, err := writer.Write([]byte(html))
	return err
}

func (r *HTMLRenderer) generateNodes(graph *DependencyGraph) string {
	var nodes []string
	modules := graph.GetAllModules()

	for i, module := range modules {
		moduleStr := module.String()
		escapedLabel := strings.ReplaceAll(moduleStr, `"`, `\"`)
		escapedLabel = strings.ReplaceAll(escapedLabel, `\`, `\\`)

		// Mark main module differently
		group := 1
		if moduleStr == graph.MainModule.String() {
			group = 2
		}

		node := fmt.Sprintf(`{"id": %d, "name": "%s", "group": %d}`, i, escapedLabel, group)
		nodes = append(nodes, node)
	}

	return "[" + strings.Join(nodes, ",\n        ") + "]"
}

func (r *HTMLRenderer) generateLinks(graph *DependencyGraph) string {
	var links []string
	modules := graph.GetAllModules()

	// Create module to index mapping
	moduleToIndex := make(map[string]int)
	for i, module := range modules {
		moduleToIndex[module.String()] = i
	}

	for _, dep := range graph.Dependencies {
		fromIndex := moduleToIndex[dep.From.String()]
		toIndex := moduleToIndex[dep.To.String()]

		link := fmt.Sprintf(`{"source": %d, "target": %d}`, fromIndex, toIndex)
		links = append(links, link)
	}

	return "[" + strings.Join(links, ",\n        ") + "]"
}

func (r *HTMLRenderer) getHTMLTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>Go Dependency Graph</title>
    <script src="https://d3js.org/d3.v7.min.js"></script>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
        }
        .node {
            stroke: #fff;
            stroke-width: 1.5px;
            cursor: pointer;
        }
        .link {
            stroke: #999;
            stroke-opacity: 0.6;
            marker-end: url(#arrowhead);
        }
        .node text {
            font-size: 12px;
            text-anchor: middle;
            pointer-events: none;
        }
        #tooltip {
            position: absolute;
            padding: 8px;
            background: rgba(0, 0, 0, 0.8);
            color: white;
            border-radius: 4px;
            pointer-events: none;
            opacity: 0;
        }
        .zoom-controls {
            position: absolute;
            top: 80px;
            left: 30px;
            background: rgba(255, 255, 255, 0.9);
            border: 1px solid #ccc;
            border-radius: 4px;
            padding: 10px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        }
        .zoom-button {
            display: block;
            width: 40px;
            height: 40px;
            margin: 5px 0;
            border: 1px solid #999;
            background: #fff;
            border-radius: 4px;
            cursor: pointer;
            font-size: 18px;
            font-weight: bold;
            text-align: center;
            line-height: 38px;
            user-select: none;
        }
        .zoom-button:hover {
            background: #f0f0f0;
        }
        .zoom-button:active {
            background: #e0e0e0;
        }
        #graph-container {
            position: relative;
            border: 1px solid #ddd;
            border-radius: 4px;
            overflow: hidden;
        }
        .minimap {
            position: absolute;
            bottom: 20px;
            right: 20px;
            width: 200px;
            height: 150px;
            border: 2px solid #666;
            border-radius: 4px;
            background: rgba(255, 255, 255, 0.95);
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
            cursor: pointer;
        }
        .minimap svg {
            width: 100%;
            height: 100%;
        }
        .minimap .minimap-node {
            fill: #4ecdc4;
            stroke: none;
        }
        .minimap .minimap-node.main {
            fill: #ff6b6b;
        }
        .minimap .minimap-link {
            stroke: #999;
            stroke-width: 0.5px;
            stroke-opacity: 0.3;
        }
        .minimap .viewport {
            fill: rgba(0, 100, 200, 0.2);
            stroke: #0064c8;
            stroke-width: 2px;
            pointer-events: none;
        }
        .breadcrumb-container {
            position: absolute;
            top: 60px;
            left: 20px;
            right: 240px;
            height: 40px;
            background: rgba(255, 255, 255, 0.95);
            border: 1px solid #ddd;
            border-radius: 4px;
            padding: 8px 12px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
            overflow-x: auto;
            white-space: nowrap;
            font-size: 14px;
            line-height: 24px;
        }
        .breadcrumb {
            display: inline-block;
            color: #666;
        }
        .breadcrumb-item {
            display: inline-block;
            max-width: 200px;
            overflow: hidden;
            text-overflow: ellipsis;
            vertical-align: top;
            color: #0066cc;
            cursor: pointer;
            padding: 2px 6px;
            border-radius: 3px;
            transition: background-color 0.2s;
        }
        .breadcrumb-item:hover {
            background-color: #f0f8ff;
            text-decoration: underline;
        }
        .breadcrumb-item.root {
            color: #cc0000;
            font-weight: bold;
        }
        .breadcrumb-item.current {
            color: #333;
            background-color: #e6f3ff;
            cursor: default;
        }
        .breadcrumb-item.current:hover {
            text-decoration: none;
        }
        .breadcrumb-separator {
            margin: 0 8px;
            color: #999;
        }
        .breadcrumb-empty {
            color: #999;
            font-style: italic;
        }
    </style>
</head>
<body>
    <h1>Go Dependency Graph</h1>
    <div id="graph-container">
        <div class="breadcrumb-container">
            <div class="breadcrumb" id="breadcrumb">
                <span class="breadcrumb-empty">Click a node to see its dependency path</span>
            </div>
        </div>
        <div class="zoom-controls">
            <button class="zoom-button" id="zoom-in">+</button>
            <button class="zoom-button" id="zoom-out">−</button>
            <button class="zoom-button" id="reset-zoom" style="font-size: 14px;">⌂</button>
        </div>
        <div id="graph"></div>
        <div class="minimap" id="minimap"></div>
    </div>
    <div id="tooltip"></div>

    <script>
        const width = 1200;
        const height = 800;

        const nodes = {{NODES}};
        const links = {{LINKS}};

        const svg = d3.select("#graph")
            .append("svg")
            .attr("width", width)
            .attr("height", height);

        // Create a container group for zoom/pan transformations
        const g = svg.append("g");

        // Define zoom behavior
        const zoom = d3.zoom()
            .scaleExtent([0.1, 10])
            .on("zoom", function(event) {
                g.attr("transform", event.transform);
            });

        // Apply zoom behavior to SVG
        svg.call(zoom);

        // Define arrow marker
        svg.append("defs").append("marker")
            .attr("id", "arrowhead")
            .attr("viewBox", "0 -5 10 10")
            .attr("refX", 15)
            .attr("refY", 0)
            .attr("markerWidth", 6)
            .attr("markerHeight", 6)
            .attr("orient", "auto")
            .append("path")
            .attr("d", "M0,-5L10,0L0,5")
            .attr("fill", "#999");

        const simulation = d3.forceSimulation(nodes)
            .force("link", d3.forceLink(links).id(d => d.id).distance(100))
            .force("charge", d3.forceManyBody().strength(-300))
            .force("center", d3.forceCenter(width / 2, height / 2));

        const link = g.append("g")
            .selectAll("line")
            .data(links)
            .join("line")
            .attr("class", "link");

        const node = g.append("g")
            .selectAll("circle")
            .data(nodes)
            .join("circle")
            .attr("class", "node")
            .attr("r", 8)
            .attr("fill", d => d.group === 2 ? "#ff6b6b" : "#4ecdc4")
            .call(d3.drag()
                .on("start", dragstarted)
                .on("drag", dragged)
                .on("end", dragended));

        const tooltip = d3.select("#tooltip");

        node.on("mouseover", function(event, d) {
            tooltip.style("opacity", 1)
                .style("left", (event.pageX + 10) + "px")
                .style("top", (event.pageY - 10) + "px")
                .text(d.name);
        })
        .on("mouseout", function() {
            tooltip.style("opacity", 0);
        })
        .on("click", function(event, d) {
            event.stopPropagation();
            selectedNode = d;
            updateBreadcrumb(d);
            highlightPath(d);
        });

        simulation.on("tick", () => {
            link
                .attr("x1", d => d.source.x)
                .attr("y1", d => d.source.y)
                .attr("x2", d => d.target.x)
                .attr("y2", d => d.target.y);

            node
                .attr("cx", d => d.x)
                .attr("cy", d => d.y);
        });

        function dragstarted(event, d) {
            if (!event.active) simulation.alphaTarget(0.3).restart();
            d.fx = d.x;
            d.fy = d.y;
        }

        function dragged(event, d) {
            d.fx = event.x;
            d.fy = event.y;
        }

        function dragended(event, d) {
            if (!event.active) simulation.alphaTarget(0);
            d.fx = null;
            d.fy = null;
        }

        // Zoom control functions
        function zoomIn() {
            svg.transition().duration(300).call(
                zoom.scaleBy, 1.5
            );
        }

        function zoomOut() {
            svg.transition().duration(300).call(
                zoom.scaleBy, 1 / 1.5
            );
        }

        function resetZoom() {
            svg.transition().duration(500).call(
                zoom.transform,
                d3.zoomIdentity
            );
        }

        // Add event listeners to zoom buttons
        d3.select("#zoom-in").on("click", zoomIn);
        d3.select("#zoom-out").on("click", zoomOut);
        d3.select("#reset-zoom").on("click", resetZoom);

        // Prevent zoom controls from interfering with drag
        d3.selectAll(".zoom-controls").on("mousedown", function(event) {
            event.stopPropagation();
        });

        // Minimap implementation
        const minimapWidth = 200;
        const minimapHeight = 150;
        
        const minimapSvg = d3.select("#minimap")
            .append("svg")
            .attr("width", minimapWidth)
            .attr("height", minimapHeight);

        const minimapG = minimapSvg.append("g");

        // Calculate bounds of all nodes
        function calculateGraphBounds() {
            if (nodes.length === 0) return { minX: 0, minY: 0, maxX: width, maxY: height };
            
            let minX = d3.min(nodes, d => d.x || 0);
            let maxX = d3.max(nodes, d => d.x || 0);
            let minY = d3.min(nodes, d => d.y || 0);
            let maxY = d3.max(nodes, d => d.y || 0);
            
            // Add padding
            const padding = 50;
            minX -= padding;
            maxX += padding;
            minY -= padding;
            maxY += padding;
            
            return { minX, minY, maxX, maxY };
        }

        // Create minimap scale functions
        let minimapScaleX, minimapScaleY;
        
        function updateMinimapScales() {
            const bounds = calculateGraphBounds();
            const graphWidth = bounds.maxX - bounds.minX;
            const graphHeight = bounds.maxY - bounds.minY;
            
            minimapScaleX = d3.scaleLinear()
                .domain([bounds.minX, bounds.maxX])
                .range([0, minimapWidth]);
                
            minimapScaleY = d3.scaleLinear()
                .domain([bounds.minY, bounds.maxY])
                .range([0, minimapHeight]);
        }

        // Create minimap elements
        const minimapLinks = minimapG.selectAll(".minimap-link")
            .data(links)
            .join("line")
            .attr("class", "minimap-link");

        const minimapNodes = minimapG.selectAll(".minimap-node")
            .data(nodes)
            .join("circle")
            .attr("class", d => d.group === 2 ? "minimap-node main" : "minimap-node")
            .attr("r", 1.5);

        // Viewport indicator
        const viewport = minimapSvg.append("rect")
            .attr("class", "viewport");

        // Update minimap positions
        function updateMinimap() {
            updateMinimapScales();
            
            minimapLinks
                .attr("x1", d => minimapScaleX(d.source.x))
                .attr("y1", d => minimapScaleY(d.source.y))
                .attr("x2", d => minimapScaleX(d.target.x))
                .attr("y2", d => minimapScaleY(d.target.y));

            minimapNodes
                .attr("cx", d => minimapScaleX(d.x))
                .attr("cy", d => minimapScaleY(d.y));
                
            updateViewportIndicator();
        }

        // Update viewport indicator based on current zoom/pan
        function updateViewportIndicator() {
            if (!minimapScaleX || !minimapScaleY) return;
            
            const transform = d3.zoomTransform(svg.node());
            const bounds = calculateGraphBounds();
            
            // Calculate visible area in graph coordinates
            const visibleLeft = (-transform.x) / transform.k;
            const visibleTop = (-transform.y) / transform.k;
            const visibleRight = visibleLeft + width / transform.k;
            const visibleBottom = visibleTop + height / transform.k;
            
            // Convert to minimap coordinates
            const minimapLeft = minimapScaleX(visibleLeft);
            const minimapTop = minimapScaleY(visibleTop);
            const minimapRight = minimapScaleX(visibleRight);
            const minimapBottom = minimapScaleY(visibleBottom);
            
            viewport
                .attr("x", Math.max(0, minimapLeft))
                .attr("y", Math.max(0, minimapTop))
                .attr("width", Math.max(0, Math.min(minimapWidth, minimapRight) - Math.max(0, minimapLeft)))
                .attr("height", Math.max(0, Math.min(minimapHeight, minimapBottom) - Math.max(0, minimapTop)));
        }

        // Update minimap on simulation tick
        simulation.on("tick", () => {
            link
                .attr("x1", d => d.source.x)
                .attr("y1", d => d.source.y)
                .attr("x2", d => d.target.x)
                .attr("y2", d => d.target.y);

            node
                .attr("cx", d => d.x)
                .attr("cy", d => d.y);
                
            updateMinimap();
        });

        // Update viewport indicator on zoom
        zoom.on("zoom", function(event) {
            g.attr("transform", event.transform);
            updateViewportIndicator();
        });

        // Minimap click interaction - pan to clicked location
        minimapSvg.on("click", function(event) {
            if (!minimapScaleX || !minimapScaleY) return;
            
            const [mouseX, mouseY] = d3.pointer(event);
            
            // Convert minimap coordinates back to graph coordinates
            const graphX = minimapScaleX.invert(mouseX);
            const graphY = minimapScaleY.invert(mouseY);
            
            // Center the main view on this point
            const transform = d3.zoomTransform(svg.node());
            const newX = width / 2 - graphX * transform.k;
            const newY = height / 2 - graphY * transform.k;
            
            svg.transition().duration(500).call(
                zoom.transform,
                d3.zoomIdentity.translate(newX, newY).scale(transform.k)
            );
        });

        // Prevent minimap from interfering with main graph interactions
        d3.select("#minimap").on("mousedown", function(event) {
            event.stopPropagation();
        });

        // Breadcrumb trail implementation
        const mainModuleName = nodes.find(n => n.group === 2)?.name || nodes[0]?.name;
        let selectedNode = null;

        // Build adjacency list for path finding (reversed - from dependencies to dependents)
        const dependents = new Map();
        links.forEach(link => {
            const fromId = link.source.id || link.source;
            const toId = link.target.id || link.target;
            
            if (!dependents.has(toId)) {
                dependents.set(toId, []);
            }
            dependents.get(toId).push(fromId);
        });

        // Find path from main module to target node using BFS
        function findPathToNode(targetNodeId) {
            if (!mainModuleName) return [];
            
            const mainNode = nodes.find(n => n.name === mainModuleName);
            if (!mainNode || mainNode.id === targetNodeId) return [mainNode];
            
            const queue = [{node: mainNode, path: [mainNode]}];
            const visited = new Set([mainNode.id]);
            
            while (queue.length > 0) {
                const {node, path} = queue.shift();
                
                // Get direct dependencies of current node
                const nodeDependencies = links
                    .filter(link => {
                        const sourceId = link.source.id || link.source;
                        return sourceId === node.id;
                    })
                    .map(link => {
                        const targetId = link.target.id || link.target;
                        return nodes.find(n => n.id === targetId);
                    })
                    .filter(n => n);
                
                for (const depNode of nodeDependencies) {
                    if (depNode.id === targetNodeId) {
                        return [...path, depNode];
                    }
                    
                    if (!visited.has(depNode.id)) {
                        visited.add(depNode.id);
                        queue.push({node: depNode, path: [...path, depNode]});
                    }
                }
            }
            
            return []; // No path found
        }

        // Update breadcrumb display
        function updateBreadcrumb(targetNode) {
            const breadcrumbEl = d3.select("#breadcrumb");
            
            if (!targetNode) {
                breadcrumbEl.html('<span class="breadcrumb-empty">Click a node to see its dependency path</span>');
                return;
            }
            
            const path = findPathToNode(targetNode.id);
            
            if (path.length === 0) {
                breadcrumbEl.html(
                    '<span class="breadcrumb-empty">No dependency path found to:</span>' +
                    '<span class="breadcrumb-item current">' + truncateModuleName(targetNode.name) + '</span>'
                );
                return;
            }
            
            let html = '';
            path.forEach((node, index) => {
                const isRoot = node.group === 2;
                const isCurrent = index === path.length - 1;
                const classes = ['breadcrumb-item'];
                
                if (isRoot) classes.push('root');
                if (isCurrent) classes.push('current');
                
                const truncatedName = truncateModuleName(node.name);
                
                if (index > 0) {
                    html += '<span class="breadcrumb-separator">→</span>';
                }
                
                html += '<span class="' + classes.join(' ') + '" data-node-id="' + node.id + '" title="' + node.name + '">' + truncatedName + '</span>';
            });
            
            breadcrumbEl.html(html);
        }

        // Truncate long module names for display
        function truncateModuleName(name) {
            if (name.length <= 40) return name;
            
            const parts = name.split('/');
            if (parts.length > 2) {
                return parts[0] + '/.../' + parts[parts.length - 1];
            }
            
            return name.substring(0, 37) + '...';
        }

        // Highlight path in the graph
        function highlightPath(targetNode) {
            // Reset all highlighting
            node.attr("stroke", "#fff").attr("stroke-width", 1.5);
            link.attr("stroke", "#999").attr("stroke-opacity", 0.6);
            
            if (!targetNode) return;
            
            const path = findPathToNode(targetNode.id);
            if (path.length === 0) return;
            
            // Highlight path nodes
            const pathNodeIds = new Set(path.map(n => n.id));
            node.attr("stroke", d => pathNodeIds.has(d.id) ? "#ff6600" : "#fff")
                .attr("stroke-width", d => pathNodeIds.has(d.id) ? 3 : 1.5);
            
            // Highlight path links
            const pathLinks = [];
            for (let i = 0; i < path.length - 1; i++) {
                const fromId = path[i].id;
                const toId = path[i + 1].id;
                
                const pathLink = links.find(link => {
                    const sourceId = link.source.id || link.source;
                    const targetId = link.target.id || link.target;
                    return sourceId === fromId && targetId === toId;
                });
                
                if (pathLink) pathLinks.push(pathLink);
            }
            
            link.attr("stroke", d => {
                const sourceId = d.source.id || d.source;
                const targetId = d.target.id || d.target;
                return pathLinks.some(pl => {
                    const plSourceId = pl.source.id || pl.source;
                    const plTargetId = pl.target.id || pl.target;
                    return plSourceId === sourceId && plTargetId === targetId;
                }) ? "#ff6600" : "#999";
            })
            .attr("stroke-opacity", d => {
                const sourceId = d.source.id || d.source;
                const targetId = d.target.id || d.target;
                return pathLinks.some(pl => {
                    const plSourceId = pl.source.id || pl.source;
                    const plTargetId = pl.target.id || pl.target;
                    return plSourceId === sourceId && plTargetId === targetId;
                }) ? 1 : 0.6;
            });
        }

        // Add breadcrumb click navigation
        d3.select("#breadcrumb").on("click", function(event) {
            const target = event.target;
            if (target.classList.contains("breadcrumb-item") && !target.classList.contains("current")) {
                const nodeId = parseInt(target.getAttribute("data-node-id"));
                const node = nodes.find(n => n.id === nodeId);
                if (node) {
                    selectedNode = node;
                    updateBreadcrumb(node);
                    highlightPath(node);
                    
                    // Center view on selected node
                    const transform = d3.zoomTransform(svg.node());
                    const newX = width / 2 - node.x * transform.k;
                    const newY = height / 2 - node.y * transform.k;
                    
                    svg.transition().duration(500).call(
                        zoom.transform,
                        d3.zoomIdentity.translate(newX, newY).scale(transform.k)
                    );
                }
            }
        });

        // Clear selection when clicking on empty space
        svg.on("click", function(event) {
            if (event.target === this) {
                selectedNode = null;
                updateBreadcrumb(null);
                highlightPath(null);
            }
        });

        // Prevent breadcrumb container from interfering with interactions
        d3.select(".breadcrumb-container").on("mousedown", function(event) {
            event.stopPropagation();
        });
    </script>
</body>
</html>`
}
