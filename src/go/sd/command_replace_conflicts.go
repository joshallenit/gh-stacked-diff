package main

import (
	"flag"
	sd "stackeddiff"
	ex "stackeddiff/execute"
)

func CreateReplaceConflictsCommand() Command {
	flagSet := flag.NewFlagSet("replace-conflicts", flag.ContinueOnError)
	return Command{
		FlagSet:     flagSet,
		Summary:     "For failed rebase: replace changes with its associated branch",
		Description: "During a rebase that failed becuase of merge conflicts, replace the current uncommitted changes (merge conflicts), with the contents (diff between origin/" + ex.GetMainBranch() + " and HEAD) of its associated branch.",
		Usage:       "sd replace-conflicts",
		OnSelected: func(command Command) {
			if flagSet.NArg() > 0 {
				commandError(flagSet, "too many arguments", command.Usage)
			}
			sd.ReplaceConflicts()
		}}
}
