package commands

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/fatih/color"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

func ExecuteCommand(stdOut io.Writer, stdErr io.Writer, stdIn io.Reader, commandLineArgs []string, sequenceEditorPrefix string, createExit func(lstdErr io.Writer, logLeveler slog.Leveler) func(err any)) {
	// Unset any color in case a previous terminal command set colors and then was
	// terminated before it could reset the colors.
	color.Unset()

	parseArguments(stdOut, stdErr, stdIn, flag.NewFlagSet("sd", flag.ContinueOnError), commandLineArgs, sequenceEditorPrefix, createExit)
}

func CreateDefaultExit(stdErr io.Writer, logLeveler slog.Leveler) func(err any) {
	return func(err any) {
		// Show panic stack trace during debug log level.
		if logLeveler.Level() <= slog.LevelDebug {
			if err == nil {
				panic("Cancelled")
			} else {
				panic(err)
			}
		} else if err != nil {
			fmt.Fprintln(stdErr, fmt.Sprint("error: ", err))
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}
}

func parseArguments(stdOut io.Writer, stdErr io.Writer, stdIn io.Reader, commandLine *flag.FlagSet, commandLineArgs []string, sequenceEditorPrefix string, createExit func(stdErr io.Writer, logLeveler slog.Leveler) func(err any)) {
	if commandLine.ErrorHandling() != flag.ContinueOnError {
		// Use ContinueOnError so that a description of the command can be included before usage
		// for help.
		panic("ErrorHandling must be ContinueOnError, not " + fmt.Sprint(commandLine.ErrorHandling()))
	}
	// clear FlagSet.Usage and discard any output so that it is not display automatically on an invalid parameter.
	commandLine.Usage = func() {}
	commandLine.SetOutput(io.Discard)
	// Parse top level flags.
	logLevelString := commandLine.String("log-level", "",
		"Possible log levels:\n"+
			"   debug\n"+
			"   info\n"+
			"   warn\n"+
			"   error\n"+
			"Default is info, except on commands that are for output purposes,\n"+
			"(namely branch-name and log), which have a default of error.")
	parseErr := commandLine.Parse(commandLineArgs)
	var logLevelVar *slog.LevelVar
	if parseErr == nil {
		// allow for setting of log level to DEBUG so that the very first execute statements can be logged.
		// logLevel will be potentially set again once we know what command is executed.
		var logLevel slog.Level
		logLevel, parseErr = stringToLogLevel(*logLevelString)
		if parseErr == nil {
			logLevelVar = setSlogLogger(stdOut, logLevel)
		}
	}
	// parseErr is dealt with below via commandError and commandHelp.

	commands := []Command{
		createAddReviewersCommand(),
		createBranchNameCommand(),
		createCodeOwnersCommand(),
		createDropAlreadyMergedCommand(),
		createLogCommand(),
		createMarkAsFixupCommand(),
		createNewCommand(),
		createPrsCommand(),
		createRebaseMainCommand(),
		createReplaceCommitCommand(),
		createReplaceConflictsCommand(),
		createUpdateCommand(),
		createCheckoutCommand(),
		createWaitForMergeCommand(),
	}

	commandLineDescription := "Stacked Diff Workflow"
	commandLineUsage := "sd [top-level-flags] <command> [<args>]\n" +
		"\n" +
		"Possible commands are:\n" +
		"\n" +
		"   " + strings.Join(getCommandSummaries(commands), "\n   ") + "\n" +
		"\n" +
		"To learn more about a command use: sd <command> --help"

	if parseErr != nil {
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

	if *logLevelString == "" {
		logLevelVar.Set(commands[selectedIndex].DefaultLogLevel)
	}
	exit := createExit(stdErr, logLevelVar)
	defer func() {
		r := recover()
		if r != nil {
			exit(r)
		}
	}()
	// Note: call GetMainBranchOrDie early as it has useful error messages.
	slog.Debug(fmt.Sprint("Using main branch " + util.GetMainBranchOrDie()))

	commands[selectedIndex].OnSelected(commands[selectedIndex], stdOut, stdErr, stdIn, sequenceEditorPrefix, exit)
}

func getCommandSummaries(commands []Command) []string {
	publicCommands := util.FilterSlice(commands, func(command Command) bool {
		return !command.Hidden
	})
	maxCommandLen := 0
	for _, command := range publicCommands {
		if len(command.FlagSet.Name()) > maxCommandLen {
			maxCommandLen = len(command.FlagSet.Name())
		}
	}
	summaries := make([]string, 0, len(commands))
	for _, command := range publicCommands {
		summary := command.FlagSet.Name() + "   " + strings.Repeat(" ", maxCommandLen-len(command.FlagSet.Name())) + command.Summary
		summaries = append(summaries, summary)
	}
	slices.Sort(summaries)
	return summaries
}

func setSlogLogger(stdOut io.Writer, logLevel slog.Level) *slog.LevelVar {
	var levelVar slog.LevelVar
	levelVar.Set(logLevel)
	opts := util.PrettyHandlerOptions{
		SlogOpts: slog.HandlerOptions{
			Level: &levelVar,
		},
	}
	handler := util.NewPrettyHandler(stdOut, opts)
	slog.SetDefault(slog.New(handler))
	return &levelVar
}

func stringToLogLevel(logLevelString string) (slog.Level, error) {
	if logLevelString == "" {
		return slog.LevelInfo, nil
	}
	var logLevel slog.Level
	var unmarshallErr = logLevel.UnmarshalText([]byte(logLevelString))
	if unmarshallErr != nil {
		return 0, unmarshallErr
	}
	return logLevel, nil
}
