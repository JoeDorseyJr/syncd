package main

import (
	"fmt"
	"os"

	"github.com/joedorseyjr/syncd/internal/cli"
	"github.com/joedorseyjr/syncd/internal/runner"
	"github.com/spf13/cobra"
)

var version = "dev"

var cfgFile string

var rootCmd = &cobra.Command{
	Use:     "syncd",
	Short:   "Declarative package manager for macOS using Homebrew",
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cli.InitColor()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ~/.config/syncd/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&cli.NoColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&runner.Verbose, "verbose", false, "show full brew command output")
	rootCmd.AddCommand(cli.NewPlanCmd(&cfgFile))
	rootCmd.AddCommand(cli.NewApplyCmd(&cfgFile))
	rootCmd.AddCommand(cli.NewUpgradeCmd(&cfgFile))
	rootCmd.AddCommand(cli.NewInitCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
