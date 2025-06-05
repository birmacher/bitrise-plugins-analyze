package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"bitrise-plugins-analyze/appbundle/core"
	"github.com/spf13/cobra"
)

type diffEntry struct {
	Path    string
	OldSize int64
	NewSize int64
}

type diffResult struct {
	Added   []diffEntry `json:"added"`
	Removed []diffEntry `json:"removed"`
	Changed []diffEntry `json:"changed"`
}

var jsonOutput string

var diffCmd = &cobra.Command{
	Use:   "diff <old.json> <new.json>",
	Short: "Diff two analysis JSON files",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldBundle, err := readBundle(args[0])
		if err != nil {
			return err
		}
		newBundle, err := readBundle(args[1])
		if err != nil {
			return err
		}

		added, removed, changed := diffBundles(oldBundle, newBundle)
		printDiff(added, removed, changed)
		if jsonOutput != "" {
			if err := writeDiffJSON(added, removed, changed, jsonOutput); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().StringVar(&jsonOutput, "json", "", "Path to write diff as JSON")
}

func readBundle(p string) (*core.AppBundle, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var b core.AppBundle
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

func flattenFiles(info core.FileInfo, m map[string]core.FileInfo) {
	if info.RelativePath != "." {
		m[info.RelativePath] = info
	}
	for _, child := range info.Children {
		flattenFiles(child, m)
	}
}

func diffBundles(oldB, newB *core.AppBundle) (added, removed, changed []diffEntry) {
	oldMap := map[string]core.FileInfo{}
	flattenFiles(oldB.Files, oldMap)
	newMap := map[string]core.FileInfo{}
	flattenFiles(newB.Files, newMap)

	for path, oldInfo := range oldMap {
		newInfo, ok := newMap[path]
		if !ok {
			removed = append(removed, diffEntry{Path: path, OldSize: oldInfo.Size})
			continue
		}
		if oldInfo.Size != newInfo.Size {
			changed = append(changed, diffEntry{Path: path, OldSize: oldInfo.Size, NewSize: newInfo.Size})
		}
	}

	for path, newInfo := range newMap {
		if _, ok := oldMap[path]; !ok {
			added = append(added, diffEntry{Path: path, NewSize: newInfo.Size})
		}
	}

	sort.Slice(added, func(i, j int) bool { return added[i].Path < added[j].Path })
	sort.Slice(removed, func(i, j int) bool { return removed[i].Path < removed[j].Path })
	sort.Slice(changed, func(i, j int) bool { return changed[i].Path < changed[j].Path })

	return
}

func writeDiffJSON(added, removed, changed []diffEntry, path string) error {
	result := diffResult{Added: added, Removed: removed, Changed: changed}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func printDiff(added, removed, changed []diffEntry) {
	if len(added) > 0 {
		fmt.Println("Added files:")
		for _, e := range added {
			fmt.Printf("  %s (%d bytes)\n", e.Path, e.NewSize)
		}
	}

	if len(removed) > 0 {
		fmt.Println("Removed files:")
		for _, e := range removed {
			fmt.Printf("  %s (%d bytes)\n", e.Path, e.OldSize)
		}
	}

	if len(changed) > 0 {
		fmt.Println("Changed files:")
		for _, e := range changed {
			fmt.Printf("  %s (%d -> %d bytes)\n", e.Path, e.OldSize, e.NewSize)
		}
	}
}
