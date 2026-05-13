package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/joedorseyjr/syncd/internal/brew"
	"github.com/joedorseyjr/syncd/internal/config"
	"github.com/joedorseyjr/syncd/internal/plan"
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

			runner := &brew.ExecRunner{}
			state, err := brew.GetState(runner)
			if err != nil {
				return fmt.Errorf("querying Homebrew state: %w", err)
			}

			p := plan.Compute(cfg, &plan.State{
				Taps:  state.Taps,
				Brews: state.Brews,
				Casks: state.Casks,
			})

			if p.IsEmpty() {
				fmt.Println("Already in sync. No changes needed.")
				return nil
			}

			printPlan(p)

			if !yes {
				if !confirm(os.Stdin) {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			fmt.Println("\nApplying changes...")
			results := brew.Execute(runner, p)
			fmt.Print(brew.FormatResults(results))

			if brew.HasErrors(results) {
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
