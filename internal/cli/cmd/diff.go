package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tiagokriok/cdp/internal/config"
	"github.com/tiagokriok/cdp/internal/ui"
)

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff <profile1> <profile2>",
	Short: "Compare two profiles",
	Long: `Shows the differences between two profiles' settings.

Compares settings.json files and metadata between the two profiles.

Example:
  cdp diff work personal`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		profile1 := args[0]
		profile2 := args[1]

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("CDP not initialized. Run 'cdp init' first")
		}

		pm := config.NewProfileManager(cfg)

		p1, err := pm.GetProfile(profile1)
		if err != nil {
			return fmt.Errorf("profile '%s' does not exist", profile1)
		}

		p2, err := pm.GetProfile(profile2)
		if err != nil {
			return fmt.Errorf("profile '%s' does not exist", profile2)
		}

		ui.Header(fmt.Sprintf("Comparing: %s vs %s", profile1, profile2))
		fmt.Println()

		// Compare metadata
		compareMetadata(p1, p2)

		// Compare settings.json
		compareSettings(p1.Path, p2.Path, profile1, profile2)

		return nil
	},
}

func compareMetadata(p1, p2 *config.Profile) {
	fmt.Println(ui.InfoStyle.Render("Metadata:"))

	if p1.Metadata.Description != p2.Metadata.Description {
		fmt.Printf("  Description:\n")
		fmt.Printf("    %s: %s\n", ui.ProfileStyle.Render(p1.Name), p1.Metadata.Description)
		fmt.Printf("    %s: %s\n", ui.ProfileStyle.Render(p2.Name), p2.Metadata.Description)
	}

	if p1.Metadata.Template != p2.Metadata.Template {
		fmt.Printf("  Template:\n")
		fmt.Printf("    %s: %s\n", ui.ProfileStyle.Render(p1.Name), valueOrNone(p1.Metadata.Template))
		fmt.Printf("    %s: %s\n", ui.ProfileStyle.Render(p2.Name), valueOrNone(p2.Metadata.Template))
	}

	fmt.Printf("  Usage count:\n")
	fmt.Printf("    %s: %d\n", ui.ProfileStyle.Render(p1.Name), p1.Metadata.UsageCount)
	fmt.Printf("    %s: %d\n", ui.ProfileStyle.Render(p2.Name), p2.Metadata.UsageCount)

	fmt.Println()
}

func compareSettings(path1, path2, name1, name2 string) {
	settings1 := loadJSONFile(filepath.Join(path1, config.ClaudeSettingsFile))
	settings2 := loadJSONFile(filepath.Join(path2, config.ClaudeSettingsFile))

	fmt.Println(ui.InfoStyle.Render("Settings differences:"))

	if settings1 == nil && settings2 == nil {
		fmt.Println("  Both profiles have empty settings")
		return
	}

	if reflect.DeepEqual(settings1, settings2) {
		fmt.Println("  " + ui.SuccessStyle.Render("Settings are identical"))
		return
	}

	// Get all keys from both
	keys := getAllKeys(settings1, settings2)

	for _, key := range keys {
		v1, has1 := settings1[key]
		v2, has2 := settings2[key]

		if !has1 {
			fmt.Printf("  %s: only in %s\n", ui.WarnStyle.Render(key), ui.ProfileStyle.Render(name2))
			printValue("    ", v2)
		} else if !has2 {
			fmt.Printf("  %s: only in %s\n", ui.WarnStyle.Render(key), ui.ProfileStyle.Render(name1))
			printValue("    ", v1)
		} else if !reflect.DeepEqual(v1, v2) {
			fmt.Printf("  %s: differs\n", ui.WarnStyle.Render(key))
			fmt.Printf("    %s: ", ui.ProfileStyle.Render(name1))
			printValue("", v1)
			fmt.Printf("    %s: ", ui.ProfileStyle.Render(name2))
			printValue("", v2)
		}
	}
}

func loadJSONFile(path string) map[string]interface{} {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}

	return result
}

func getAllKeys(m1, m2 map[string]interface{}) []string {
	keys := make(map[string]bool)
	for k := range m1 {
		keys[k] = true
	}
	for k := range m2 {
		keys[k] = true
	}

	result := make([]string, 0, len(keys))
	for k := range keys {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

func printValue(indent string, v interface{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		data, _ := json.MarshalIndent(val, indent, "  ")
		fmt.Printf("%s\n", strings.TrimPrefix(string(data), indent))
	case []interface{}:
		data, _ := json.MarshalIndent(val, indent, "  ")
		fmt.Printf("%s\n", strings.TrimPrefix(string(data), indent))
	default:
		fmt.Printf("%v\n", val)
	}
}

func valueOrNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
