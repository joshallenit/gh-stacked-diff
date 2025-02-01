package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"

	ex "stackeddiff/execute"
)

/*
sd - stacked diff - command line interface.
*/
func main() {
	ParseArguments(os.Stdout, flag.CommandLine, os.Args[1:])
}

func ParseArguments(stdOut io.Writer, commandLine *flag.FlagSet, commandLineArgs []string) {
	commands := []Command{
		CreateAddReviewersCommand(),
		CreateBranchNameCommand(stdOut),
		CreateCodeOwnersCommand(stdOut),
		CreateLogCommand(stdOut),
		CreateNewCommand(),
		CreateRebaseMainCommand(),
		CreateReplaceCommitCommand(),
		CreateUpdateCommand(),
		CreateCheckoutCommand(),
		CreateWaitForMergeCommand(),
	}

	commandLine.Usage = func() {
		fmt.Fprintln(commandLine.Output(), "Stacked Diff Workflow")
		fmt.Fprintln(commandLine.Output(), "")
		fmt.Fprintln(commandLine.Output(), "usage: sd [flags] <command> [<args>]")
		fmt.Fprintln(commandLine.Output(), "")
		fmt.Fprintln(commandLine.Output(), "Possible commands are:")
		fmt.Fprintln(commandLine.Output(), "")
		fmt.Fprintln(commandLine.Output(), "   "+strings.Join(getCommandSummaries(commands), "\n   "))
		fmt.Fprintln(commandLine.Output(), "")
		fmt.Fprintln(commandLine.Output(), "To learn more about a command use: sd <command> --help")
		fmt.Fprintln(commandLine.Output(), "")
		fmt.Fprintln(commandLine.Output(), "Top level flags:")
		fmt.Fprintln(commandLine.Output(), "")
		commandLine.PrintDefaults()
	}
	// Parse flags common for every command.
	var logLevelString string

	commandLine.StringVar(&logLevelString, "log-level", "", "Log level: debug, info, warn, or error. Default is info except on command that are for output purposes like branch-name and log which are error.")

	commandLine.Parse(commandLineArgs)

	if commandLine.NArg() == 0 {
		commandLine.Usage()
		os.Exit(1)
	}
	selectedIndex := slices.IndexFunc(commands, func(command Command) bool {
		return command.FlagSet.Name() == commandLine.Arg(0)
	})
	if selectedIndex == -1 {
		fmt.Fprintln(commandLine.Output(), "error: unknown command", commandLine.Arg(0))
		commandLine.Usage()
		os.Exit(1)
	}

	commands[selectedIndex].FlagSet.Parse(commandLine.Args()[1:])

	setSlogLogger(stdOut, logLevelString, commands[selectedIndex])

	commands[selectedIndex].OnSelected()
}

func getCommandSummaries(commands []Command) []string {
	maxCommandLen := 0
	for _, command := range commands {
		if len(command.FlagSet.Name()) > maxCommandLen {
			maxCommandLen = len(command.FlagSet.Name())
		}
	}
	summaries := make([]string, 0, len(commands))
	for _, command := range commands {
		summary := command.FlagSet.Name() + "   " + strings.Repeat(" ", maxCommandLen-len(command.FlagSet.Name())) + command.UsageSummary
		summaries = append(summaries, summary)
	}
	slices.Sort(summaries)
	return summaries
}

func setSlogLogger(stdOut io.Writer, logLevelString string, selectedCommand Command) {
	var logLevel slog.Level
	if logLevelString == "" {
		logLevel = selectedCommand.DefaultLogLevel
	} else {
		var unmarshallErr = logLevel.UnmarshalText([]byte(logLevelString))
		if unmarshallErr != nil {
			panic("Invalid log level " + logLevelString + ": " + unmarshallErr.Error())
		}
	}
	opts := ex.PrettyHandlerOptions{
		SlogOpts: slog.HandlerOptions{
			Level: logLevel,
		},
	}
	handler := ex.NewPrettyHandler(stdOut, opts)
	slog.SetDefault(slog.New(handler))
}
