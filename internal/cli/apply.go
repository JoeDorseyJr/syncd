package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/joedorseyjr/syncd/internal/brew"
	"github.com/joedorseyjr/syncd/internal/config"
	"github.com/joedorseyjr/syncd/internal/defaults"
	"github.com/joedorseyjr/syncd/internal/plan"
	"github.com/joedorseyjr/syncd/internal/runner"
	"github.com/spf13/cobra"
)

func NewApplyCmd(cfgFile *string) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Reconcile system state to match config",
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

			if !yes {
				if !confirm(os.Stdin) {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			hasFailure := false

			if !p.IsEmpty() {
				fmt.Println("\nApplying changes...")
				results := brew.Execute(r, p)
				fmt.Print(brew.FormatResults(results, Green, Red, Reset))
				if brew.HasErrors(results) {
					hasFailure = true
				}
			}

			if len(drifted) > 0 {
				fmt.Println("\nWriting defaults...")
				writeResults, killResults := defaults.WriteDrifted(r, drifted)
				for _, wr := range writeResults {
					if wr.Err != nil {
						fmt.Printf("  %s✗%s %s %s: %v\n", Red, Reset, wr.Domain, wr.Key, wr.Err)
						hasFailure = true
					} else {
						fmt.Printf("  %s✓%s %s %s\n", Green, Reset, wr.Domain, wr.Key)
					}
				}
				for _, kr := range killResults {
					if kr.Err != nil {
						fmt.Printf("  %s~%s killall %s (not running)\n", Yellow, Reset, kr.App)
					} else {
						fmt.Printf("  %s✓%s killall %s\n", Green, Reset, kr.App)
					}
				}
			}

			if hasFailure {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompt")
	return cmd
}

func confirm(r io.Reader) bool {
	fmt.Print("\nApply these changes? [y/N] ")
	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return answer == "y" || answer == "yes"
	}
	return false
}
