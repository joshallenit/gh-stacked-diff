package commands

import (
	"flag"
	"io"

	ex "github.com/joshallenit/gh-stacked-diff/v2/execute"
	"github.com/joshallenit/gh-stacked-diff/v2/templates"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

func createCheckoutCommand() Command {
	flagSet := flag.NewFlagSet("checkout", flag.ContinueOnError)
	indicatorTypeString := addIndicatorFlag(flagSet)
	return Command{
		FlagSet: flagSet,
		Summary: "Checks out branch associated with commit indicator",
		Description: "Checks out the branch associated with commit indicator.\n" +
			"\n" +
			"For when you want to merge only the branch with with origin/" + util.GetMainBranchForHelp() + ",\n" +
			"rather than your entire local " + util.GetMainBranchForHelp() + " branch, verify why \n" +
			"CI is failing on that particular branch, or for any other reason.\n" +
			"\n" +
			"After modifying the branch you can use \"sd replace-commit\" to sync local " + util.GetMainBranchForHelp() + ".",
		Usage: "sd " + flagSet.Name() + " [flags] <commitIndicator>",
		OnSelected: func(command Command, stdOut io.Writer, stdErr io.Writer, sequenceEditorPrefix string, exit func(err any)) {
			if flagSet.NArg() == 0 {
				commandError(flagSet, "missing commitIndicator", command.Usage)
			}
			if flagSet.NArg() > 1 {
				commandError(flagSet, "too many arguments", command.Usage)
			}
			indicatorType := checkIndicatorFlag(command, indicatorTypeString)
			branchName := templates.GetBranchInfo(flagSet.Arg(0), indicatorType).Branch
			ex.ExecuteOrDie(ex.ExecuteOptions{Output: ex.NewStandardOutput()}, "git", "checkout", branchName)
		}}
}
