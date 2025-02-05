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
	ParseArguments(os.Stdout, flag.NewFlagSet("sd", flag.ContinueOnError), os.Args[1:])
}

func ParseArguments(stdOut io.Writer, commandLine *flag.FlagSet, commandLineArgs []string) {
	if commandLine.ErrorHandling() != flag.ContinueOnError {
		// Use ContinueOnError so that a description of the command can be included before usage
		// for help.
		panic("ErrorHandling must be ContinueOnError, not " + fmt.Sprint(commandLine.ErrorHandling()))
	}
	commands := []Command{
		CreateAddReviewersCommand(),
		CreateBranchNameCommand(stdOut),
		CreateCodeOwnersCommand(stdOut),
		CreateLogCommand(stdOut),
		CreateNewCommand(),
		CreatePrsCommand(stdOut),
		CreateRebaseMainCommand(),
		CreateReplaceCommitCommand(),
		CreateReplaceConflictsCommand(stdOut),
		CreateUpdateCommand(),
		CreateCheckoutCommand(),
		CreateWaitForMergeCommand(),
	}

	commandLineDescription := "Stacked Diff Workflow"
	commandLineUsage := "sd [top-level-flags] <command> [<args>]\n" +
		"\n" +
		"Possible commands are:\n" +
		"\n" +
		"   " + strings.Join(getCommandSummaries(commands), "\n   ") + "\n" +
		"\n" +
		"To learn more about a command use: sd <command> --help"
	// clear FlagSet.Usage and discard any output so that it is not display automatically on an invalid parameter.
	commandLine.Usage = func() {}
	commandLine.SetOutput(io.Discard)
	// Parse flags common for every command.
	var logLevelString string

	commandLine.StringVar(&logLevelString, "log-level", "",
		"Possible log levels:\n"+
			"   debug\n"+
			"   info\n"+
			"   warn\n"+
			"   error\n"+
			"Default is info, except on commands that are for output purposes,\n"+
			"(namely branch-name and log), which have a default of error.")

	if parseErr := commandLine.Parse(commandLineArgs); parseErr != nil {
		if parseErr == flag.ErrHelp {
			commandHelp(commandLine, commandLineDescription, commandLineUsage, false)
		} else {
			commandError(commandLine, parseErr.Error(), commandLineUsage)
		}
	}

	if commandLine.NArg() == 0 {
		commandHelp(commandLine, commandLineDescription, commandLineUsage, true)
	}
	selectedIndex := slices.IndexFunc(commands, func(command Command) bool {
		return command.FlagSet.Name() == commandLine.Arg(0)
	})
	if selectedIndex == -1 {
		commandError(commandLine, "unknown command "+commandLine.Arg(0), commandLineUsage)
	}

	if commands[selectedIndex].FlagSet.ErrorHandling() != flag.ContinueOnError {
		panic("ErrorHandling must be ContinueOnError, not " + fmt.Sprint(commands[selectedIndex].FlagSet.ErrorHandling()))
	}
	commands[selectedIndex].FlagSet.Usage = func() {}
	commands[selectedIndex].FlagSet.SetOutput(io.Discard)
	if parseErr := commands[selectedIndex].FlagSet.Parse(commandLine.Args()[1:]); parseErr != nil {
		if parseErr == flag.ErrHelp {
			commandHelp(commands[selectedIndex].FlagSet, commands[selectedIndex].Description, commands[selectedIndex].Usage, false)
		} else {
			commandError(commands[selectedIndex].FlagSet, parseErr.Error(), commands[selectedIndex].Usage)
		}
	}

	setSlogLogger(stdOut, logLevelString, commands[selectedIndex])

	commands[selectedIndex].OnSelected(commands[selectedIndex])
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
		summary := command.FlagSet.Name() + "   " + strings.Repeat(" ", maxCommandLen-len(command.FlagSet.Name())) + command.Summary
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
