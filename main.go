/*
Stacked Diff Workflow

usage: sd [top-level-flags] <command> [<args>]

Possible commands are:

	add-reviewers       Add reviewers to Pull Request on Github once its checks have passed
	branch-name         Outputs branch name of commit
	checkout            Checks out branch associated with commit indicator
	code-owners         Outputs code owners for all of the changes in branch
	log                 Displays git log of your changes
	new                 Create a new pull request from a commit on main
	prs                 Lists all Pull Requests you have open.
	rebase-main         Bring your main branch up to date with remote
	replace-commit      Replaces a commit on main branch with its associated branch
	replace-conflicts   For failed rebase: replace changes with its associated branch
	update              Add commits from main to an existing PR
	wait-for-merge      Waits for a pull request to be merged

To learn more about a command use: sd <command> --help

flags:

	-log-level string
	      Possible log levels:
	         debug
	         info
	         warn
	         error
	      Default is info, except on commands that are for output purposes,
	      (namely branch-name and log), which have a default of error.
*/
package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/joshallenit/stacked-diff/util"
)

/*
sd - stacked diff -
*/
func main() {
	parseArguments(os.Stdout, os.Stderr, flag.NewFlagSet("sd", flag.ContinueOnError), os.Args[1:], defaultExit)
}

func defaultExit(stdErr io.Writer, errorCode int, logLevelVar *slog.LevelVar, err any) {
	// Show panic stack trace during debug log level.
	if logLevelVar.Level() <= slog.LevelDebug {
		panic(err)
	} else {
		fmt.Fprintln(stdErr, fmt.Sprint(util.Reset, "error: ", err))
		os.Exit(1)
	}
}
