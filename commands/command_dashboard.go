package commands

import (
	"flag"
	"log/slog"

	"github.com/joshallenit/gh-stacked-diff/v2/interactive"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

func createDashboardCommand() Command {
	flagSet := flag.NewFlagSet("dashboard", flag.ContinueOnError)
	minChecks := flagSet.Int("min-checks", -1, "Minimum number of checks that must pass for a PR to be considered passing")

	return Command{
		FlagSet: flagSet,
		Summary: "Displays an interactive dashboard of your stacked changes",
		Description: "Shows an interactive dashboard view of your commits and their associated PRs.\n" +
			"\n" +
			"The dashboard displays information about:\n" +
			"- PR status\n" +
			"- CI checks\n" +
			"- Approvals\n" +
			"- Commit information",
		Usage:           "sd " + flagSet.Name() + " [flags]",
		DefaultLogLevel: slog.LevelError,
		OnSelected: func(asyncConfig util.AsyncAppConfig, command Command) {
			if flagSet.NArg() != 0 {
				commandError(asyncConfig.App, flagSet, "too many arguments", command.Usage)
			}
			interactive.ShowDashboard(asyncConfig, *minChecks)
		}}
}
