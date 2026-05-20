package cli

import (
	"fmt"
	"os"

	"github.com/joedorseyjr/syncd/internal/brew"
	"github.com/joedorseyjr/syncd/internal/config"
	"github.com/joedorseyjr/syncd/internal/plan"
	"github.com/spf13/cobra"
)

func NewPlanCmd(cfgFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "plan",
		Short: "Show what changes would be made",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(*cfgFile)
			if err != nil {
				return err
			}

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

			p := plan.Compute(cfg, &plan.State{
				Taps:   state.Taps,
				Brews:  state.Brews,
				Leaves: leaves,
				Casks:  state.Casks,
			})
			if p.IsEmpty() {
				fmt.Println("Already in sync. No changes needed.")
				return nil
			}

			printPlan(p)

			// Exit 2 only when there are package changes (drift).
			// Cleanup-only plans (autoremove/clear_cache) are maintenance, not drift.
			if p.HasChanges() {
				os.Exit(2)
			}
			return nil
		},
	}
}

func printPlan(p *plan.Plan) {
	printSection("Taps to add", p.TapsToAdd, Green+"+"+Reset)
	printSection("Taps to remove", p.TapsToRemove, Red+"-"+Reset)
	printSection("Brews to install", p.BrewsToInstall, Green+"+"+Reset)
	printSection("Brews to remove", p.BrewsToRemove, Red+"-"+Reset)
	printSection("Casks to install", p.CasksToInstall, Green+"+"+Reset)
	printSection("Casks to remove", p.CasksToRemove, Red+"-"+Reset)
	if p.Autoremove {
		fmt.Printf("  %s~%s brew autoremove\n", Yellow, Reset)
	}
	if p.ClearCache {
		fmt.Printf("  %s~%s brew cleanup\n", Yellow, Reset)
	}
}

func printSection(header string, items []string, prefix string) {
	if len(items) == 0 {
		return
	}
	fmt.Printf("\n%s:\n", header)
	for _, item := range items {
		fmt.Printf("  %s %s\n", prefix, item)
	}
}
