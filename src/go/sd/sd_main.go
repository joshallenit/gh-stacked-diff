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
	"strings"
	"time"
)

type Command struct {
	flagSet    *flag.FlagSet
	onSelected func()
	// Default if not set is 0 which is Info.
	defaultLogLevel slog.Level
}

/*
sd - stacked diff - command line interface.
*/
func main() {
	ParseArguments(os.Stdout, flag.CommandLine, os.Args[1:])
}

func ParseArguments(stdOut io.Writer, commandLine *flag.FlagSet, commandLineArgs []string) {

	// Parse flags common for every command.
	var logFlags int
	var logLevelString string

	commandLine.IntVar(&logFlags, "log-flags", 0, "Log flags, see https://pkg.go.dev/log#pkg-constants")
	commandLine.StringVar(&logLevelString, "log-level", "", "Log level: debug, info, warn, or error. Default is info except for branch-name which is error.")

	commandLine.Parse(commandLineArgs)

	commands := []Command{
		createLogCommand(stdOut),
		createNewCommand(),
		createUpdateCommand(),
		createAddReviewersCommand(),
		createBranchNameCommand(stdOut),
	}

	var commandName string
	if commandLine.NArg() > 0 {
		commandName = commandLine.Arg(0)
	}
	selectedIndex := slices.IndexFunc(commands, func(command Command) bool {
		return command.flagSet.Name() == commandName
	})
	if selectedIndex == -1 {
		if commandName != "" {
			fmt.Fprintln(os.Stderr, "Unknown command", commandName)
		}
		fmt.Fprintln(os.Stderr, "Usage: sd [", strings.Join(getCommandNames(commands), " | "), "]")
		os.Exit(1)
	}

	commands[selectedIndex].flagSet.Parse(commandLine.Args()[1:])

	log.SetFlags(logFlags)
	setLogLevel(logLevelString, commands[selectedIndex])

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
	flagSet.BoolVar(&draft, "draft", true, "Whether to create the PR as draft")
	flagSet.StringVar(&featureFlag, "feature-flag", "", "Value for FEATURE_FLAG in PR description")
	flagSet.StringVar(&baseBranch, "base", ex.GetMainBranch(), "Base branch for Pull Request")

	var reviewers string
	flagSet.StringVar(&reviewers, "reviewers", "", "Comma-separated list of Github usernames to add as reviewers once checks have passed.")
	var silent bool
	var minChecks int
	flagSet.BoolVar(&silent, "silent", false, "Whether to use voice output (false) or be silent (true) to notify that reviewers have been added.")
	flagSet.IntVar(&minChecks, "min-checks", 4,
		"Minimum number of checks to wait for before verifying that checks have passed. "+
			"It takes some time for checks to be added to a PR by Github, "+
			"and if you add-reviewers too soon it will think that they have all passed.")

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
		branchInfo := sd.GetBranchInfo(flagSet.Arg(0))
		sd.CreateNewPr(draft, featureFlag, baseBranch, branchInfo, log.Default())
		if reviewers != "" {
			sd.AddReviewersToPr([]string{branchInfo.CommitHash}, true, silent, minChecks, reviewers, 30*time.Second)
		}
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

func createAddReviewersCommand() Command {
	flagSet := flag.NewFlagSet("add-reviewers", flag.ExitOnError)
	var reviewers string

	flagSet.StringVar(&reviewers, "reviewers", "", "Comma-separated list of Github usernames to add as reviewers. "+
		"Falls back to "+ex.White+"PR_REVIEWERS"+ex.Reset+" environment variable. "+
		"You can specify more than one reviewer using a comma-delimited string.")
	var whenChecksPass bool
	var pollFrequency time.Duration
	var defaultPollFrequency time.Duration = 30 * time.Second
	var silent bool
	var minChecks int
	flagSet.BoolVar(&whenChecksPass, "when-checks-pass", true, "Poll until all checks pass before adding reviewers")
	flagSet.DurationVar(&pollFrequency, "poll-frequency", defaultPollFrequency,
		"Frequency which to poll checks. For valid formats see https://pkg.go.dev/time#ParseDuration")
	flagSet.BoolVar(&silent, "silent", false, "Whether to use voice output")
	flagSet.IntVar(&minChecks, "min-checks", 4,
		"Minimum number of checks to wait for before verifying that checks have passed. "+
			"It takes some time for checks to be added to a PR by Github, "+
			"and if you add-reviewers too soon it will think that they have all passed.")
	flagSet.Usage = func() {
		fmt.Fprint(os.Stderr,
			ex.Reset+"Mark a Draft PR as \"Ready for Review\" and automatically add reviewers.\n"+
				"\n"+
				"add-reviewers [flags] <commit hash or pull request number>\n"+
				"\n"+
				ex.White+"Flags:"+ex.Reset+"\n")
		flagSet.PrintDefaults()
	}
	return Command{flagSet: flagSet, onSelected: func() {
		if flagSet.NArg() == 0 {
			flagSet.Usage()
			os.Exit(1)
		}
		sd.AddReviewersToPr(flagSet.Args(), whenChecksPass, silent, minChecks, reviewers, pollFrequency)
	}}
}

func createBranchNameCommand(stdOut io.Writer) Command {
	flagSet := flag.NewFlagSet("branch-name", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Outputs the branch name for a given commit hash or pull request number. Useful for custom scripting.")
		fmt.Fprintln(os.Stderr, "sd branch-name <commit hash or pull request number>")
		flagSet.PrintDefaults()
	}

	return Command{flagSet: flagSet, defaultLogLevel: slog.LevelError, onSelected: func() {
		if flagSet.NArg() != 1 {
			flagSet.Usage()
			os.Exit(1)
		}
		branchName := sd.GetBranchInfo(flagSet.Arg(0)).BranchName
		fmt.Fprint(stdOut, branchName)
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

func setLogLevel(logLevelString string, selectedCommand Command) {
	var logLevel slog.Level
	if logLevelString == "" {
		logLevel = selectedCommand.defaultLogLevel
	} else {
		var unmarshallErr = logLevel.UnmarshalText([]byte(logLevelString))
		if unmarshallErr != nil {
			panic("Invalid log level " + logLevelString + ": " + unmarshallErr.Error())
		}
	}
	slog.SetLogLoggerLevel(logLevel)
}
