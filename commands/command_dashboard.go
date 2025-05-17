package commands

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/joshallenit/gh-stacked-diff/v2/interactive"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

func createDashboardCommand() Command {
	flagSet := flag.NewFlagSet("dashboard", flag.ContinueOnError)
	minChecks := flagSet.Int("min-checks", -1, "Minimum number of checks that must pass for a PR to be considered passing")
	watch := flagSet.Bool("watch", true, "Watch for changes in the .git directory and refresh automatically")

	return Command{
		FlagSet: flagSet,
		Summary: "Displays an interactive dashboard of your stacked changes",
		Description: "Shows an interactive dashboard view of your commits and their associated PRs.\n" +
			"\n" +
			"The dashboard displays information about:\n" +
			"- PR status\n" +
			"- CI checks\n" +
			"- Approvals\n" +
			"- Commit information\n" +
			"\n" +
			"By default, the dashboard will refresh automatically when changes are detected\n" +
			"in the .git directory. Use --watch=false to disable this behavior.",
		Usage:           "sd " + flagSet.Name() + " [flags]",
		DefaultLogLevel: slog.LevelError,
		OnSelected: func(asyncConfig util.AsyncAppConfig, command Command) {
			if flagSet.NArg() != 0 {
				commandError(asyncConfig.App, flagSet, "too many arguments", command.Usage)
			}

			// If watch is enabled, set up file system watcher
			if *watch {
				watcher, err := fsnotify.NewWatcher()
				if err != nil {
					slog.Error("Failed to create file system watcher", "error", err)
					interactive.ShowDashboard(asyncConfig, *minChecks, context.Background())
					return
				}
				defer watcher.Close()

				// Find .git directory
				gitDir, err := findGitDir()
				if err != nil {
					slog.Error("Failed to find .git directory", "error", err)
					interactive.ShowDashboard(asyncConfig, *minChecks, context.Background())
					return
				}

				// Watch .git directory
				err = watcher.Add(gitDir)
				if err != nil {
					slog.Error("Failed to watch .git directory", "error", err)
					interactive.ShowDashboard(asyncConfig, *minChecks, context.Background())
					return
				}

				// Create a debounced refresh function with context cancellation
				var lastRefresh time.Time
				var currentCtx context.Context
				var currentCancel context.CancelFunc

				refreshDebounced := func() {
					now := time.Now()
					if now.Sub(lastRefresh) < time.Second {
						return // Debounce frequent updates
					}
					lastRefresh = now

					// Cancel previous dashboard if it exists
					if currentCancel != nil {
						currentCancel()
					}

					// Create new context for this dashboard instance
					currentCtx, currentCancel = context.WithCancel(context.Background())
					interactive.ShowDashboard(asyncConfig, *minChecks, currentCtx)
				}

				// Initial display
				refreshDebounced()

				// Watch for events
				for {
					select {
					case event, ok := <-watcher.Events:
						if !ok {
							return
						}
						if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
							refreshDebounced()
						}
					case err, ok := <-watcher.Errors:
						if !ok {
							return
						}
						slog.Error("Watcher error", "error", err)
					}
				}
			} else {
				// Just show the dashboard once if watch is disabled
				interactive.ShowDashboard(asyncConfig, *minChecks, context.Background())
			}
		}}
}

// findGitDir looks for the .git directory starting from the current directory
// and moving up through parent directories
func findGitDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return gitPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
