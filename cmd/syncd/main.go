package main

import (
	"fmt"
	"os"

	"github.com/joedorseyjr/syncd/internal/cli"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

var cfgFile string

var rootCmd = &cobra.Command{
	Use:     "syncd",
	Short:   "Declarative package manager for macOS using Homebrew",
	Version: version,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ~/.config/syncd/config.yaml)")
	rootCmd.AddCommand(cli.NewPlanCmd(&cfgFile))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
