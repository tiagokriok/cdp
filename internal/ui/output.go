package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/tiagokriok/cdp/internal/config"
)

var (
	// Color styles
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	InfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	WarnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	HeaderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	DimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	// Symbols
	SuccessSymbol = SuccessStyle.Render("✓")
	ErrorSymbol   = ErrorStyle.Render("✗")
	InfoSymbol    = InfoStyle.Render("ℹ")
	WarnSymbol    = WarnStyle.Render("⚠")
	CurrentSymbol = SuccessStyle.Render("▶")

	// List styles
	ProfileStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true)
	DescriptionStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	TimestampStyle      = DimStyle
	CurrentProfileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
)

// Success prints a success message
func Success(msg string) {
	fmt.Printf("%s %s\n", SuccessSymbol, msg)
}

// Error prints an error message
func Error(msg string) {
	fmt.Printf("%s %s\n", ErrorSymbol, lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(msg))
}

// Info prints an info message
func Info(msg string) {
	fmt.Printf("%s %s\n", InfoSymbol, msg)
}

// Warn prints a warning message
func Warn(msg string) {
	fmt.Printf("%s %s\n", WarnSymbol, lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(msg))
}

// Header prints a styled header
func Header(msg string) {
	fmt.Println(HeaderStyle.Render(msg))
}

// PrintProfileList prints a styled list of profiles
func PrintProfileList(profiles []config.Profile, currentProfile string) {
	if len(profiles) == 0 {
		Info("No profiles found.")
		fmt.Println()
		fmt.Println("Create a profile:")
		fmt.Println("  cdp create <name> [description]")
		return
	}

	Header(fmt.Sprintf("Found %d profile(s):", len(profiles)))
	fmt.Println()

	for _, profile := range profiles {
		// Mark current profile
		marker := "  "
		nameStyle := ProfileStyle
		if profile.Name == currentProfile {
			marker = CurrentSymbol + " "
			nameStyle = CurrentProfileStyle
		}

		fmt.Printf("%s%s\n", marker, nameStyle.Render(profile.Name))

		if profile.Metadata.Description != "" {
			fmt.Printf("   %s %s\n", DimStyle.Render("│"), DescriptionStyle.Render(profile.Metadata.Description))
		}

		fmt.Printf("   %s Created: %s\n", DimStyle.Render("│"), TimestampStyle.Render(formatTime(profile.Metadata.CreatedAt)))

		if !profile.Metadata.LastUsed.IsZero() {
			fmt.Printf("   %s Last used: %s\n", DimStyle.Render("│"), TimestampStyle.Render(formatTime(profile.Metadata.LastUsed)))
		}

		fmt.Println()
	}

	if currentProfile != "" {
		fmt.Printf("%s Current profile: %s\n", CurrentSymbol, CurrentProfileStyle.Render(currentProfile))
	}
}

// PrintProfileInfo prints detailed information about a profile
func PrintProfileInfo(profile *config.Profile, isCurrent bool) {
	Header(fmt.Sprintf("Profile: %s", profile.Name))
	fmt.Println()

	if isCurrent {
		fmt.Printf("%s %s\n", CurrentSymbol, SuccessStyle.Render("Current profile"))
		fmt.Println()
	}

	// Description
	if profile.Metadata.Description != "" {
		fmt.Printf("Description:  %s\n", DescriptionStyle.Render(profile.Metadata.Description))
	}

	// Location
	fmt.Printf("Location:     %s\n", DimStyle.Render(profile.Path))

	// Created
	fmt.Printf("Created:      %s (%s)\n",
		profile.Metadata.CreatedAt.Format("2006-01-02 15:04:05"),
		TimestampStyle.Render(formatTime(profile.Metadata.CreatedAt)))

	// Last used
	if !profile.Metadata.LastUsed.IsZero() {
		fmt.Printf("Last used:    %s (%s)\n",
			profile.Metadata.LastUsed.Format("2006-01-02 15:04:05"),
			TimestampStyle.Render(formatTime(profile.Metadata.LastUsed)))
	} else {
		fmt.Printf("Last used:    %s\n", DimStyle.Render("Never"))
	}

	// Usage Count
	fmt.Printf("Usage count:  %d\n", profile.Metadata.UsageCount)

	// Template
	if profile.Metadata.Template != "" {
		fmt.Printf("Template:     %s\n", DescriptionStyle.Render(profile.Metadata.Template))
	}

	// Custom Flags
	if len(profile.Metadata.CustomFlags) > 0 {
		fmt.Printf("Custom Flags: %s\n", DescriptionStyle.Render(strings.Join(profile.Metadata.CustomFlags, " ")))
	}
}

// formatTime formats a time.Time into a human-readable relative time string
func formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case diff < 365*24*time.Hour:
		months := int(diff.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(diff.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}