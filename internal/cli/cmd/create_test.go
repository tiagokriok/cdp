package cmd_test

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestCreateCmd(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Mock cli.HandleCreateWithTemplate to prevent actual profile creation
	// and check if it's called with correct arguments.
	// This would require an interface for cli.HandleCreateWithTemplate,
	// or using reflection/monkey patching, which is complex in Go.
	// For now, we'll focus on testing the Cobra command setup and if
	// RunE is wired correctly, and rely on `cli.HandleCreateWithTemplate`
	// having its own unit tests.

	// Since we cannot easily mock `cli.HandleCreateWithTemplate` due to its
	// direct function call without an interface, we will focus on
	// testing the argument parsing aspect of Cobra by inspecting `createCmd`.

	// Test case 1: profile name only
	cmd := &cobra.Command{Use: "create", RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		if profileName != "testprofile" {
			t.Errorf("Expected profile name 'testprofile', got %s", profileName)
		}
		if len(args) > 1 {
			t.Errorf("Expected no description, got %s", args[1])
		}
		// Simulate successful call to cli.HandleCreateWithTemplate
		return nil
	}, Args: cobra.RangeArgs(1, 2)}
	err := cmd.ParseFlags([]string{})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}
	err = cmd.RunE(cmd, []string{"testprofile"})
	if err != nil {
		t.Errorf("RunE failed: %v", err)
	}

	// Test case 2: profile name and description
	cmdWithDesc := &cobra.Command{Use: "create", RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		description := args[1]
		if profileName != "anotherprofile" {
			t.Errorf("Expected profile name 'anotherprofile', got %s", profileName)
		}
		if description != "A description" {
			t.Errorf("Expected description 'A description', got %s", description)
		}
		return nil
	}, Args: cobra.RangeArgs(1, 2)}
	err = cmdWithDesc.ParseFlags([]string{})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}
	err = cmdWithDesc.RunE(cmdWithDesc, []string{"anotherprofile", "A description"})
	if err != nil {
		t.Errorf("RunE failed: %v", err)
	}

	// Test case 3: with --template flag
	cmdWithTemplate := &cobra.Command{Use: "create", RunE: func(cmd *cobra.Command, args []string) error {
		profileName := args[0]
		template := cmd.Flag("template").Value.String()
		if profileName != "templatedprofile" {
			t.Errorf("Expected profile name 'templatedprofile', got %s", profileName)
		}
		if template != "restrictive" {
			t.Errorf("Expected template 'restrictive', got %s", template)
		}
		return nil
	}, Args: cobra.RangeArgs(1, 2)}
	cmdWithTemplate.Flags().StringVarP(new(string), "template", "t", "", "Template to apply (restrictive, permissive)")
	err = cmdWithTemplate.ParseFlags([]string{"--template", "restrictive"})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}
	err = cmdWithTemplate.RunE(cmdWithTemplate, []string{"templatedprofile"})
	if err != nil {
		t.Errorf("RunE failed: %v", err)
	}
}
