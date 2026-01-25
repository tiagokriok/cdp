package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tiagokriok/cdp/internal/config"
	"github.com/tiagokriok/cdp/internal/ui"
	"github.com/tiagokriok/cdp/pkg/aliases"
)

type wizardStep int

const (
	selectProfileStep wizardStep = iota
	enterAliasStep
	reviewStep
)

type aliasWizardModel struct {
	step            wizardStep
	profiles        []config.Profile
	cursor          int
	aliasInput      textinput.Model
	selectedProfile string
	suggestion      string
	validationErr   string
	aliases         map[string]string
	existingAliases map[string]string // Track already-configured aliases
	shellType       string
	completed       bool
	focusInput      bool // Flag to focus input on next render
	isRenaming      bool // True if renaming existing alias, false if creating new
}

func initialWizardModel(profiles []config.Profile, shellType string, existingAliases map[string]string) aliasWizardModel {
	ti := textinput.New()
	ti.Placeholder = "Enter alias name"
	ti.CharLimit = 50
	ti.Width = 30

	return aliasWizardModel{
		step:            selectProfileStep,
		profiles:        profiles,
		cursor:          0,
		aliasInput:      ti,
		aliases:         make(map[string]string),
		existingAliases: existingAliases,
		shellType:       shellType,
		completed:       false,
		focusInput:      false,
		isRenaming:      false,
	}
}

func (m aliasWizardModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m aliasWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.step {
		case selectProfileStep:
			return m.updateProfileSelection(msg)
		case enterAliasStep:
			// Handle special keys for alias input
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.step = selectProfileStep
				m.aliasInput.Blur()
				return m, nil
			case "enter":
				alias := m.aliasInput.Value()

				// For new aliases, use suggestion if empty
				if !m.isRenaming && alias == "" {
					alias = m.suggestion
				}

				// Validate
				if err := validateAlias(alias, m.aliases); err != nil {
					m.validationErr = err.Error()
					return m, nil
				}

				// Save the alias
				m.aliases[m.selectedProfile] = alias

				// Return to profile selection
				m.step = selectProfileStep
				m.aliasInput.Blur()
				return m, nil
			case "ctrl+u":
				m.aliasInput.SetValue("")
				m.validationErr = ""
				return m, nil
			}

			// For other keys, pass to textinput
			m.aliasInput, cmd = m.aliasInput.Update(msg)

			// Real-time validation
			currentValue := m.aliasInput.Value()
			if currentValue != "" {
				if err := validateAlias(currentValue, m.aliases); err != nil {
					m.validationErr = err.Error()
				} else {
					m.validationErr = ""
				}
			} else {
				m.validationErr = ""
			}

			return m, cmd
		case reviewStep:
			return m.updateReview(msg)
		}
	}

	return m, cmd
}

func (m aliasWizardModel) updateProfileSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		maxIndex := len(m.profiles) // +1 for "Done" option
		if m.cursor < maxIndex {
			m.cursor++
		}
	case "enter", "r":
		// Last item is "Done"
		if m.cursor == len(m.profiles) {
			m.step = reviewStep
			return m, nil
		}

		// Profile selected
		profile := m.profiles[m.cursor]

		m.selectedProfile = profile.Name
		m.isRenaming = false
		m.validationErr = ""

		// Check if profile already has alias in this session
		if newAlias, exists := m.aliases[profile.Name]; exists {
			m.aliasInput.SetValue(newAlias)
			m.isRenaming = true
		} else if existingAlias, exists := m.existingAliases[profile.Name]; exists {
			// Check if already has alias in RC file
			m.aliasInput.SetValue(existingAlias)
			m.isRenaming = true
		} else {
			// Create new alias
			m.suggestion = generateSafeSuggestion(profile.Name, m.aliases)
			m.aliasInput.SetValue("")
			m.aliasInput.Placeholder = fmt.Sprintf("(suggestion: %s)", m.suggestion)
		}

		m.step = enterAliasStep
		m.focusInput = true
		m.aliasInput.Focus()
		return m, textinput.Blink
	}
	return m, nil
}

func (m aliasWizardModel) updateReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "n":
		return m, tea.Quit
	case "y", "enter":
		m.completed = true
		return m, tea.Quit
	}
	return m, nil
}

func (m aliasWizardModel) View() string {
	switch m.step {
	case selectProfileStep:
		return m.viewProfileSelection()
	case enterAliasStep:
		return m.viewAliasInput()
	case reviewStep:
		return m.viewReview()
	}
	return ""
}

func (m aliasWizardModel) viewProfileSelection() string {
	var b strings.Builder

	b.WriteString(ui.HeaderStyle.Render("Select a profile to manage aliases:") + "\n\n")

	// List profiles
	for i, profile := range m.profiles {
		cursor := "  "
		if m.cursor == i {
			cursor = ui.CurrentSymbol + " "
		}

		status := ""
		// Check if already aliased in this session
		if alias, exists := m.aliases[profile.Name]; exists {
			status = ui.SuccessStyle.Render(fmt.Sprintf(" → '%s' (new)", alias))
		} else if existingAlias, exists := m.existingAliases[profile.Name]; exists {
			// Check if already has an alias in RC file
			status = ui.InfoStyle.Render(fmt.Sprintf(" → '%s'", existingAlias))
		}

		description := ""
		if profile.Metadata.Description != "" {
			description = ui.DimStyle.Render(profile.Metadata.Description)
		}

		b.WriteString(fmt.Sprintf("%s%-15s %s%s\n",
			cursor,
			profile.Name,
			description,
			status))
	}

	// "Done" option
	cursor := "  "
	if m.cursor == len(m.profiles) {
		cursor = ui.CurrentSymbol + " "
	}
	b.WriteString(fmt.Sprintf("\n%sDone          %s\n",
		cursor,
		ui.InfoStyle.Render("(Finish and install)")))

	b.WriteString("\n" + ui.DimStyle.Render("[↑↓ or j/k: navigate, Enter: create/rename, r: rename, q: quit]"))

	return b.String()
}

func (m aliasWizardModel) viewAliasInput() string {
	var b strings.Builder

	// Show different title based on create or rename
	if m.isRenaming {
		b.WriteString(ui.HeaderStyle.Render(fmt.Sprintf("Rename alias for '%s':", m.selectedProfile)) + "\n\n")
	} else {
		b.WriteString(ui.HeaderStyle.Render(fmt.Sprintf("Create alias for '%s':", m.selectedProfile)) + "\n\n")
	}

	// Show suggestion only for new aliases
	if !m.isRenaming && m.suggestion != "" {
		b.WriteString(fmt.Sprintf("Suggestion: %s\n", ui.InfoStyle.Render(m.suggestion)))
		b.WriteString(ui.DimStyle.Render("(Type your own or press Enter to use suggestion)") + "\n\n")
	} else if m.isRenaming {
		b.WriteString(ui.DimStyle.Render("(Edit the current alias or press Ctrl+U to clear and start fresh)") + "\n\n")
	}

	b.WriteString(m.aliasInput.View() + "\n\n")

	if m.validationErr != "" {
		b.WriteString(ui.ErrorStyle.Render("✗ "+m.validationErr) + "\n")
	} else if m.aliasInput.Value() != "" {
		b.WriteString(ui.SuccessStyle.Render("✓ Valid alias") + "\n")
	}

	b.WriteString("\n" + ui.DimStyle.Render("[Enter: confirm, Esc: back, Ctrl+U: clear, Ctrl+C: quit]"))

	return b.String()
}

func (m aliasWizardModel) viewReview() string {
	var b strings.Builder

	b.WriteString(ui.HeaderStyle.Render("Review aliases to install:") + "\n\n")

	if len(m.aliases) == 0 {
		b.WriteString(ui.InfoStyle.Render("No aliases selected.\n"))
		b.WriteString("\nPress any key to exit.")
		return b.String()
	}

	for profile, alias := range m.aliases {
		b.WriteString(fmt.Sprintf("  %s → %s\n",
			ui.SuccessStyle.Render(alias),
			fmt.Sprintf("cdp %s", profile)))
	}

	b.WriteString(fmt.Sprintf("\nShell: %s\n", ui.InfoStyle.Render(m.shellType)))
	b.WriteString("\n" + ui.InfoStyle.Render("Install these aliases? [Y/n]: "))

	return b.String()
}

// RunAliasWizard launches the interactive alias setup wizard
func RunAliasWizard() error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("CDP not initialized. Run 'cdp init' first")
	}

	// Get profiles
	pm := config.NewProfileManager(cfg)
	profiles, err := pm.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	if len(profiles) == 0 {
		ui.Info("No profiles found.")
		fmt.Println("Create a profile first:")
		fmt.Println("  cdp create <name> [description]")
		return nil
	}

	// Detect shell
	am, err := aliases.New()
	if err != nil {
		return fmt.Errorf("failed to detect shell: %w", err)
	}

	shellType := am.GetShellName()

	// Load existing aliases from RC file
	existingAliases, err := am.ListAliases()
	if err != nil {
		return fmt.Errorf("failed to read existing aliases: %w", err)
	}

	// Run wizard with existing aliases
	initial := initialWizardModel(profiles, shellType, existingAliases)
	p := tea.NewProgram(initial)
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running alias wizard: %w", err)
	}

	// Get results
	wizardModel := finalModel.(aliasWizardModel)

	// Check if user completed the wizard
	if !wizardModel.completed {
		ui.Info("Alias installation cancelled.")
		return nil
	}

	// Install aliases
	if len(wizardModel.aliases) > 0 {
		if err := am.InstallAliases(wizardModel.aliases); err != nil {
			return fmt.Errorf("failed to install aliases: %w", err)
		}

		ui.Success("Shell aliases installed!")
		for profile, alias := range wizardModel.aliases {
			fmt.Printf("  %s → cdp %s\n", alias, profile)
		}
		fmt.Printf("\nTo activate aliases immediately, run:\n  source %s\n\n", am.GetRCFile())
		fmt.Println("They will be available in new terminal sessions automatically.")
	} else {
		ui.Info("No aliases were configured.")
	}

	return nil
}

// Validation helpers

var reservedCommands = map[string]bool{
	"cd":     true,
	"ls":     true,
	"cp":     true,
	"mv":     true,
	"rm":     true,
	"cat":    true,
	"grep":   true,
	"sed":    true,
	"awk":    true,
	"git":    true,
	"docker": true,
	"npm":    true,
	"sudo":   true,
	"vim":    true,
	"nano":   true,
	"echo":   true,
	"pwd":    true,
	"chmod":  true,
	"chown":  true,
	"find":   true,
	"make":   true,
	"go":     true,
	"python": true,
	"node":   true,
	"yarn":   true,
	"curl":   true,
	"wget":   true,
	"ssh":    true,
	"scp":    true,
	"kill":   true,
	"ps":     true,
	"top":    true,
	"man":    true,
	"less":   true,
	"more":   true,
	"tail":   true,
	"head":   true,
	"sort":   true,
	"uniq":   true,
}

var aliasPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func validateAlias(alias string, existingAliases map[string]string) error {
	if alias == "" {
		return fmt.Errorf("alias cannot be empty")
	}

	if len(alias) > 50 {
		return fmt.Errorf("alias too long (max 50 characters)")
	}

	if !aliasPattern.MatchString(alias) {
		return fmt.Errorf("use only letters, numbers, hyphens, and underscores")
	}

	if reservedCommands[alias] {
		return fmt.Errorf("'%s' conflicts with shell command", alias)
	}

	for _, existingAlias := range existingAliases {
		if existingAlias == alias {
			return fmt.Errorf("alias '%s' already used", alias)
		}
	}

	return nil
}

func generateSafeSuggestion(profileName string, existingAliases map[string]string) string {
	if len(profileName) == 0 {
		return profileName
	}

	candidates := []string{
		"c" + string(profileName[0]),
		"c" + profileName[:minInt(2, len(profileName))],
		"c" + profileName[:minInt(3, len(profileName))],
		"c" + profileName,
		profileName,
	}

	for _, candidate := range candidates {
		// Check if safe to use
		if !reservedCommands[candidate] {
			// Check if not already used
			alreadyUsed := false
			for _, existingAlias := range existingAliases {
				if existingAlias == candidate {
					alreadyUsed = true
					break
				}
			}
			if !alreadyUsed {
				return candidate
			}
		}
	}

	// Fallback: use profile name
	return profileName
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
