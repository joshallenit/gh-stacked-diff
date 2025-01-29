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

	// Parse flags common for every command.
	var logLevelString string

	commandLine.StringVar(&logLevelString, "log-level", "", "Log level: debug, info, warn, or error. Default is info except for branch-name which is error.")

	commandLine.Parse(commandLineArgs)

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

	var commandName string
	if commandLine.NArg() > 0 {
		commandName = commandLine.Arg(0)
	}
	selectedIndex := slices.IndexFunc(commands, func(command Command) bool {
		return command.FlagSet.Name() == commandName
	})
	if selectedIndex == -1 {
		if commandName != "" {
			fmt.Fprintln(os.Stderr, "Unknown command", commandName)
		}
		fmt.Fprintln(os.Stderr, "Usage: sd [", strings.Join(getCommandNames(commands), " | "), "]")
		os.Exit(1)
	}

	commands[selectedIndex].FlagSet.Parse(commandLine.Args()[1:])

	setSlogLogger(stdOut, logLevelString, commands[selectedIndex])

	commands[selectedIndex].OnSelected()
}

func getCommandNames(commands []Command) []string {
	var names []string
	names = slices.Grow(names, len(commands))
	for _, command := range commands {
		names = append(names, command.FlagSet.Name())
	}
	slices.Sort(names)
	return names
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
