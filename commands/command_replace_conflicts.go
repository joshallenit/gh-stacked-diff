package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"strings"

	"github.com/joshallenit/gh-stacked-diff/v2/templates"

	"github.com/joshallenit/gh-stacked-diff/v2/interactive"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

func createReplaceConflictsCommand() Command {
	flagSet := flag.NewFlagSet("replace-conflicts", flag.ContinueOnError)
	confirmed := flagSet.Bool("confirm", false, "Whether to automatically confirm to do this rather than ask for y/n input")
	return Command{
		FlagSet: flagSet,
		Summary: "For failed rebase: replace changes with its associated branch",
		Description: "During a rebase that failed because of merge conflicts, replace the\n" +
			"current uncommitted changes (merge conflicts), with the contents\n" +
			"(diff between origin/" + util.GetMainBranchForHelp() + " and HEAD) of its associated branch.",
		Usage: "sd " + flagSet.Name(),
		OnSelected: func(asyncConfig util.AsyncAppConfig, command Command) {
			if flagSet.NArg() > 0 {
				commandError(asyncConfig.App, flagSet, "too many arguments", command.Usage)
			}
			replaceConflicts(asyncConfig.App, *confirmed)
		}}
}

// For failed rebase: replace changes with its associated branch.
func replaceConflicts(appConfig util.AppConfig, confirmed bool) {
	commitWithConflicts := getCommitWithConflicts()
	gitLog := templates.GetBranchInfo(commitWithConflicts, templates.IndicatorTypeCommit)
	checkConfirmed(appConfig, confirmed)
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "reset", "--hard", "HEAD")
	slog.Info(fmt.Sprint("Replacing changes (merge conflicts) for failed rebase of commit ", commitWithConflicts, ", with changes from associated branch, ", gitLog.Branch))
	diff := util.ExecuteOrDie(util.ExecuteOptions{}, "git", "diff", "--binary", "origin/"+util.GetMainBranchOrDie(), gitLog.Branch)
	util.ExecuteOrDie(util.ExecuteOptions{Io: util.StdIo{In: strings.NewReader(diff), Out: nil, Err: nil}},
		"git", "apply",
	)
	slog.Info("Adding changes and continuing rebase")
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "add", ".")
	continueOptions := util.ExecuteOptions{EnvironmentVariables: []string{"GIT_EDITOR=true"}}
	util.ExecuteOrDie(continueOptions, "git", "rebase", "--continue")
}

func getCommitWithConflicts() string {
	statusLines := strings.Split(util.ExecuteOrDie(util.ExecuteOptions{}, "git", "status"), "\n")
	lastCommandDoneLine := -1
	inLast := false
	for i, line := range statusLines {
		if strings.HasPrefix(line, "Last ") {
			// find last pick line
			inLast = true
		} else if inLast {
			if strings.HasPrefix(line, "   ") {
				lastCommandDoneLine = i
			} else {
				break
			}
		}
	}
	if lastCommandDoneLine == -1 {
		panic("Cannot determine which commit is being rebased with because \"git status\" does not have a \"Last commands done\" section. To use this command you must be in the middle of a rebase")
	}
	// Return the 2nd field, from a string such as "pick f52e867 next1"
	return strings.Fields(statusLines[lastCommandDoneLine])[1]
}

func checkConfirmed(appConfig util.AppConfig, confirmed bool) {
	if confirmed {
		return
	}
	interactive.ConfirmOrDie(appConfig, "This will clear any uncommitted changes, are you sure (y/n)?")
}
