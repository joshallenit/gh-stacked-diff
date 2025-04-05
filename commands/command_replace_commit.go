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

func createReplaceCommitCommand() Command {
	flagSet := flag.NewFlagSet("replace-commit", flag.ContinueOnError)
	indicatorTypeString := addIndicatorFlag(flagSet)
	return Command{
		FlagSet: flagSet,
		Summary: "Replaces a commit on " + util.GetMainBranchForHelp() + " branch with its associated branch",
		Description: "Replaces a commit on " + util.GetMainBranchForHelp() + " branch with the squashed contents of its\n" +
			"associated branch.\n" +
			"\n" +
			"This is useful when you make changes within a branch, for example to\n" +
			"fix a problem found on CI, and want to bring the changes over to your\n" +
			"local " + util.GetMainBranchForHelp() + " branch.",
		Usage: "sd " + flagSet.Name() + " [flags] <commitIndicator>",
		OnSelected: func(appConfig util.AppConfig, command Command) {
			if flagSet.NArg() > 1 {
				commandError(appConfig, flagSet, "too many arguments", command.Usage)
			}
			selectCommitOptions := interactive.CommitSelectionOptions{
				Prompt:      "What commit do you want to replace with the contents of its associated branch?",
				CommitType:  interactive.CommitTypePr,
				MultiSelect: false,
			}
			targetCommit := getTargetCommits(appConfig, command, []string{flagSet.Arg(0)}, indicatorTypeString, selectCommitOptions)
			replaceCommit(targetCommit[0])
		}}
}

// Replaces a commit on main branch with its associated branch.
func replaceCommit(targetCommit templates.GitLog) {
	util.RequireMainBranch()
	templates.RequireCommitOnMain(targetCommit.Commit)
	shouldPopStash := util.Stash("replace-commit " + targetCommit.Commit + " " + targetCommit.Subject)
	replaceCommitOfBranchInfo(targetCommit)
	util.PopStash(shouldPopStash)
}

// Replaces commit `gitLog.Commit“ with the contents of branch `gitLog.Branch`
func replaceCommitOfBranchInfo(gitLog templates.GitLog) {
	commitsAfter := strings.Fields(strings.TrimSpace(util.ExecuteOrDie(util.ExecuteOptions{}, "git", "--no-pager", "log", gitLog.Commit+"..HEAD", "--pretty=format:%h")))
	reverseArrayInPlace(commitsAfter)
	commitToDiffFrom := util.FirstOriginMainCommit(gitLog.Branch)
	slog.Info("Resetting to " + gitLog.Commit + "~1")
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "reset", "--hard", gitLog.Commit+"~1")
	slog.Info("Adding diff from commits " + gitLog.Branch)
	diff := util.ExecuteOrDie(util.ExecuteOptions{}, "git", "diff", "--binary", commitToDiffFrom, gitLog.Branch)
	util.ExecuteOrDie(
		util.ExecuteOptions{Io: util.StdIo{In: strings.NewReader(diff), Out: nil, Err: nil}},
		"git", "apply",
	)
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "add", ".")
	commitSummary := util.ExecuteOrDie(util.ExecuteOptions{}, "git", "--no-pager", "show", "--no-patch", "--format=%s", gitLog.Commit)
	util.ExecuteOrDie(util.ExecuteOptions{}, "git", "commit", "-m", strings.TrimSpace(commitSummary))
	if len(commitsAfter) != 0 {
		slog.Info(fmt.Sprint("Cherry picking commits back on top ", commitsAfter))
		cherryPickAndSkipAllEmpty(commitsAfter)
	}
}

func reverseArrayInPlace(array []string) {
	for i, j := 0, len(array)-1; i < j; i, j = i+1, j-1 {
		array[i], array[j] = array[j], array[i]
	}
}

func cherryPickAndSkipAllEmpty(commits []string) {
	cherryPickArgs := make([]string, 2+len(commits))
	cherryPickArgs[0] = "cherry-pick"
	cherryPickArgs[1] = "--ff"
	for i, commit := range commits {
		cherryPickArgs[i+2] = commit
	}
	out, err := util.Execute(util.ExecuteOptions{}, "git", cherryPickArgs...)
	for err != nil {
		if strings.Contains(out, "git commit --allow-empty") {
			out, err = util.Execute(util.ExecuteOptions{}, "git", "cherry-pick", "--skip")
		} else {
			panic(fmt.Sprint("Unexpected cherry-pick error", out, cherryPickArgs, err))
		}
	}
}
