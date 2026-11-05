package cmd

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tiagokriok/hat/internal/config"
	"github.com/tiagokriok/hat/internal/ui"
)

var sharedCmd = &cobra.Command{
	Use:   "shared",
	Short: "Manage the shared layer (skills, agents, commands, ... ) linked into profiles",
	Long: `Content in ~/.hat/shared/ can be symlinked per entry into profiles that opt in.

Kinds: skills, agents, commands, output-styles, rules, CLAUDE.md`,
}

func loadProfileManager() (*config.Config, *config.ProfileManager, error) {
	if !config.Exists() {
		return nil, nil, fmt.Errorf("hat not initialized. Run 'hat init' first")
	}
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}
	return cfg, config.NewProfileManager(cfg), nil
}

func printSyncResult(profile string, res config.SyncResult) {
	for _, l := range res.Linked {
		ui.Success(fmt.Sprintf("[%s] linked %s", profile, l))
	}
	for _, p := range res.Pruned {
		ui.Info(fmt.Sprintf("[%s] pruned %s", profile, p))
	}
	for _, c := range res.Conflicts {
		ui.Warn(fmt.Sprintf("[%s] conflict, profile's own %s kept", profile, c))
	}
}

var sharedListCmd = &cobra.Command{
	Use:   "list",
	Short: "List shared content and which profiles opt in",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, pm, err := loadProfileManager()
		if err != nil {
			return err
		}
		shared, err := config.ListShared()
		if err != nil {
			return err
		}
		profiles, err := pm.ListProfiles()
		if err != nil {
			return err
		}
		sort.Slice(profiles, func(i, j int) bool { return profiles[i].Name < profiles[j].Name })

		for _, kind := range config.SharedKinds() {
			var users []string
			for _, p := range profiles {
				for _, k := range p.Metadata.Shared {
					if k == kind {
						users = append(users, p.Name)
					}
				}
			}
			if len(shared[kind]) == 0 && len(users) == 0 {
				continue
			}
			fmt.Println(ui.InfoStyle.Render(kind))
			fmt.Printf("  entries: %s\n", valueOrNone(strings.Join(shared[kind], ", ")))
			fmt.Printf("  profiles: %s\n", valueOrNone(strings.Join(users, ", ")))
		}
		if len(shared) == 0 {
			ui.Info("Shared layer is empty. Use 'hat shared add <kind> <path>'.")
		}
		return nil
	},
}

func sharedToggle(use string, enable bool) *cobra.Command {
	verb := "Opt a profile into"
	if !enable {
		verb = "Opt a profile out of"
	}
	return &cobra.Command{
		Use:   use + " <profile> <kind...>",
		Short: verb + " shared kinds",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, pm, err := loadProfileManager()
			if err != nil {
				return err
			}
			if err := pm.SetSharedKinds(args[0], args[1:], enable); err != nil {
				return err
			}
			res, err := pm.SyncShared(args[0])
			printSyncResult(args[0], res)
			return err
		},
	}
}

var sharedSyncCmd = &cobra.Command{
	Use:   "sync [profile]",
	Short: "Sync shared links (all profiles if none given)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, pm, err := loadProfileManager()
		if err != nil {
			return err
		}
		if len(args) == 1 {
			res, err := pm.SyncShared(args[0])
			printSyncResult(args[0], res)
			return err
		}
		all, err := pm.SyncAllShared()
		for _, n := range slices.Sorted(maps.Keys(all)) {
			printSyncResult(n, all[n])
		}
		return err
	},
}

var sharedAddCmd = &cobra.Command{
	Use:   "add <kind> <path>",
	Short: "Move a profile's skill/agent/... into shared and symlink it back",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, pm, err := loadProfileManager()
		if err != nil {
			return err
		}
		if err := config.ValidateSharedKind(args[0]); err != nil {
			return err
		}
		if !config.PromptYesNo(fmt.Sprintf("Move %s into the shared layer and replace it with a symlink?", args[1]), false) {
			ui.Info("Cancelled")
			return nil
		}
		profile, err := pm.AddToShared(args[0], args[1])
		if err != nil {
			return err
		}
		ui.Success(fmt.Sprintf("Moved to shared, profile '%s' now opts into %s", profile, args[0]))
		return nil
	},
}

func init() {
	sharedCmd.AddCommand(sharedListCmd, sharedToggle("link", true), sharedToggle("unlink", false), sharedSyncCmd, sharedAddCmd)
	rootCmd.AddCommand(sharedCmd)
}
