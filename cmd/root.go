package cmd

import (
	"github.com/nattergabriel/reseed/internal/project"
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "reseed",
	Short: "Manage agent skills across projects",
	Long:  "Reseed manages a personal skill library and lets you install skills into any project's .agents/skills/ directory.",
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true
	},
}

func Execute() error {
	return rootCmd.Execute()
}

const (
	groupLibrary = "library"
	groupProject = "project"
)

func init() {
	rootCmd.PersistentFlags().StringVar(&project.SkillsDirOverride, "dir", "", "override the skills directory (default .agents/skills)")

	rootCmd.AddGroup(
		&cobra.Group{ID: groupLibrary, Title: "Library:"},
		&cobra.Group{ID: groupProject, Title: "Project:"},
	)

	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("reseed", version)
	},
}
