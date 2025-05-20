package cmd

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const (
	AppExtension       = ".app"
	IpaExtension       = ".ipa"
	XcarchiveExtension = ".xcarchive"
)

var (
	context string
	style   string
)

func isSupportedAppPath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == AppExtension || ext == IpaExtension || ext == XcarchiveExtension
}

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

		err := analyzeAppBundle(app_path)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(annotateCmd)

	annotateCmd.Flags().StringVarP(&context, "context", "c", "",
		"the context to find existing annotations and replace their content")
	annotateCmd.Flags().StringVarP(&style, "style", "s", "default",
		"the style to use for this annotation, such as default, error, warning, info")
}
