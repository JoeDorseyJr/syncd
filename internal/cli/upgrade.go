package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joedorseyjr/syncd/internal/brew"
	"github.com/joedorseyjr/syncd/internal/config"
	"github.com/joedorseyjr/syncd/internal/download"
	"github.com/joedorseyjr/syncd/internal/progress"
	"github.com/joedorseyjr/syncd/internal/runner"
	"github.com/spf13/cobra"
)

// securityPkgs are packages where updates are security-critical.
var securityPkgs = map[string]bool{
	"ca-certificates": true, "openssl": true, "openssl@3": true,
	"gnutls": true, "libssh2": true, "gnupg": true, "gpg": true,
}

func NewUpgradeCmd(cfgFile *string) *cobra.Command {
	var yes bool
	var concurrency int

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade all outdated packages (respects pin list)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := brew.CheckAvailable(); err != nil {
				return err
			}

			r := &runner.ExecRunner{}

			// Update tap metadata before checking for outdated packages
			fmt.Printf("Updating Homebrew")
			done := make(chan struct{})
			go func() {
				for {
					select {
					case <-done:
						return
					case <-time.After(500 * time.Millisecond):
						fmt.Print(".")
					}
				}
			}()
			r.RunMutate("brew", "update")
			close(done)
			fmt.Println(" done")

			// Load config for pin list (optional)
			var pinned map[string]struct{}
			if *cfgFile != "" {
				cfg, err := config.Load(*cfgFile)
				if err != nil {
					return err
				}
				pinned = toSet(cfg.Pin)
			} else {
				cfg, err := config.Load("")
				if err == nil {
					pinned = toSet(cfg.Pin)
				}
			}

			outdatedBrews, err := brew.GetOutdated(r)
			if err != nil {
				return fmt.Errorf("querying outdated formulae: %w", err)
			}
			outdatedCasks, err := brew.GetOutdatedCasks(r)
			if err != nil {
				return fmt.Errorf("querying outdated casks: %w", err)
			}

			// Filter out pinned
			brewPkgs := filterPinnedPkgs(outdatedBrews, pinned)
			caskPkgs := filterPinnedPkgs(outdatedCasks, pinned)

			if len(brewPkgs) == 0 && len(caskPkgs) == 0 {
				fmt.Println("All packages are up to date.")
				return nil
			}

			// Print upgrade plan
			printUpgradePlanV2(brewPkgs, caskPkgs)

			if !yes {
				if !confirm(os.Stdin) {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			// Collect names for upgrade execution
			brewNames := pkgNames(brewPkgs)
			caskNames := pkgNames(caskPkgs)

			// Pre-download phase (non-verbose only)
			if !runner.Verbose {
				dlConcurrency := concurrency
				if dlConcurrency < 1 {
					dlConcurrency = 1
				}
				platform := download.Platform(r)
				urls := download.GetFormulaURLs(r, brewNames, platform)
				caskURLs := download.GetCaskURLs(r, caskNames)
				urls = append(urls, caskURLs...)

				if len(urls) > 0 {
					dd := download.NewDownloadDisplay(dlConcurrency)
					stop := dd.Start()

					results := download.PreDownload(urls, download.Options{
						Concurrency: dlConcurrency,
						OnProgress: func(name string, downloaded, total int64) {
							dd.UpdateProgress(name, downloaded, total)
						},
						OnComplete: func(res download.Result) {
							dd.MarkDone(res.Package.Name, res.Err)
						},
					})

					stop()

					for _, res := range results {
						if res.Err != nil {
							fmt.Printf("  ⚠ %s: %v\n", res.Package.Name, res.Err)
						}
					}
				}
			}

			fmt.Printf("\nUpgrading (%d packages)...\n", len(brewNames)+len(caskNames))
			var results []brew.Result
			total := len(brewNames) + len(caskNames)

			if runner.Verbose {
				for i, name := range brewNames {
					fmt.Printf("  [%d/%d] %s\n", i+1, total, name)
					_, err := r.RunMutate("brew", "upgrade", name)
					res := brew.Result{Action: "upgrade", Package: name, Err: err}
					results = append(results, res)
					if err != nil {
						fmt.Printf("  %s✗%s %s\n", Red, Reset, name)
					} else {
						fmt.Printf("  %s✓%s %s\n", Green, Reset, name)
					}
				}
				for i, name := range caskNames {
					fmt.Printf("  [%d/%d] %s\n", len(brewNames)+i+1, total, name)
					_, err := r.RunMutate("brew", "upgrade", "--cask", name)
					res := brew.Result{Action: "upgrade-cask", Package: name, Err: err}
					results = append(results, res)
					if err != nil {
						fmt.Printf("  %s✗%s %s\n", Red, Reset, name)
					} else {
						fmt.Printf("  %s✓%s %s\n", Green, Reset, name)
					}
				}
			} else {
				display := progress.NewDisplay()
				for i, name := range brewNames {
					idx := i + 1
					output, err := runWithSpinner(r, display, idx, total, name, "brew", "upgrade", name)
					res := brew.Result{Action: "upgrade", Package: name, Err: err}
					results = append(results, res)
					if err != nil {
						errMsg := progress.ExtractError(output)
						display.Finish("  %s✗%s %s: %s", Red, Reset, name, errMsg)
					} else {
						display.Finish("  %s✓%s %s", Green, Reset, name)
					}
				}
				for i, name := range caskNames {
					idx := len(brewNames) + i + 1
					output, err := runWithSpinner(r, display, idx, total, name, "brew", "upgrade", "--cask", name)
					res := brew.Result{Action: "upgrade-cask", Package: name, Err: err}
					results = append(results, res)
					if err != nil {
						errMsg := progress.ExtractError(output)
						display.Finish("  %s✗%s %s: %s", Red, Reset, name, errMsg)
					} else {
						display.Finish("  %s✓%s %s", Green, Reset, name)
					}
				}
			}

			if brew.HasErrors(results) {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompt")
	cmd.Flags().IntVar(&concurrency, "concurrency", 4, "parallel download workers")
	return cmd
}

func printUpgradePlanV2(brews, casks []brew.OutdatedPkg) {
	total := len(brews) + len(casks)
	fmt.Printf("\n%d package%s to upgrade", total, plural(total))
	if len(brews) > 0 && len(casks) > 0 {
		fmt.Printf(" (%d formulae, %d casks)", len(brews), len(casks))
	}
	fmt.Println()

	// Security packages first
	var secBrews, normalBrews []brew.OutdatedPkg
	for _, p := range brews {
		if securityPkgs[p.Name] {
			secBrews = append(secBrews, p)
		} else {
			normalBrews = append(normalBrews, p)
		}
	}

	if len(secBrews) > 0 {
		fmt.Printf("\n%s⚠ Security:%s\n", Yellow, Reset)
		for _, p := range secBrews {
			printPkgLine(p)
		}
	}

	if len(normalBrews) > 0 {
		fmt.Println("\nFormulae:")
		for _, p := range normalBrews {
			printPkgLine(p)
		}
	}

	if len(casks) > 0 {
		fmt.Println("\nCasks:")
		for _, p := range casks {
			printPkgLine(p)
		}
	}
}

func printPkgLine(p brew.OutdatedPkg) {
	if p.Current != "" && p.Latest != "" {
		fmt.Printf("  %s~%s %s %s → %s\n", Yellow, Reset, p.Name, p.Current, p.Latest)
	} else {
		fmt.Printf("  %s~%s %s\n", Yellow, Reset, p.Name)
	}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func pkgNames(pkgs []brew.OutdatedPkg) []string {
	names := make([]string, len(pkgs))
	for i, p := range pkgs {
		names[i] = p.Name
	}
	return names
}

func toSet(items []string) map[string]struct{} {
	s := make(map[string]struct{}, len(items))
	for _, item := range items {
		s[item] = struct{}{}
	}
	return s
}

func filterPinnedPkgs(pkgs []brew.OutdatedPkg, pinned map[string]struct{}) []brew.OutdatedPkg {
	if len(pinned) == 0 {
		return pkgs
	}
	var result []brew.OutdatedPkg
	for _, p := range pkgs {
		if _, ok := pinned[p.Name]; !ok {
			result = append(result, p)
		}
	}
	return result
}

// filterPinned filters string slices (kept for other callers).
func filterPinned(items []string, pinned map[string]struct{}) []string {
	if len(pinned) == 0 {
		return items
	}
	var result []string
	for _, item := range items {
		if _, ok := pinned[strings.TrimSpace(item)]; !ok {
			result = append(result, item)
		}
	}
	return result
}

var spinnerFrames = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

// runWithSpinner runs a command while showing a spinner, returns captured output and error.
func runWithSpinner(r runner.CommandRunner, display *progress.Display, idx, total int, name, cmd string, args ...string) ([]byte, error) {
	done := make(chan struct{})
	go func() {
		i := 0
		for {
			select {
			case <-done:
				return
			case <-time.After(100 * time.Millisecond):
				display.Status("  [%d/%d] %s %c", idx, total, name, spinnerFrames[i%len(spinnerFrames)])
				i++
			}
		}
	}()
	output, err := r.Run(cmd, args...)
	close(done)
	return output, err
}
