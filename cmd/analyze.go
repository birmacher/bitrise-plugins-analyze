package cmd

import (
	"bitrise-plugins-analyze/internal/visualize"
	"errors"
	"io"
	"os"

	"github.com/spf13/cobra"
)

const (
	AppExtension       = ".app"
	IpaExtension       = ".ipa"
	XcarchiveExtension = ".xcarchive"
)

var generateHTML bool

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

		if generateHTML {
			err = visualize.GenerateHTML(bundle, "index.html")
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
}
