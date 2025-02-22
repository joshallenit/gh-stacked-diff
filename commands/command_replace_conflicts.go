package commands

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	ex "github.com/joshallenit/stacked-diff/v2/execute"
	"github.com/joshallenit/stacked-diff/v2/templates"

	"github.com/joshallenit/stacked-diff/v2/util"
)

func createReplaceConflictsCommand(stdOut io.Writer) Command {
	flagSet := flag.NewFlagSet("replace-conflicts", flag.ContinueOnError)
	confirmed := flagSet.Bool("confirm", false, "Whether to automatically confirm to do this rather than ask for y/n input")
	return Command{
		FlagSet: flagSet,
		Summary: "For failed rebase: replace changes with its associated branch",
		Description: "During a rebase that failed because of merge conflicts, replace the\n" +
			"current uncommitted changes (merge conflicts), with the contents\n" +
			"(diff between origin/" + util.GetMainBranchForHelp() + " and HEAD) of its associated branch.",
		Usage: "sd " + flagSet.Name(),
		OnSelected: func(command Command) {
			if flagSet.NArg() > 0 {
				commandError(flagSet, "too many arguments", command.Usage)
			}
			replaceConflicts(stdOut, *confirmed)
		}}
}

// For failed rebase: replace changes with its associated branch.
func replaceConflicts(stdOut io.Writer, confirmed bool) {
	commitWithConflicts := getCommitWithConflicts()
	gitLog := templates.GetBranchInfo(commitWithConflicts, templates.IndicatorTypeCommit)
	checkConfirmed(stdOut, confirmed)
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "reset", "--hard", "HEAD")
	slog.Info(fmt.Sprint("Replacing changes (merge conflicts) for failed rebase of commit ", commitWithConflicts, ", with changes from associated branch, ", gitLog.Branch))
	diff := ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "diff", "--binary", "origin/"+util.GetMainBranchOrDie(), gitLog.Branch)
	ex.ExecuteOrDie(ex.ExecuteOptions{Stdin: &diff}, "git", "apply")
	slog.Info("Adding changes and continuing rebase")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "add", ".")
	continueOptions := ex.ExecuteOptions{EnvironmentVariables: []string{"GIT_EDITOR=true"}}
	ex.ExecuteOrDie(continueOptions, "git", "rebase", "--continue")
}

func getCommitWithConflicts() string {
	statusLines := strings.Split(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "status"), "\n")
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

func checkConfirmed(stdOut io.Writer, confirmed bool) {
	if confirmed {
		return
	}

	fmt.Fprint(stdOut, "This will clear any uncommitted changes, are you sure (y/n)? ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.ToLower(scanner.Text())
	slog.Debug(fmt.Sprint("Got input ", input))
	if input != "y" && input != "yes" {
		slog.Info("Cancelled by user")
		os.Exit(0)
	}
}
