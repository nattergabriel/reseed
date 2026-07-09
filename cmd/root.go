package cmd

import (
	"fmt"

	"github.com/nattergabriel/reseed/internal/reseed"
	"github.com/spf13/cobra"
)

var version = "dev"

// flagDir is the --dir persistent flag: overrides the project skills directory.
var flagDir string

var rootCmd = &cobra.Command{
	Use:   "reseed",
	Short: "Manage agent skills across projects",
	Long:  "reseed manages a personal skill library and lets you install skills into any project's skills directory.",
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true
	},
	RunE: runTUI,
}

func Execute() error {
	return rootCmd.Execute()
}

const (
	groupLibrary = "library"
	groupProject = "project"
)

func init() {
	rootCmd.PersistentFlags().StringVar(&flagDir, "dir", "",
		fmt.Sprintf("override the skills directory (default %s)", reseed.DefaultSkillsDir))

	rootCmd.AddGroup(
		&cobra.Group{ID: groupLibrary, Title: "Library:"},
		&cobra.Group{ID: groupProject, Title: "Project:"},
	)

	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("reseed", version)
	},
}
