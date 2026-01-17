package ui

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tiagokriok/cdp/internal/config"
)

// captureOutput captures stdout during a function execution
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestSuccess(t *testing.T) {
	output := captureOutput(func() {
		Success("test message")
	})

	if !strings.Contains(output, "test message") {
		t.Errorf("Success() output = %q, want it to contain 'test message'", output)
	}
}

func TestError(t *testing.T) {
	output := captureOutput(func() {
		Error("error message")
	})

	if !strings.Contains(output, "error message") {
		t.Errorf("Error() output = %q, want it to contain 'error message'", output)
	}
}

func TestInfo(t *testing.T) {
	output := captureOutput(func() {
		Info("info message")
	})

	if !strings.Contains(output, "info message") {
		t.Errorf("Info() output = %q, want it to contain 'info message'", output)
	}
}

func TestWarn(t *testing.T) {
	output := captureOutput(func() {
		Warn("warning message")
	})

	if !strings.Contains(output, "warning message") {
		t.Errorf("Warn() output = %q, want it to contain 'warning message'", output)
	}
}

func TestHeader(t *testing.T) {
	output := captureOutput(func() {
		Header("header message")
	})

	if !strings.Contains(output, "header message") {
		t.Errorf("Header() output = %q, want it to contain 'header message'", output)
	}
}

func TestPrintProfileList_Empty(t *testing.T) {
	output := captureOutput(func() {
		PrintProfileList([]config.Profile{}, "")
	})

	if !strings.Contains(output, "No profiles found") {
		t.Errorf("PrintProfileList() with empty list = %q, want it to contain 'No profiles found'", output)
	}
}

func TestPrintProfileList_WithProfiles(t *testing.T) {
	profiles := []config.Profile{
		{
			Name: "work",
			Path: "/home/user/.claude-profiles/work",
			Metadata: config.ProfileMetadata{
				Description: "Work profile",
				CreatedAt:   time.Now().Add(-24 * time.Hour),
				LastUsed:    time.Now(),
				UsageCount:  5,
			},
		},
		{
			Name: "personal",
			Path: "/home/user/.claude-profiles/personal",
			Metadata: config.ProfileMetadata{
				Description: "Personal projects",
				CreatedAt:   time.Now().Add(-48 * time.Hour),
			},
		},
	}

	output := captureOutput(func() {
		PrintProfileList(profiles, "work")
	})

	if !strings.Contains(output, "work") {
		t.Errorf("PrintProfileList() output = %q, want it to contain 'work'", output)
	}
	if !strings.Contains(output, "personal") {
		t.Errorf("PrintProfileList() output = %q, want it to contain 'personal'", output)
	}
	if !strings.Contains(output, "Found 2 profile") {
		t.Errorf("PrintProfileList() output = %q, want it to contain 'Found 2 profile'", output)
	}
	if !strings.Contains(output, "Work profile") {
		t.Errorf("PrintProfileList() output = %q, want it to contain 'Work profile'", output)
	}
}

func TestPrintProfileInfo(t *testing.T) {
	profile := &config.Profile{
		Name: "test",
		Path: "/home/user/.claude-profiles/test",
		Metadata: config.ProfileMetadata{
			Description: "Test profile",
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			LastUsed:    time.Now(),
			UsageCount:  10,
			Template:    "default",
			CustomFlags: []string{"--verbose", "--debug"},
		},
	}

	output := captureOutput(func() {
		PrintProfileInfo(profile, true)
	})

	if !strings.Contains(output, "Profile: test") {
		t.Errorf("PrintProfileInfo() output = %q, want it to contain 'Profile: test'", output)
	}
	if !strings.Contains(output, "Current profile") {
		t.Errorf("PrintProfileInfo() output = %q, want it to contain 'Current profile'", output)
	}
	if !strings.Contains(output, "Test profile") {
		t.Errorf("PrintProfileInfo() output = %q, want it to contain 'Test profile'", output)
	}
	if !strings.Contains(output, "Usage count:  10") {
		t.Errorf("PrintProfileInfo() output = %q, want it to contain 'Usage count:  10'", output)
	}
	if !strings.Contains(output, "Template:") {
		t.Errorf("PrintProfileInfo() output = %q, want it to contain 'Template:'", output)
	}
	if !strings.Contains(output, "--verbose --debug") {
		t.Errorf("PrintProfileInfo() output = %q, want it to contain '--verbose --debug'", output)
	}
}

func TestPrintProfileInfo_NotCurrent(t *testing.T) {
	profile := &config.Profile{
		Name: "test",
		Path: "/home/user/.claude-profiles/test",
		Metadata: config.ProfileMetadata{
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
	}

	output := captureOutput(func() {
		PrintProfileInfo(profile, false)
	})

	if strings.Contains(output, "Current profile") {
		t.Errorf("PrintProfileInfo() output = %q, should not contain 'Current profile' when isCurrent is false", output)
	}
	if !strings.Contains(output, "Never") {
		t.Errorf("PrintProfileInfo() output = %q, want it to contain 'Never' for unused profile", output)
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		contains string
	}{
		{
			name:     "just now",
			time:     time.Now(),
			contains: "just now",
		},
		{
			name:     "1 minute ago",
			time:     time.Now().Add(-1 * time.Minute),
			contains: "1 minute ago",
		},
		{
			name:     "5 minutes ago",
			time:     time.Now().Add(-5 * time.Minute),
			contains: "5 minutes ago",
		},
		{
			name:     "1 hour ago",
			time:     time.Now().Add(-1 * time.Hour),
			contains: "1 hour ago",
		},
		{
			name:     "3 hours ago",
			time:     time.Now().Add(-3 * time.Hour),
			contains: "3 hours ago",
		},
		{
			name:     "1 day ago",
			time:     time.Now().Add(-24 * time.Hour),
			contains: "1 day ago",
		},
		{
			name:     "5 days ago",
			time:     time.Now().Add(-5 * 24 * time.Hour),
			contains: "5 days ago",
		},
		{
			name:     "1 week ago",
			time:     time.Now().Add(-7 * 24 * time.Hour),
			contains: "1 week ago",
		},
		{
			name:     "3 weeks ago",
			time:     time.Now().Add(-21 * 24 * time.Hour),
			contains: "3 weeks ago",
		},
		{
			name:     "1 month ago",
			time:     time.Now().Add(-35 * 24 * time.Hour),
			contains: "1 month ago",
		},
		{
			name:     "6 months ago",
			time:     time.Now().Add(-180 * 24 * time.Hour),
			contains: "6 months ago",
		},
		{
			name:     "1 year ago",
			time:     time.Now().Add(-400 * 24 * time.Hour),
			contains: "1 year ago",
		},
		{
			name:     "3 years ago",
			time:     time.Now().Add(-3 * 365 * 24 * time.Hour),
			contains: "3 years ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTime(tt.time)
			if result != tt.contains {
				t.Errorf("formatTime() = %q, want %q", result, tt.contains)
			}
		})
	}
}
