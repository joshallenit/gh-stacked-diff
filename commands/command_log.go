package commands

import (
	"flag"
	"io"
	"log/slog"

	"fmt"
	"slices"
	"strings"

	"github.com/fatih/color"

	ex "github.com/joshallenit/gh-stacked-diff/v2/execute"
	"github.com/joshallenit/gh-stacked-diff/v2/templates"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

func createLogCommand() Command {
	flagSet := flag.NewFlagSet("log", flag.ContinueOnError)

	return Command{
		FlagSet: flagSet,
		Summary: "Displays git log of your changes",
		Description: "Displays summary of the git commits on current branch that are not\n" +
			"in the remote branch.\n" +
			"\n" +
			"Useful to view list indexes, or copy commit hashes, to use for the\n" +
			"commitIndicator required by other commands.\n" +
			"\n" +
			"A " + color.GreenString("✅") + " means that there is a PR associated with the commit (actually it\n" +
			"means there is a branch, but having a branch means there is a PR when\n" +
			"using this workflow). If there is more than one commit on the\n" +
			"associated branch, those commits are also listed (indented under the\n" +
			"their associated commit summary).",
		Usage:           "sd " + flagSet.Name(),
		DefaultLogLevel: slog.LevelError,
		OnSelected: func(command Command, stdOut io.Writer, stdErr io.Writer, stdIn io.Reader, sequenceEditorPrefix string, exit func(err any)) {
			if flagSet.NArg() != 0 {
				commandError(flagSet, "too many arguments", command.Usage)
			}
			printGitLog(stdOut)
		}}
}

// Prints changes in the current branch compared to the main branch to out.
func printGitLog(out io.Writer) {
	if util.GetCurrentBranchName() != util.GetMainBranchOrDie() {
		gitArgs := []string{"--no-pager", "log", "--pretty=oneline", "--abbrev-commit"}
		if util.RemoteHasBranch(util.GetMainBranchOrDie()) {
			gitArgs = append(gitArgs, "origin/"+util.GetMainBranchOrDie()+"..HEAD")
		}
		gitArgs = append(gitArgs, "--color=always")
		ex.ExecuteOrDie(ex.ExecuteOptions{Output: &ex.ExecutionOutput{Stdout: out, Stderr: out}}, "git", gitArgs...)
		return
	}
	logs := templates.GetNewCommits("HEAD")
	gitBranchArgs := make([]string, 0, len(logs)+2)
	gitBranchArgs = append(gitBranchArgs, "branch", "-l")
	for _, log := range logs {
		gitBranchArgs = append(gitBranchArgs, log.Branch)
	}
	checkedBranches := strings.Fields(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", gitBranchArgs...))
	for i, log := range logs {
		numberPrefix := getNumberPrefix(i, len(logs))
		if slices.Contains(checkedBranches, log.Branch) {
			// Use color for ✅ otherwise in Git Bash on Windows it will appear as black and white.
			fmt.Fprint(out, numberPrefix+color.GreenString("✅ "))
		} else {
			fmt.Fprint(out, numberPrefix+"   ")
		}
		fmt.Fprintln(out, color.YellowString(log.Commit)+" "+log.Subject)
		// find first commit that is not in main branch
		if slices.Contains(checkedBranches, log.Branch) {
			branchCommits := templates.GetNewCommits(log.Branch)
			if len(branchCommits) > 1 {
				for _, branchCommit := range branchCommits {
					padding := strings.Repeat(" ", len(numberPrefix))
					fmt.Fprintln(out, padding+"   - "+branchCommit.Subject)
				}
			}
		}
	}
}

func getNumberPrefix(i int, numLogs int) string {
	maxIndex := fmt.Sprint(numLogs)
	currentIndex := fmt.Sprint(i + 1)
	padding := strings.Repeat(" ", len(maxIndex)-len(currentIndex))
	return padding + currentIndex + ". "
}
