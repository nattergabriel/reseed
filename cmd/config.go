package cmd

import (
	"fmt"

	"github.com/nattergabriel/reseed/internal/config"
	"github.com/nattergabriel/reseed/internal/project"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config <key> [value]",
	Short: "Get or set configuration",
	Long: `Get or set global configuration values.

Available keys:
  dir    Default skills directory (e.g. .claude/skills)`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		if key != "dir" {
			return fmt.Errorf("unknown key: %s", key)
		}

		cfg, err := config.LoadGlobal()
		if err != nil {
			return err
		}

		if len(args) == 1 {
			if cfg.Dir == "" {
				fmt.Printf("%s (default)\n", project.DefaultSkillsDir)
			} else {
				fmt.Println(cfg.Dir)
			}
			return nil
		}

		cfg.Dir = args[1]
		if err := config.SaveGlobal(cfg); err != nil {
			return err
		}

		fmt.Printf("%s set to %s\n", key, args[1])
		return nil
	},
}
