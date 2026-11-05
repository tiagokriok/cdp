package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tiagokriok/hat/internal/config"
	"github.com/tiagokriok/hat/internal/ui"
)

func normalizeRulePath(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") || filepath.IsAbs(path) {
		return path, nil
	}
	return filepath.Abs(path)
}

func HandleRuleAdd(path, profile string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	if !config.NewProfileManager(cfg).ProfileExists(profile) {
		return fmt.Errorf("profile '%s' does not exist", profile)
	}
	stored, err := normalizeRulePath(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}
	cfg.AddRule(stored, profile)
	if err := cfg.Save(); err != nil {
		return err
	}
	ui.Success(fmt.Sprintf("Rule added: %s -> %s", stored, profile))
	return nil
}

func HandleRuleList() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	if len(cfg.Rules) == 0 {
		ui.Info("No rules. Add one with 'hat rule add <path> <profile>'.")
		return nil
	}
	cwd, _ := os.Getwd()
	matched := cfg.MatchRule(cwd)
	for i := range cfg.Rules {
		marker := " "
		if matched == &cfg.Rules[i] {
			marker = ui.CurrentSymbol
		}
		fmt.Printf("%s %s -> %s\n", marker, cfg.Rules[i].Path, cfg.Rules[i].Profile)
	}
	return nil
}

func HandleRuleRemove(path string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	stored, err := normalizeRulePath(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}
	if !cfg.RemoveRule(stored) {
		return fmt.Errorf("no rule for path '%s'", path)
	}
	if err := cfg.Save(); err != nil {
		return err
	}
	ui.Success(fmt.Sprintf("Rule removed: %s", path))
	return nil
}
