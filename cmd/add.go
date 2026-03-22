package cmd

import (
	"fmt"

	"github.com/nattergabriel/reseed/internal/library"
	"github.com/nattergabriel/reseed/internal/project"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add [skills or packs...]",
	Short:   "Add skills or packs to the current project",
	GroupID: groupProject,
	Long:  "Copies skills or packs from your library into the project's .agents/skills/ directory. Use --all to add every skill in your library.",
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		if !all && len(args) == 0 {
			return fmt.Errorf("provide one or more skill/pack names, or use --all")
		}

		lib, err := library.Open()
		if err != nil {
			return err
		}

		var skills []string
		if all {
			skills, err = lib.ListSkills()
			if err != nil {
				return err
			}
			if len(skills) == 0 {
				fmt.Println("No skills in library.")
				return nil
			}
		} else {
			for _, arg := range args {
				resolved, err := lib.ResolveSkillOrPack(arg)
				if err != nil {
					return err
				}
				skills = append(skills, resolved...)
			}
		}

		for _, name := range skills {
			if err := project.AddSkill(lib, name); err != nil {
				return fmt.Errorf("adding %s: %w", name, err)
			}
			fmt.Printf("  + %s\n", name)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().Bool("all", false, "Add all skills from the library")
}
