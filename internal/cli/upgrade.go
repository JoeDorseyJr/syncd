package cli

import (
	"fmt"
	"os"

	"github.com/joedorseyjr/syncd/internal/brew"
	"github.com/joedorseyjr/syncd/internal/config"
	"github.com/spf13/cobra"
)

func NewUpgradeCmd(cfgFile *string) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade all outdated packages (respects pin list)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := brew.CheckAvailable(); err != nil {
				return err
			}

			// Load config for pin list (optional)
			var pinned map[string]struct{}
			if *cfgFile != "" {
				cfg, err := config.Load(*cfgFile)
				if err != nil {
					return err
				}
				pinned = toSet(cfg.Pin)
			} else {
				// Try default config, ignore if missing
				cfg, err := config.Load("")
				if err == nil {
					pinned = toSet(cfg.Pin)
				}
			}

			runner := &brew.ExecRunner{}

			outdatedBrews, err := brew.GetOutdated(runner)
			if err != nil {
				return fmt.Errorf("querying outdated formulae: %w", err)
			}
			outdatedCasks, err := brew.GetOutdatedCasks(runner)
			if err != nil {
				return fmt.Errorf("querying outdated casks: %w", err)
			}

			// Filter out pinned
			brews := filterPinned(outdatedBrews, pinned)
			casks := filterPinned(outdatedCasks, pinned)

			if len(brews) == 0 && len(casks) == 0 {
				fmt.Println("All packages are up to date.")
				return nil
			}

			// Print upgrade plan
			printUpgradePlan(brews, casks)

			if !yes {
				if !confirm(os.Stdin) {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			fmt.Println("\nUpgrading...")
			results := brew.Upgrade(runner, brews, casks)
			fmt.Print(brew.FormatResults(results, Green, Red, Reset))

			if brew.HasErrors(results) {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompt")
	return cmd
}

func printUpgradePlan(brews, casks []string) {
	if len(brews) > 0 {
		fmt.Println("\nFormulae to upgrade:")
		for _, name := range brews {
			fmt.Printf("  %s~%s %s\n", Yellow, Reset, name)
		}
	}
	if len(casks) > 0 {
		fmt.Println("\nCasks to upgrade:")
		for _, name := range casks {
			fmt.Printf("  %s~%s %s\n", Yellow, Reset, name)
		}
	}
}

func toSet(items []string) map[string]struct{} {
	s := make(map[string]struct{}, len(items))
	for _, item := range items {
		s[item] = struct{}{}
	}
	return s
}

func filterPinned(items []string, pinned map[string]struct{}) []string {
	if len(pinned) == 0 {
		return items
	}
	var result []string
	for _, item := range items {
		if _, ok := pinned[item]; !ok {
			result = append(result, item)
		}
	}
	return result
}
