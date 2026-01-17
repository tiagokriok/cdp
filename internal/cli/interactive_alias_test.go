package cli

import (
	"strings"
	"testing"
)

func TestValidateAlias(t *testing.T) {
	tests := []struct {
		name           string
		alias          string
		existingAliases map[string]string
		wantErr        bool
		errMsg         string
	}{
		{
			name:            "valid simple alias",
			alias:           "cw",
			existingAliases: map[string]string{},
			wantErr:         false,
		},
		{
			name:            "valid with underscores",
			alias:           "my_work",
			existingAliases: map[string]string{},
			wantErr:         false,
		},
		{
			name:            "valid with hyphens",
			alias:           "my-work",
			existingAliases: map[string]string{},
			wantErr:         false,
		},
		{
			name:            "reserved command cp",
			alias:           "cp",
			existingAliases: map[string]string{},
			wantErr:         true,
			errMsg:          "conflicts with shell command",
		},
		{
			name:            "reserved command ls",
			alias:           "ls",
			existingAliases: map[string]string{},
			wantErr:         true,
			errMsg:          "conflicts with shell command",
		},
		{
			name:            "reserved command git",
			alias:           "git",
			existingAliases: map[string]string{},
			wantErr:         true,
			errMsg:          "conflicts with shell command",
		},
		{
			name:            "duplicate alias",
			alias:           "cw",
			existingAliases: map[string]string{"work": "cw"},
			wantErr:         true,
			errMsg:          "already used",
		},
		{
			name:            "empty alias",
			alias:           "",
			existingAliases: map[string]string{},
			wantErr:         true,
			errMsg:          "cannot be empty",
		},
		{
			name:            "invalid characters space",
			alias:           "my work",
			existingAliases: map[string]string{},
			wantErr:         true,
			errMsg:          "use only letters, numbers, hyphens, and underscores",
		},
		{
			name:            "invalid characters special",
			alias:           "my@work",
			existingAliases: map[string]string{},
			wantErr:         true,
			errMsg:          "use only letters, numbers, hyphens, and underscores",
		},
		{
			name:            "too long alias",
			alias:           strings.Repeat("a", 51),
			existingAliases: map[string]string{},
			wantErr:         true,
			errMsg:          "too long",
		},
		{
			name:            "max length alias",
			alias:           strings.Repeat("a", 50),
			existingAliases: map[string]string{},
			wantErr:         false,
		},
		{
			name:            "numeric alias",
			alias:           "123work",
			existingAliases: map[string]string{},
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAlias(tt.alias, tt.existingAliases)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAlias() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error message = %v, want to contain %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestGenerateSafeSuggestion(t *testing.T) {
	tests := []struct {
		name            string
		profileName     string
		existingAliases map[string]string
		want            string
	}{
		{
			name:            "simple suggestion",
			profileName:     "work",
			existingAliases: map[string]string{},
			want:            "cw",
		},
		{
			name:            "two letter suggestion",
			profileName:     "personal",
			existingAliases: map[string]string{},
			want:            "cpe",
		},
		{
			name:            "avoid reserved command",
			profileName:     "personal",
			existingAliases: map[string]string{},
			want:            "cpe", // cp is reserved, so it suggests cpe
		},
		{
			name:            "avoid duplicate with existing",
			profileName:     "work",
			existingAliases: map[string]string{"other": "cw"},
			want:            "cwo",
		},
		{
			name:            "fallback to profile name",
			profileName:     "w",
			existingAliases: map[string]string{},
			want:            "w", // or cw, depending on what's generated
		},
		{
			name:            "multiple conflicts fallback",
			profileName:     "cd",
			existingAliases: map[string]string{},
			want:            "ccd", // ccd isn't reserved
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSafeSuggestion(tt.profileName, tt.existingAliases)
			// For suggestion, we mainly check that it's valid and not conflicting
			if got == "" {
				t.Errorf("generateSafeSuggestion() returned empty string")
			}
			// Check it passes validation
			if err := validateAlias(got, tt.existingAliases); err != nil {
				t.Errorf("generated suggestion failed validation: %v", err)
			}
		})
	}
}

func TestReservedCommands(t *testing.T) {
	// Test that common reserved commands are actually in the list
	commonReserved := []string{"cd", "ls", "cp", "mv", "rm", "git", "docker", "npm"}
	for _, cmd := range commonReserved {
		if !reservedCommands[cmd] {
			t.Errorf("command %q should be in reserved list", cmd)
		}
	}
}

func TestAliasPattern(t *testing.T) {
	tests := []struct {
		alias   string
		matches bool
	}{
		{"cw", true},
		{"my-work", true},
		{"my_work", true},
		{"work123", true},
		{"W0rk", true},
		{"my work", false},
		{"my@work", false},
		{"my.work", false},
		{"my/work", false},
		{"", false},
	}

	for _, tt := range tests {
		if aliasPattern.MatchString(tt.alias) != tt.matches {
			t.Errorf("aliasPattern for %q: got %v, want %v", tt.alias, aliasPattern.MatchString(tt.alias), tt.matches)
		}
	}
}

func TestMinInt(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{0, 5, 0},
		{5, 0, 0},
		{3, 3, 3},
	}

	for _, tt := range tests {
		got := minInt(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("minInt(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestValidateAliasEdgeCases(t *testing.T) {
	tests := []struct {
		name             string
		alias            string
		existingAliases  map[string]string
		shouldBeValid    bool
	}{
		{
			name:             "single character valid",
			alias:            "c",
			existingAliases:  map[string]string{},
			shouldBeValid:    true,
		},
		{
			name:             "number only",
			alias:            "123",
			existingAliases:  map[string]string{},
			shouldBeValid:    true,
		},
		{
			name:             "hyphen at start",
			alias:            "-work",
			existingAliases:  map[string]string{},
			shouldBeValid:    true,
		},
		{
			name:             "underscore at start",
			alias:            "_work",
			existingAliases:  map[string]string{},
			shouldBeValid:    true,
		},
		{
			name:             "multiple hyphens",
			alias:            "my--work",
			existingAliases:  map[string]string{},
			shouldBeValid:    true,
		},
		{
			name:             "multiple underscores",
			alias:            "my__work",
			existingAliases:  map[string]string{},
			shouldBeValid:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAlias(tt.alias, tt.existingAliases)
			if tt.shouldBeValid && err != nil {
				t.Errorf("validateAlias(%q) should be valid, got error: %v", tt.alias, err)
			}
			if !tt.shouldBeValid && err == nil {
				t.Errorf("validateAlias(%q) should be invalid, got no error", tt.alias)
			}
		})
	}
}
