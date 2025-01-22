package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"slices"
	sd "stackeddiff"
	ex "stackeddiff/execute"
)

type Command struct {
	flagSet    *flag.FlagSet
	onSelected func()
}

/*
Outputs abbreviated git log that only shows what has changed, useful for copying commit hashes.
Adds a checkmark beside commits that have an associated branch.
*/
func main() {
	ParseArguments(os.Stdout, os.Args[1:])
}

func ParseArguments(stdOut io.Writer, commandLineArgs []string) {

	// Parse flags common for every command.
	var logFlags int
	var logLevelString string

	flag.IntVar(&logFlags, "log-flags", 0, "Log flags, see https://pkg.go.dev/log#pkg-constants")
	flag.StringVar(&logLevelString, "log-level", "info", "Log level: debug, info, warn, or error")

	flag.CommandLine.Parse(commandLineArgs)

	commands := []Command{createLogCommand(stdOut), createNewCommand(), createUpdateCommand()}

	var commandName string
	if flag.NArg() > 0 {
		commandName = flag.Arg(0)
	}
	selectedIndex := slices.IndexFunc(commands, func(command Command) bool {
		return command.flagSet.Name() == commandName
	})
	if selectedIndex == -1 {
		fmt.Fprint(os.Stderr, "Usage: sd", getCommandNames(commands))
		os.Exit(1)
	}

	commands[selectedIndex].flagSet.Parse(flag.Args()[1:])

	log.SetFlags(logFlags)
	var logLevel slog.Level
	var unmarshallErr = logLevel.UnmarshalText([]byte(logLevelString))
	if unmarshallErr != nil {
		panic("Invalid log level " + logLevelString + ": " + unmarshallErr.Error())
	}
	slog.SetLogLoggerLevel(logLevel)

	commands[selectedIndex].onSelected()
}

func createLogCommand(stdOut io.Writer) Command {
	flagSet := flag.NewFlagSet("log", flag.ExitOnError)

	return Command{flagSet: flagSet, onSelected: func() {
		sd.PrintGitLog(stdOut)
	}}
}

func createNewCommand() Command {
	flagSet := flag.NewFlagSet("new", flag.ExitOnError)

	var draft bool
	var featureFlag string
	var baseBranch string
	var logFlags int
	flagSet.BoolVar(&draft, "draft", true, "Whether to create the PR as draft")
	flagSet.StringVar(&featureFlag, "feature-flag", "", "Value for FEATURE_FLAG in PR description")
	flagSet.StringVar(&baseBranch, "base", ex.GetMainBranch(), "Base branch for Pull Request")
	flagSet.Usage = func() {
		fmt.Fprint(os.Stderr,
			ex.Reset+"Create a new PR with a cherry-pick of the given commit hash\n"+
				"\n"+
				"new-pr [flags] [commit hash to make PR for (default is top commit on "+ex.GetMainBranch()+")]\n"+
				"\n"+
				ex.White+"Note on Ticket Number:"+ex.Reset+"\n"+
				"\n"+
				"If you prefix the Jira ticket to the git commit summary then the `Ticket` section of the PR description will be populated with it.\n"+
				"For example:\n"+
				"\"CONV-9999 Add new feature\"\n"+
				"\n"+
				ex.White+"Note on Templates:"+ex.Reset+"\n"+
				"\n"+
				"The Pull Request Title, Body (aka Description), and Branch Name are created from golang templates. The defaults are:\n"+
				"\n"+
				"- branch-name.template - src/config/branch-name.template\n"+
				"- pr-description.template - src/config/pr-description.template\n"+
				"- pr-title.template - src/config/pr-title.template\n"+
				"\n"+
				"The possible values for the templates are:\n"+
				"\n"+
				"- **CommitBody** - Body of the commit message\n"+
				"- **CommitSummary** - Summary line of the commit message\n"+
				"- **CommitSummaryCleaned** - Summary line of the commit message without spaces or special characters\n"+
				"- **CommitSummaryWithoutTicket** - Summary line of the commit message without the prefix of the ticket number\n"+
				"- **FeatureFlag** - Value passed to feature-flag flag\n"+
				"- **TicketNumber** - Jira ticket as parsed from the commit summary\n"+
				"- **Username** -  Name as parsed from git config email\n"+
				"\n"+
				"To change a template, copy the default from [src/config/](src/config/) into `~/.stacked-diff-workflow/` and modify.\n"+
				"\n"+
				ex.White+"Flags:"+ex.Reset+"\n")
		flagSet.PrintDefaults()
	}

	return Command{flagSet: flagSet, onSelected: func() {
		if flagSet.NArg() > 1 {
			flagSet.Usage()
			os.Exit(1)
		}
		log.SetFlags(logFlags)
		branchInfo := sd.GetBranchInfo(flagSet.Arg(0))
		sd.CreateNewPr(draft, featureFlag, baseBranch, logFlags, branchInfo, log.Default())
	}}
}

func createUpdateCommand() Command {
	flagSet := flag.NewFlagSet("update", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprint(os.Stderr,
			ex.Reset+"Add one or more commits to a PR.\n"+
				"\n"+
				"update-pr <pr-commit> [fixup commit (defaults to top commit)] [other fixup commit...]\n"+
				ex.White+"Flags:"+ex.Reset+"\n")
		flagSet.PrintDefaults()
	}

	return Command{flagSet: flagSet, onSelected: func() {
		if flagSet.NArg() == 0 {
			flagSet.Usage()
			os.Exit(1)
		}

		var otherCommits []string
		if len(flagSet.Args()) > 1 {
			otherCommits = flagSet.Args()[1:]
		}
		sd.UpdatePr(flagSet.Arg(0), otherCommits, log.Default())
	}}
}

func getCommandNames(commands []Command) []string {
	var names []string
	names = slices.Grow(names, len(commands))
	for _, command := range commands {
		names = append(names, command.flagSet.Name())
	}
	slices.Sort(names)
	return names
}
