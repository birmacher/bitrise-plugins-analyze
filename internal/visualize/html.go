package visualize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"bitrise-plugins-analyze/internal/analyzer"
)

// templateData represents the data structure for the HTML template
type templateData struct {
	Title        string
	AppName      string
	BundleID     string
	Platform     string
	Version      string
	DownloadSize string
	InstallSize  string
	FileTree     template.JS
}

// formatSize converts bytes to a human-readable string
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GenerateHTML generates an index.html file using the template and provided bundle data
func GenerateHTML(bundle *analyzer.AppBundle, outputPath string) error {
	// Parse the template from the constant string
	tmpl, err := template.New("template").Parse(templateHTML)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// Extract app name from the bundle path
	appName := filepath.Base(bundle.Files.RelativePath)
	if filepath.Ext(appName) == ".app" {
		appName = appName[:len(appName)-4] // Remove .app extension
	}

	// Convert FileTree to JSON string to make it safe for JavaScript
	fileTreeJSON, err := json.Marshal(bundle.Files)
	if err != nil {
		return fmt.Errorf("failed to marshal file tree: %v", err)
	}

	// Create template data
	data := templateData{
		Title:        "App Bundle Analysis",
		AppName:      appName,
		BundleID:     bundle.BundleID,
		Platform:     bundle.SupportedPlatforms[0], // Use first platform
		Version:      bundle.Version,
		DownloadSize: formatSize(bundle.DownloadSize),
		InstallSize:  formatSize(bundle.InstallSize),
		FileTree:     template.JS(fileTreeJSON),
	}

	// Create a buffer to store the rendered template
	var buf bytes.Buffer

	// Execute the template with the data
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Write the rendered template to the output file
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	return nil
}

// templateHTML contains the HTML template for visualization
const templateHTML = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>{{.Title}}</title>
  <script src="https://cdn.plot.ly/plotly-latest.min.js"></script>
  <style>
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      margin: 0;
      padding: 20px;
      background: #f5f5f7;
    }
    .container {
      max-width: 1200px;
      margin: 0 auto;
      background: white;
      border-radius: 10px;
      padding: 20px;
      box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    }
    #chart {
      width: 100%;
      max-width: 1160px;
      height: 700px;
      margin: auto;
    }
    h1 {
      color: #1d1d1f;
      margin-top: 0;
      margin-bottom: 20px;
      text-align: center;
    }
    .info-banner {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 20px;
      background: #f8f8fa;
      border-radius: 8px;
      padding: 20px;
      margin-bottom: 20px;
    }
    .info-item {
      display: flex;
      flex-direction: column;
    }
    .info-label {
      font-size: 12px;
      color: #666;
      margin-bottom: 4px;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }
    .info-value {
      font-size: 16px;
      color: #1d1d1f;
      font-weight: 500;
    }
    .size-info {
      color: #0066cc;
    }
    .tabs {
      margin: 0 0 20px 0;
      border-bottom: 1px solid #e1e1e1;
      display: flex;
      gap: 32px;
      padding: 0 16px;
    }
    .tab {
      padding: 12px 0;
      color: #666;
      cursor: pointer;
      position: relative;
      font-size: 14px;
      font-weight: 500;
      transition: color 0.2s;
      border: none;
      background: none;
      margin: 0;
    }
    .tab:hover {
      color: #000;
    }
    .tab.active {
      color: #0066cc;
    }
    .tab.active::after {
      content: '';
      position: absolute;
      bottom: -1px;
      left: 0;
      right: 0;
      height: 2px;
      background: #0066cc;
      border-radius: 2px;
    }
    .tab-content {
      display: none;
    }
    .tab-content.active {
      display: block;
    }
    .insights-grid {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 20px;
      padding: 20px 0;
    }
    .insight-card {
      background: #f8f8fa;
      border-radius: 8px;
      padding: 16px;
    }
    .insight-title {
      font-size: 14px;
      font-weight: 600;
      margin-bottom: 8px;
    }
    .insight-value {
      font-size: 24px;
      font-weight: 500;
      color: #0066cc;
    }
    .insight-description {
      font-size: 13px;
      color: #666;
      margin-top: 8px;
    }
    .breakdown-list {
      list-style: none;
      padding: 0;
      margin: 0;
    }
    .breakdown-item {
      display: flex;
      align-items: center;
      padding: 12px 16px;
      border-bottom: 1px solid #e1e1e1;
    }
    .breakdown-item:last-child {
      border-bottom: none;
    }
    .breakdown-type {
      width: 120px;
      font-weight: 500;
    }
    .breakdown-bar {
      flex-grow: 1;
      height: 8px;
      background: #e1e1e1;
      border-radius: 4px;
      margin: 0 16px;
      overflow: hidden;
    }
    .breakdown-bar-fill {
      height: 100%;
      background: #0066cc;
      border-radius: 4px;
    }
    .breakdown-size {
      width: 100px;
      text-align: right;
      color: #666;
    }
  </style>
</head>
<body>
<div class="container">
  <h1>{{.Title}}</h1>
  <div class="info-banner">
    <div class="info-item">
      <span class="info-label">Name</span>
      <span class="info-value" id="appName">{{.AppName}}</span>
    </div>
    <div class="info-item">
      <span class="info-label">Bundle ID</span>
      <span class="info-value" id="bundleId">{{.BundleID}}</span>
    </div>
    <div class="info-item">
      <span class="info-label">Platform</span>
      <span class="info-value" id="platform">{{.Platform}}</span>
    </div>
    <div class="info-item">
      <span class="info-label">Version</span>
      <span class="info-value" id="version">{{.Version}}</span>
    </div>
    <div class="info-item">
      <span class="info-label">Download Size</span>
      <span class="info-value size-info" id="downloadSize">{{.DownloadSize}}</span>
    </div>
    <div class="info-item">
      <span class="info-label">Install Size</span>
      <span class="info-value size-info" id="installSize">{{.InstallSize}}</span>
    </div>
  </div>
  
  <div class="tabs">
    <button class="tab active" data-tab="overview">Overview</button>
    <button class="tab" data-tab="breakdown">Breakdown</button>
    <button class="tab" data-tab="insights">Insights</button>
  </div>

  <div id="overview" class="tab-content active">
    <div id="chart"></div>
  </div>

  <div id="breakdown" class="tab-content">
    <ul class="breakdown-list" id="typeBreakdown">
      <!-- Will be populated by JavaScript -->
    </ul>
  </div>

  <div id="insights" class="tab-content">
    <div class="insights-grid">
      <div class="insight-card">
        <div class="insight-title">Largest File</div>
        <div class="insight-value" id="largestFile">-</div>
        <div class="insight-description">The single largest file in the bundle</div>
      </div>
      <div class="insight-card">
        <div class="insight-title">Asset Catalog Size</div>
        <div class="insight-value" id="assetCatalogSize">-</div>
        <div class="insight-description">Total size of .car files</div>
      </div>
      <div class="insight-card">
        <div class="insight-title">Binary Size</div>
        <div class="insight-value" id="binarySize">-</div>
        <div class="insight-description">Total size of binary files</div>
      </div>
      <div class="insight-card">
        <div class="insight-title">Resource Files</div>
        <div class="insight-value" id="resourceCount">-</div>
        <div class="insight-description">Number of resource files</div>
      </div>
    </div>
  </div>
</div>
<script>
const fileTree = {{.FileTree}};

// Helper function to format file size
function formatSize(bytes) {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// Calculate type breakdown
function calculateTypeBreakdown(node) {
  const breakdown = {};
  
  function processNode(node) {
    const type = node.type || 'unknown';
    breakdown[type] = (breakdown[type] || 0) + node.size;
    
    if (node.children) {
      node.children.forEach(processNode);
    }
  }
  
  processNode(node);
  return breakdown;
}

// Update insights
function updateInsights() {
  let largestFile = { size: 0 };
  let assetCatalogSize = 0;
  let binarySize = 0;
  let resourceCount = 0;
  
  function processNode(node) {
    if (node.size > largestFile.size) {
      largestFile = node;
    }
    
    if (node.type === 'asset_catalog') {
      assetCatalogSize += node.size;
    } else if (node.type === 'binary') {
      binarySize += node.size;
    }
    
    if (!node.children) {
      resourceCount++;
    }
    
    if (node.children) {
      node.children.forEach(processNode);
    }
  }
  
  processNode(fileTree);
  
  document.getElementById('largestFile').textContent = formatSize(largestFile.size);
  document.getElementById('assetCatalogSize').textContent = formatSize(assetCatalogSize);
  document.getElementById('binarySize').textContent = formatSize(binarySize);
  document.getElementById('resourceCount').textContent = resourceCount;
}

// Update type breakdown
function updateBreakdown() {
  const breakdown = calculateTypeBreakdown(fileTree);
  const totalSize = fileTree.size;
  const breakdownList = document.getElementById('typeBreakdown');
  breakdownList.innerHTML = '';
  
  Object.entries(breakdown)
    .sort((a, b) => b[1] - a[1])
    .forEach(([type, size]) => {
      const percentage = (size / totalSize * 100).toFixed(1);
      const li = document.createElement('li');
      li.className = 'breakdown-item';
      li.innerHTML = ` + "`" + `
        <span class="breakdown-type">${type}</span>
        <div class="breakdown-bar">
          <div class="breakdown-bar-fill" style="width: ${percentage}%"></div>
        </div>
        <span class="breakdown-size">${formatSize(size)}</span>
      ` + "`" + `;
      breakdownList.appendChild(li);
    });
}

// Tab switching functionality
document.querySelectorAll('.tab').forEach(tab => {
  tab.addEventListener('click', () => {
    // Remove active class from all tabs and content
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
    
    // Add active class to clicked tab and corresponding content
    tab.classList.add('active');
    const contentId = tab.getAttribute('data-tab');
    document.getElementById(contentId).classList.add('active');
  });
});

// Helper function to flatten the tree
function flatten(node, parentLabel, labels, parents, values, types) {
  const name = node.relative_path.split('/').pop();
  const label = (parentLabel === null) ? "MyApp" : name;
  labels.push(label);
  parents.push(parentLabel === null ? "" : parentLabel);
  values.push(node.size);
  types.push(node.type);
  if (node.children) {
    for (const child of node.children) {
      flatten(child, label, labels, parents, values, types);
    }
  }
}

const labels = [], parents = [], values = [], types = [];
flatten(fileTree, null, labels, parents, values, types);

// Optional: assign color per type
const colorMap = {
  directory: "#b0b4ff",
  binary: "#a5d8ff",
  asset_catalog: "#ffe066",
  "": "#ddd"
};
const markerColors = types.map(type => colorMap[type] || "#ddd");

const data = [{
  type: "treemap",
  labels,
  parents,
  values,
  branchvalues: 'total',
  textinfo: 'label+value',
  outsidetextfont: { size: 14, color: "#888" },
  leaf: { opacity: 0.8 },
  marker: { colors: markerColors },
  maxdepth: 3,
  pathbar: {
    visible: true,
    textfont: {
      size: 12,
      color: "#000"
    },
    side: "top",
    thickness: 25
  }
}];

// Initialize everything
function initChart() {
  const chartDiv = document.getElementById('chart');
  const layout = {
    margin: { l: 0, r: 0, b: 0, t: 32 },
    width: chartDiv.clientWidth,
    height: 700,
    uniformtext: {
      minsize: 10
    }
  };

  const config = {
    displayModeBar: false,
    responsive: true
  };

  Plotly.newPlot('chart', data, layout, config);
}

// Call initChart initially and on window resize
initChart();
window.addEventListener('resize', initChart);
updateBreakdown();
updateInsights();

// Add click handler for the chart
document.getElementById('chart').on('plotly_click', function(data) {
  const clickedLabel = data.points[0].label;
  console.log('Clicked label:', clickedLabel);
});
</script>
</body>
</html>`
