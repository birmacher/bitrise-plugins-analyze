package visualize

import (
	"bitrise-plugins-analyze/internal/analyzer"
)

type Chart struct {
	Labels  []string `json:"labels"`
	Parents []string `json:"parents"`
	Values  []int64  `json:"values"`
	Types   []string `json:"types"`
	Ids     []string `json:"ids"`
}

func GeneratePlotlyChart(bundle *analyzer.AppBundle) Chart {
	chart := Chart{
		Labels:  []string{},
		Parents: []string{},
		Values:  []int64{},
		Types:   []string{},
		Ids:     []string{},
	}

	var traverseFiles func(files analyzer.FileInfo, parent string)
	traverseFiles = func(files analyzer.FileInfo, parent string) {
		if parent == "" {
			chart.Labels = append(chart.Labels, bundle.AppName)
			chart.Values = append(chart.Values, bundle.InstallSize)
		} else {
			chart.Labels = append(chart.Labels, files.RelativePath)
			chart.Values = append(chart.Values, files.Size)
		}
		chart.Parents = append(chart.Parents, parent)
		chart.Types = append(chart.Types, files.Type)
		chart.Ids = append(chart.Ids, files.RelativePath)

		for _, child := range files.Children {
			traverseFiles(child, files.RelativePath)
		}
	}
	traverseFiles(bundle.Files, "")

	return chart
}
