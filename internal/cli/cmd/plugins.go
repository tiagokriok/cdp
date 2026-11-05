package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tiagokriok/hat/internal/cli"
	"github.com/tiagokriok/hat/internal/config"
	"github.com/tiagokriok/hat/internal/ui"
)

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func profileForArg(cfg *config.Config, pm *config.ProfileManager, args []string) (*config.Profile, error) {
	name := ""
	if len(args) > 0 {
		name = args[0]
	}
	name, err := cli.ResolveProfileName(cfg, name)
	if err != nil {
		return nil, err
	}
	return pm.GetProfile(name)
}

var pluginsCmd = &cobra.Command{
	Use:   "plugins [profile]",
	Short: "List plugins of a profile (default: current)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, pm, err := loadProfileManager()
		if err != nil {
			return err
		}
		profile, err := profileForArg(cfg, pm, args)
		if err != nil {
			return err
		}
		plugins, err := config.ListPlugins(profile.Path)
		if err != nil {
			return err
		}
		if len(plugins) == 0 {
			ui.Info(fmt.Sprintf("No plugins in profile '%s'", profile.Name))
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "PLUGIN\tENABLED\tINSTALLED\tVERSION")
		for _, p := range plugins {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.ID, yesNo(p.Enabled), yesNo(p.Installed), valueOrNone(p.Version))
		}
		return w.Flush()
	},
}

func pluginsToggle(use string, enable bool) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <profile> <plugin@marketplace>",
		Short: use + " a plugin in a profile",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, pm, err := loadProfileManager()
			if err != nil {
				return err
			}
			profile, err := pm.GetProfile(args[0])
			if err != nil {
				return err
			}
			if err := config.SetPluginEnabled(profile.Path, args[1], enable); err != nil {
				return err
			}
			ui.Success(fmt.Sprintf("%sd %s in '%s'", use, args[1], profile.Name))
			return nil
		},
	}
}

var pluginsCopyCmd = &cobra.Command{
	Use:   "copy <from> <to> <plugin@marketplace>",
	Short: "Copy a plugin's marketplace and enablement to another profile",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, pm, err := loadProfileManager()
		if err != nil {
			return err
		}
		from, err := pm.GetProfile(args[0])
		if err != nil {
			return err
		}
		to, err := pm.GetProfile(args[1])
		if err != nil {
			return err
		}
		if err := config.CopyPlugin(from.Path, to.Path, args[2]); err != nil {
			return err
		}
		ui.Success(fmt.Sprintf("Enabled %s in '%s' (marketplace copied from '%s')", args[2], to.Name, from.Name))
		ui.Info("Claude installs the plugin on the next launch of that profile.")
		return nil
	},
}

var pluginsDiffCmd = &cobra.Command{
	Use:   "diff <a> <b>",
	Short: "Compare enabled plugins between two profiles",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, pm, err := loadProfileManager()
		if err != nil {
			return err
		}
		a, err := pm.GetProfile(args[0])
		if err != nil {
			return err
		}
		b, err := pm.GetProfile(args[1])
		if err != nil {
			return err
		}
		onlyA, onlyB, err := config.DiffPlugins(a.Path, b.Path)
		if err != nil {
			return err
		}
		if len(onlyA) == 0 && len(onlyB) == 0 {
			ui.Success("Same enabled plugins")
			return nil
		}
		fmt.Printf("only in %s: %s\n", a.Name, valueOrNone(strings.Join(onlyA, ", ")))
		fmt.Printf("only in %s: %s\n", b.Name, valueOrNone(strings.Join(onlyB, ", ")))
		return nil
	},
}

func init() {
	pluginsCmd.AddCommand(pluginsToggle("enable", true), pluginsToggle("disable", false), pluginsCopyCmd, pluginsDiffCmd)
	rootCmd.AddCommand(pluginsCmd)
}
