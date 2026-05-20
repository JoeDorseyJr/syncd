package cli

import (
	"fmt"
	"os"

	"github.com/joedorseyjr/syncd/internal/brew"
	"github.com/joedorseyjr/syncd/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewInitCmd() *cobra.Command {
	var output string
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a config file from current system state",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := brew.CheckAvailable(); err != nil {
				return err
			}

			runner := &brew.ExecRunner{}
			state, err := brew.GetState(runner)
			if err != nil {
				return fmt.Errorf("querying Homebrew state: %w", err)
			}
			leaves, err := brew.GetLeaves(runner)
			if err != nil {
				return fmt.Errorf("querying Homebrew leaves: %w", err)
			}

			cfg := &config.Config{
				Taps:  state.Taps,
				Brews: leaves,
				Casks: state.Casks,
				Pin:   []string{},
				Cleanup: config.Cleanup{
					RemoveUnlisted: false,
					ClearCache:     false,
					Autoremove:     true,
				},
			}

			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("marshaling config: %w", err)
			}

			if output == "" {
				fmt.Print(string(data))
				return nil
			}

			if !force {
				if _, err := os.Stat(output); err == nil {
					return fmt.Errorf("file already exists: %s (use --force to overwrite)", output)
				}
			}

			if err := os.WriteFile(output, data, 0644); err != nil {
				return fmt.Errorf("writing config file: %w", err)
			}
			fmt.Printf("Config written to %s\n", output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "write config to file instead of stdout")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing file")
	return cmd
}
