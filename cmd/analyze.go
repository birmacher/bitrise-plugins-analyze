package cmd

import (
	"bitrise-plugins-analyze/internal/visualize"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	AppExtension       = ".app"
	IpaExtension       = ".ipa"
	XcarchiveExtension = ".xcarchive"
)

var (
	generateHTML bool
	outputDir    string
	generateJSON bool
)

var annotateCmd = &cobra.Command{
	Use:   "analyze [path]",
	Short: "Analyze App",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var app_path string

		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			stdin, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}

			app_path = string(stdin)
		}

		if len(args) == 1 {
			if app_path != "" {
				return errors.New("if stdin piping is used then app_path argument can't be set")
			}

			app_path = args[0]
		}

		if app_path == "" {
			return errors.New("app_path is empty")
		}

		bundle, err := analyzeAppBundle(app_path)
		if err != nil {
			return err
		}

		// Handle output directory
		if outputDir == "" {
			outputDir, err = os.Getwd()
			if err != nil {
				return err
			}
		}

		// Create output directory if it doesn't exist
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return err
		}

		if generateJSON {
			// Create JSON file named after bundle ID
			jsonFileName := fmt.Sprintf("%s.json", bundle.BundleID)
			jsonPath := filepath.Join(outputDir, jsonFileName)

			// Marshal the bundle with indentation for better readability
			jsonData, err := json.MarshalIndent(bundle, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal bundle data: %v", err)
			}

			// Write JSON file
			if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
				return fmt.Errorf("failed to write JSON file: %v", err)
			}
		}

		if generateHTML {
			// Generate HTML file named after bundle ID
			htmlFileName := fmt.Sprintf("%s.html", bundle.BundleID)
			htmlPath := filepath.Join(outputDir, htmlFileName)
			err = visualize.GenerateHTML(bundle, htmlPath)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(annotateCmd)
	annotateCmd.Flags().BoolVar(&generateHTML, "html", false, "Generate HTML visualization")
	annotateCmd.Flags().BoolVar(&generateJSON, "json", false, "Generate JSON output file")
	annotateCmd.Flags().StringVar(&outputDir, "output-dir", "", "Directory where the output files will be generated (default: current directory)")
}
