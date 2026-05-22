package cli

import (
	"fmt"
	"os"

	"github.com/joedorseyjr/syncd/internal/brew"
	"github.com/joedorseyjr/syncd/internal/config"
	"github.com/joedorseyjr/syncd/internal/defaults"
	"github.com/joedorseyjr/syncd/internal/plan"
	"github.com/joedorseyjr/syncd/internal/runner"
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

			r := &runner.ExecRunner{}
			state, err := brew.GetState(r)
			if err != nil {
				return fmt.Errorf("querying Homebrew state: %w", err)
			}

			leaves, err := brew.GetLeaves(r)
			if err != nil {
				return fmt.Errorf("querying Homebrew leaves: %w", err)
			}

			p := plan.Compute(cfg, &plan.State{
				Taps:   state.Taps,
				Brews:  state.Brews,
				Leaves: leaves,
				Casks:  state.Casks,
			})

			var drifted []defaults.DriftEntry
			if len(cfg.Defaults) > 0 {
				drifted, err = defaults.ComputeDrift(r, cfg.Defaults)
				if err != nil {
					return fmt.Errorf("computing defaults drift: %w", err)
				}
			}

			if p.IsEmpty() && len(drifted) == 0 {
				fmt.Println("Already in sync. No changes needed.")
				return nil
			}

			printPlan(p)
			printDefaultsDrift(drifted)

			if p.HasChanges() || len(drifted) > 0 {
				os.Exit(2)
			}
			return nil
		},
	}
}

func printDefaultsDrift(drifted []defaults.DriftEntry) {
	if len(drifted) == 0 {
		return
	}
	fmt.Printf("\nDefaults drift:\n")
	for _, d := range drifted {
		if d.Current == "unset" {
			fmt.Printf("  %s+%s %s %s: unset → %s\n", Green, Reset, d.Domain, d.Key, d.Desired)
		} else {
			fmt.Printf("  %s~%s %s %s: %s → %s\n", Yellow, Reset, d.Domain, d.Key, d.Current, d.Desired)
		}
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
