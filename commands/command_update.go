package commands

import (
	"flag"
	"io"
	"log/slog"
	"strings"

	"fmt"

	ex "github.com/joshallenit/gh-stacked-diff/v2/execute"
	"github.com/joshallenit/gh-stacked-diff/v2/interactive"
	"github.com/joshallenit/gh-stacked-diff/v2/templates"
	"github.com/joshallenit/gh-stacked-diff/v2/util"

	"time"
)

func createUpdateCommand() Command {
	flagSet := flag.NewFlagSet("update", flag.ContinueOnError)
	indicatorTypeString := addIndicatorFlag(flagSet)
	reviewers, silent, minChecks := addReviewersFlags(flagSet, "")
	return Command{
		FlagSet: flagSet,
		Summary: "Add commits from " + util.GetMainBranchForHelp() + " to an existing PR",
		Description: "Add commits from local " + util.GetMainBranchForHelp() + " branch to an existing PR.\n" +
			"\n" +
			"Can also add reviewers once PR checks have passed, see \"--reviewers\" flag.",
		Usage: "sd " + flagSet.Name() + " [flags] <PR commitIndicator> [fixup commitIndicator (defaults to head commit) [fixup commitIndicator...]]",
		OnSelected: func(command Command, stdOut io.Writer, stdErr io.Writer, sequenceEditorPrefix string, exit func(err any)) {
			indicatorType := checkIndicatorFlag(command, indicatorTypeString)
			destCommit := getDestCommit(flagSet, command, indicatorType, exit)
			commitsToCherryPick := getCommitsToCherryPick(flagSet, command, indicatorType, exit)
			updatePr(destCommit, commitsToCherryPick, indicatorType, sequenceEditorPrefix)
			if *reviewers != "" {
				addReviewersToPr([]string{destCommit.Commit}, templates.IndicatorTypeCommit, true, *silent, *minChecks, *reviewers, 30*time.Second)
			}
		}}
}

func getDestCommit(flagSet *flag.FlagSet, command Command, indicatorType templates.IndicatorType, exit func(any)) templates.GitLog {
	if flagSet.NArg() == 0 {
		var err error
		destCommit, err := interactive.GetPrSelection("What PR do you want to update?")
		if err != nil {
			if err == interactive.UserCancelled {
				exit(nil)
			} else {
				commandError(flagSet, err.Error(), command.Usage)
			}
		}
		return destCommit
	} else {
		return templates.GetBranchInfo(flagSet.Arg(0), indicatorType)
	}
}

func getCommitsToCherryPick(flagSet *flag.FlagSet, command Command, indicatorType templates.IndicatorType, exit func(any)) []string {
	if flagSet.NArg() < 2 {
		selectedCommits, err := interactive.GetCommitSelection("What commits do you want to add?")
		if err != nil {
			if err == interactive.UserCancelled {
				exit(nil)
			} else {
				commandError(flagSet, err.Error(), command.Usage)
			}
		}
		return util.MapSlice(selectedCommits, func(commit templates.GitLog) string {
			return commit.Commit
		})
	} else {
		return commitIndicatorsToCommitHashes(flagSet.Args()[1:], indicatorType)
	}
}

// Add commits from main to an existing PR.
func updatePr(destCommit templates.GitLog, commitsToCherryPick []string, indicatorType templates.IndicatorType, sequenceEditorPrefix string) {
	util.RequireMainBranch()
	templates.RequireCommitOnMain(destCommit.Commit)
	shouldPopStash := util.Stash("before update-pr " + destCommit.Commit)
	rollbackManager := &util.GitRollbackManager{}
	rollbackManager.SaveState()
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", destCommit.Branch)
	rollbackManager.SaveState() // Save state again for associated branch.
	defer func() {
		r := recover()
		if r != nil {
			rollbackManager.Restore(r)
		}
		util.PopStash(shouldPopStash)
		if r != nil {
			panic(r)
		}
	}()
	slog.Info("Fast forwarding in case there were any commits made via github web interface")
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "fetch", "origin", destCommit.Branch)
	forcePush := false
	if _, err := ex.Execute(ex.ExecuteOptions{Output: ex.NewStandardOutput()}, "git", "merge", "--ff-only", "origin/"+destCommit.Branch); err != nil {
		slog.Info(fmt.Sprint("Could not fast forward to match origin. Rebasing instead. ", err))
		ex.ExecuteOrDie(ex.ExecuteOptions{Output: ex.NewStandardOutput()}, "git", "rebase", "origin", destCommit.Branch)
		// As we rebased, a force push may be required.
		forcePush = true
	}

	slog.Info(fmt.Sprint("Cherry picking ", commitsToCherryPick))
	cherryPickArgs := make([]string, 1+len(commitsToCherryPick))
	cherryPickArgs[0] = "cherry-pick"
	for i, commit := range commitsToCherryPick {
		cherryPickArgs[i+1] = commit
	}
	_, cherryPickError := ex.Execute(ex.ExecuteOptions{}, "git", cherryPickArgs...)
	if cherryPickError != nil {
		slog.Info("First attempt at cherry-pick failed")
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "cherry-pick", "--abort")
		rebaseCommit := util.FirstOriginMainCommit(util.GetMainBranchOrDie())
		slog.Info(fmt.Sprint("Rebasing with the base commit on "+util.GetMainBranchOrDie()+" branch, ", rebaseCommit,
			", in case the local "+util.GetMainBranchOrDie()+" was rebased with origin/"+util.GetMainBranchOrDie()))
		ex.ExecuteOrDie(ex.ExecuteOptions{Output: ex.NewStandardOutput()}, "git", "rebase", rebaseCommit)
		slog.Info(fmt.Sprint("Cherry picking again ", commitsToCherryPick))
		ex.ExecuteOrDie(ex.ExecuteOptions{Output: ex.NewStandardOutput()}, "git", cherryPickArgs...)
		forcePush = true
	}
	slog.Info("Pushing to remote")
	if forcePush {
		if _, err := ex.Execute(ex.ExecuteOptions{}, "git", "push", "origin", destCommit.Branch); err != nil {
			slog.Info("Regular push failed, force pushing instead.")
			ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "-f", "origin", destCommit.Branch)
		}
	} else {
		ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "push", "origin", destCommit.Branch)
	}
	slog.Info("Switching back to " + util.GetMainBranchOrDie())
	ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "switch", util.GetMainBranchOrDie())
	slog.Info(fmt.Sprint("Rebasing, marking as fixup ", commitsToCherryPick, " for target ", destCommit.Commit))
	environmentVariables := []string{
		"GIT_SEQUENCE_EDITOR=" + sequenceEditorPrefix + "sequence-editor-mark-as-fixup " +
			destCommit.Commit + " " +
			strings.Join(commitsToCherryPick, " "),
	}
	slog.Debug(fmt.Sprint("Using sequence editor ", environmentVariables))
	options := ex.ExecuteOptions{EnvironmentVariables: environmentVariables, Output: ex.NewStandardOutput()}
	ex.ExecuteOrDie(options, "git", "rebase", "-i", destCommit.Commit+"^")
	rollbackManager.Clear()
}

func commitIndicatorsToCommitHashes(otherCommits []string, indicatorType templates.IndicatorType) []string {
	var commitsToCherryPick []string
	if len(otherCommits) > 0 {
		if indicatorType == templates.IndicatorTypeGuess || indicatorType == templates.IndicatorTypeList {
			commitsToCherryPick = util.MapSlice(otherCommits, func(commit string) string {
				return templates.GetBranchInfo(commit, indicatorType).Commit
			})
		} else {
			commitsToCherryPick = otherCommits
		}
	} else {
		commitsToCherryPick = make([]string, 1)
		commitsToCherryPick[0] = strings.TrimSpace(ex.ExecuteOrDie(ex.ExecuteOptions{}, "git", "rev-parse", "--short", "HEAD"))
	}
	return commitsToCherryPick
}
