package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/joedorseyjr/syncd/internal/brew"
	"github.com/joedorseyjr/syncd/internal/config"
	"github.com/joedorseyjr/syncd/internal/defaults"
	"github.com/joedorseyjr/syncd/internal/runner"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewInitCmd() *cobra.Command {
	var output string
	var force bool
	var defaultsFlag string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a config file from current system state",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if defaultsFlag != "" {
				cfg.Defaults = snapshotDefaults(r, defaultsFlag)
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
	cmd.Flags().StringVar(&defaultsFlag, "defaults", "", "comma-separated domain:key pairs to snapshot")
	return cmd
}

func snapshotDefaults(r runner.CommandRunner, flag string) []config.DefaultEntry {
	var entries []config.DefaultEntry
	pairs := strings.Split(flag, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		idx := strings.Index(pair, ":")
		if idx < 0 {
			fmt.Fprintf(os.Stderr, "Warning: invalid pair %q (expected domain:key)\n", pair)
			continue
		}
		domain := pair[:idx]
		key := pair[idx+1:]

		typeStr, ok, err := defaults.ReadType(r, domain, key)
		if err != nil || !ok {
			fmt.Fprintf(os.Stderr, "Warning: cannot read %s:%s, skipping\n", domain, key)
			continue
		}

		valStr, ok, err := defaults.ReadValue(r, domain, key)
		if err != nil || !ok {
			fmt.Fprintf(os.Stderr, "Warning: cannot read %s:%s, skipping\n", domain, key)
			continue
		}

		cfgType := mapType(typeStr)
		value := parseValue(cfgType, valStr)

		entry := config.DefaultEntry{
			Domain: domain,
			Key:    key,
			Type:   cfgType,
			Value:  value,
		}

		if app, known := defaults.KnownApps[domain]; known && app != "" {
			entry.Kill = []string{app}
		}

		entries = append(entries, entry)
	}
	return entries
}

func mapType(raw string) string {
	switch {
	case strings.Contains(raw, "integer"):
		return "int"
	case strings.Contains(raw, "float"):
		return "float"
	case strings.Contains(raw, "boolean"):
		return "bool"
	default:
		return "string"
	}
}

func parseValue(cfgType, raw string) interface{} {
	switch cfgType {
	case "int":
		var v int
		fmt.Sscanf(raw, "%d", &v)
		return v
	case "float":
		var v float64
		fmt.Sscanf(raw, "%f", &v)
		return v
	case "bool":
		return raw == "1" || strings.EqualFold(raw, "true")
	default:
		return raw
	}
}
