package cli

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    *ParsedArgs
		wantErr bool
	}{
		{
			name: "no arguments",
			args: []string{},
			want: &ParsedArgs{Command: "help"},
		},
		{
			name: "init command",
			args: []string{"init"},
			want: &ParsedArgs{Command: "init"},
		},
		{
			name: "list command",
			args: []string{"list"},
			want: &ParsedArgs{Command: "list"},
		},
		{
			name: "list alias (ls)",
			args: []string{"ls"},
			want: &ParsedArgs{Command: "list"},
		},
		{
			name: "current command",
			args: []string{"current"},
			want: &ParsedArgs{Command: "current"},
		},
		{
			name: "help command",
			args: []string{"help"},
			want: &ParsedArgs{Command: "help"},
		},
		{
			name: "help flag (--help)",
			args: []string{"--help"},
			want: &ParsedArgs{Command: "help"},
		},
		{
			name: "help flag (-h)",
			args: []string{"-h"},
			want: &ParsedArgs{Command: "help"},
		},
		{
			name: "version command",
			args: []string{"version"},
			want: &ParsedArgs{Command: "version"},
		},
		{
			name: "version flag (--version)",
			args: []string{"--version"},
			want: &ParsedArgs{Command: "version"},
		},
		{
			name: "version flag (-v)",
			args: []string{"-v"},
			want: &ParsedArgs{Command: "version"},
		},
		{
			name: "create with name only",
			args: []string{"create", "work"},
			want: &ParsedArgs{
				Command:     "create",
				ProfileName: "work",
			},
		},
		{
			name: "create with description",
			args: []string{"create", "work", "Work profile for SmartTalks"},
			want: &ParsedArgs{
				Command:     "create",
				ProfileName: "work",
				Description: "Work profile for SmartTalks",
			},
		},
		{
			name: "create with multi-word description",
			args: []string{"create", "personal", "My", "personal", "profile"},
			want: &ParsedArgs{
				Command:     "create",
				ProfileName: "personal",
				Description: "My personal profile",
			},
		},
		{
			name:    "create without name",
			args:    []string{"create"},
			wantErr: true,
		},
		{
			name: "delete command",
			args: []string{"delete", "old-profile"},
			want: &ParsedArgs{
				Command:     "delete",
				ProfileName: "old-profile",
			},
		},
		{
			name: "delete alias (del)",
			args: []string{"del", "old-profile"},
			want: &ParsedArgs{
				Command:     "delete",
				ProfileName: "old-profile",
			},
		},
		{
			name: "delete alias (rm)",
			args: []string{"rm", "old-profile"},
			want: &ParsedArgs{
				Command:     "delete",
				ProfileName: "old-profile",
			},
		},
		{
			name:    "delete without name",
			args:    []string{"delete"},
			wantErr: true,
		},
		{
			name: "switch to profile (basic)",
			args: []string{"work"},
			want: &ParsedArgs{
				Command:     "switch",
				ProfileName: "work",
				ClaudeFlags: []string{},
			},
		},
		{
			name: "switch with --no-run",
			args: []string{"work", "--no-run"},
			want: &ParsedArgs{
				Command:     "switch",
				ProfileName: "work",
				NoRun:       true,
				ClaudeFlags: []string{},
			},
		},
		{
			name: "switch with Claude flags",
			args: []string{"work", "--continue", "--verbose"},
			want: &ParsedArgs{
				Command:     "switch",
				ProfileName: "work",
				ClaudeFlags: []string{"--continue", "--verbose"},
			},
		},
		{
			name: "switch with Claude flags and --no-run",
			args: []string{"work", "--continue", "--no-run", "--verbose"},
			want: &ParsedArgs{
				Command:     "switch",
				ProfileName: "work",
				NoRun:       true,
				ClaudeFlags: []string{"--continue", "--verbose"},
			},
		},
		{
			name: "switch with complex flags",
			args: []string{"personal", "--model", "opus", "--temperature", "0.7"},
			want: &ParsedArgs{
				Command:     "switch",
				ProfileName: "personal",
				ClaudeFlags: []string{"--model", "opus", "--temperature", "0.7"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestIsValidCommand(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want bool
	}{
		{"init is valid", "init", true},
		{"create is valid", "create", true},
		{"list is valid", "list", true},
		{"delete is valid", "delete", true},
		{"current is valid", "current", true},
		{"switch is valid", "switch", true},
		{"help is valid", "help", true},
		{"version is valid", "version", true},
		{"invalid command", "invalid", false},
		{"random is invalid", "random", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidCommand(tt.cmd); got != tt.want {
				t.Errorf("IsValidCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestGetCommandHelp(t *testing.T) {
	tests := []struct {
		cmd     string
		wantNil bool
	}{
		{"init", false},
		{"create", false},
		{"list", false},
		{"delete", false},
		{"current", false},
		{"switch", false},
		{"help", false},
		{"version", false},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			help := GetCommandHelp(tt.cmd)
			if (help == "") == !tt.wantNil {
				t.Errorf("GetCommandHelp(%q) returned empty string", tt.cmd)
			}
		})
	}
}

func TestGetUsage(t *testing.T) {
	usage := GetUsage()
	if usage == "" {
		t.Error("GetUsage() returned empty string")
	}

	// Check for key sections
	requiredStrings := []string{
		"CDP",
		"USAGE",
		"COMMANDS",
		"EXAMPLES",
		"init",
		"create",
		"list",
		"delete",
	}

	for _, s := range requiredStrings {
		if !contains(usage, s) {
			t.Errorf("GetUsage() missing required string: %q", s)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
