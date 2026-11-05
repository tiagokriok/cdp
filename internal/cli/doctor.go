package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/tiagokriok/hat/internal/config"
	"github.com/tiagokriok/hat/internal/executor"
	"github.com/tiagokriok/hat/internal/ui"
)

var doctorTimeout = 15 * time.Second

var ErrDoctorFailed = errors.New("doctor found problems")

type authStatus struct {
	LoggedIn   bool   `json:"loggedIn"`
	AuthMethod string `json:"authMethod"`
	Email      string `json:"email"`
	OrgName    string `json:"orgName"`
}

type doctorReport struct {
	failures int
}

func (r *doctorReport) ok(msg string)   { ui.Success(msg) }
func (r *doctorReport) warn(msg string) { ui.Warn(msg) }
func (r *doctorReport) fail(msg string) {
	r.failures++
	ui.Error(msg)
}

func runClaude(claudePath string, env []string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), doctorTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, claudePath, args...)
	cmd.Env = env
	out, err := cmd.Output()
	if ctx.Err() != nil {
		return string(out), fmt.Errorf("timed out after %s", doctorTimeout)
	}
	return string(out), err
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	line, _, _ := strings.Cut(s, "\n")
	return line
}

func checkAuth(r *doctorReport, claudePath string, profile *config.Profile) {
	env := executor.BuildEnv(os.Environ(), profile.Path, config.IsDefaultClaudeDir(profile.Path))
	out, err := runClaude(claudePath, env, "auth", "status")

	var st authStatus
	if jsonErr := json.Unmarshal([]byte(out), &st); jsonErr != nil {
		detail := firstLine(out)
		if err != nil && (detail == "" || strings.HasPrefix(err.Error(), "timed out")) {
			detail = err.Error()
		}
		r.warn(fmt.Sprintf("auth status unavailable: %s", detail))
		return
	}

	if !st.LoggedIn {
		r.warn("not logged in")
		return
	}
	who := st.Email
	if st.OrgName != "" {
		who += " (" + st.OrgName + ")"
	}
	r.ok(fmt.Sprintf("logged in: %s via %s", who, st.AuthMethod))
}

func brokenLinks(dir string) []string {
	var broken []string
	scan := func(d string) {
		entries, err := os.ReadDir(d)
		if err != nil {
			return
		}
		for _, e := range entries {
			p := filepath.Join(d, e.Name())
			if e.Type()&os.ModeSymlink == 0 {
				continue
			}
			if _, err := os.Stat(p); err != nil {
				broken = append(broken, p)
			}
		}
	}
	scan(dir)
	for _, sub := range []string{"skills", "agents", "commands"} {
		scan(filepath.Join(dir, sub))
	}
	return broken
}

func checkProfile(r *doctorReport, pm *config.ProfileManager, claudePath string, name, path string) {
	fmt.Printf("\nProfile %s\n", name)

	if _, err := os.Stat(path); err != nil {
		r.fail(fmt.Sprintf("profile directory unreachable: %v", err))
		return
	}
	profile, err := pm.GetProfile(name)
	if err != nil {
		r.fail(fmt.Sprintf("cannot load profile: %v", err))
		return
	}
	if err := pm.ValidateProfile(profile); err != nil {
		r.fail(fmt.Sprintf("invalid profile: %v", err))
	} else {
		r.ok("structure valid")
	}

	if config.IsDefaultClaudeDir(path) {
		if home, err := os.UserHomeDir(); err == nil {
			stale := filepath.Join(home, ".claude", config.ClaudeConfigFile)
			if _, err := os.Stat(stale); err == nil {
				if _, err := os.Stat(filepath.Join(home, config.ClaudeConfigFile)); err == nil {
					r.warn(fmt.Sprintf("%s is ignored (stale), Claude uses ~/%s for the default profile", stale, config.ClaudeConfigFile))
				}
			}
		}
	}

	if broken := brokenLinks(path); len(broken) > 0 {
		for _, b := range broken {
			r.fail(fmt.Sprintf("broken symlink: %s", b))
		}
	} else {
		r.ok("no broken symlinks")
	}

	if claudePath != "" {
		checkAuth(r, claudePath, profile)
	}
}

func HandleDoctor() error {
	r := &doctorReport{}

	ui.Header("hat doctor")
	fmt.Println()

	claudePath, err := executor.NewExecutor().FindClaude()
	if err != nil {
		r.fail("claude binary not found")
		claudePath = ""
	} else {
		out, verr := runClaude(claudePath, os.Environ(), "--version")
		if verr != nil {
			r.warn(fmt.Sprintf("claude found at %s but --version failed: %v", claudePath, verr))
		} else {
			r.ok(fmt.Sprintf("claude %s (%s)", firstLine(out), claudePath))
		}
	}

	if overrides := executor.AuthEnvOverrides(os.Environ()); len(overrides) > 0 {
		r.warn(fmt.Sprintf("%s set in environment, overriding profile logins", strings.Join(overrides, ", ")))
	} else {
		r.ok("no auth env overrides")
	}

	if !config.Exists() {
		r.fail("hat is not initialized, run 'hat init'")
		return ErrDoctorFailed
	}
	cfg, err := config.Load()
	if err != nil {
		r.fail(fmt.Sprintf("config unreadable: %v", err))
		return ErrDoctorFailed
	}
	configDir, _ := config.GetConfigDir()
	r.ok(fmt.Sprintf("config dir %s", configDir))

	if info, err := os.Stat(cfg.ProfilesDir); err != nil || !info.IsDir() {
		r.fail(fmt.Sprintf("profiles dir missing: %s", cfg.ProfilesDir))
	} else {
		r.ok(fmt.Sprintf("profiles dir %s", cfg.ProfilesDir))
		pm := config.NewProfileManager(cfg)
		entries, _ := os.ReadDir(cfg.ProfilesDir)
		for _, e := range entries {
			path := filepath.Join(cfg.ProfilesDir, e.Name())
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				continue
			}
			checkProfile(r, pm, claudePath, e.Name(), path)
		}
	}

	if len(cfg.Rules) > 0 {
		fmt.Println("\nRules")
		pm := config.NewProfileManager(cfg)
		for _, rule := range cfg.Rules {
			if !pm.ProfileExists(rule.Profile) {
				r.fail(fmt.Sprintf("%s -> %s: profile does not exist", rule.Path, rule.Profile))
				continue
			}
			if _, err := os.Stat(config.ResolveRulePath(rule.Path)); err != nil {
				r.warn(fmt.Sprintf("%s -> %s: path does not exist", rule.Path, rule.Profile))
				continue
			}
			r.ok(fmt.Sprintf("%s -> %s", rule.Path, rule.Profile))
		}
	}

	fmt.Println()
	if r.failures > 0 {
		ui.Error(fmt.Sprintf("%d problem(s) found", r.failures))
		return ErrDoctorFailed
	}
	ui.Success("all checks passed")
	return nil
}
