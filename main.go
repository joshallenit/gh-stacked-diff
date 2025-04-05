/*
Stacked Diff Workflow Command Line Interface

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
	"fmt"
	"os"
	"strings"

	"github.com/joshallenit/gh-stacked-diff/v2/commands"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

func main() {
	appConfig := util.AppConfig{
		Io:            util.StdIo{Out: os.Stdout, Err: os.Stderr, In: os.Stdin},
		AppExecutable: getAppExecutable(),
		Exit:          os.Exit,
	}

	commands.ExecuteCommand(appConfig, os.Args[1:])
}

func getAppExecutable() string {
	appExecutable, err := os.Executable()
	if err != nil {
		panic(fmt.Sprint("Cannot determine executable ", err))
	}
	// Escape back slashes so the exectuable works from GitBash on Windows
	appExecutable = strings.ReplaceAll(appExecutable, "\\", "\\\\")
	// Quote in case the path has a space.
	return "\"" + appExecutable + "\""
}
