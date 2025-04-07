package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"slices"

	"strings"

	"github.com/joshallenit/gh-stacked-diff/v2/testutil"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

// Program name
const programName string = "gh-stacked-diff"

// Calls [parseArguments] for unit tests.
func testParseArguments(commandLineArgs ...string) string {
	out := testutil.NewWriteRecorder()
	testParseArgumentsWithOut(out, commandLineArgs...)
	return out.String()
}

func testParseArgumentsWithOut(out *testutil.WriteRecorder, commandLineArgs ...string) {
	slog.Debug(fmt.Sprint("***Testing parse arguments*** ", strings.Join(commandLineArgs, " ")))
	panicOnExit := func(code int) {
		panic("Panicking instead of exiting with code " + fmt.Sprint(code))
	}
	// Executable must be on PATH for tests to pass so that sequenceEditorPrefix will execute.
	// PATH is set in ../Makefile
	appExecutable := programName

	// Use debug log level if currently set to debug
	if prettyHandler, ok := slog.Default().Handler().(*util.PrettyHandler); ok {
		if prettyHandler.Opts.Level == slog.LevelDebug {
			appExecutable += " --log-level=debug"
			if !slices.ContainsFunc(commandLineArgs, func(next string) bool {
				return strings.HasPrefix(next, "--log-level")
			}) {
				commandLineArgs = slices.Insert(commandLineArgs, 0, "--log-level=debug")
			}
		}
	}

	// Set stdin in unit tests to avoid error with bubbletea:
	// "error creating cancelreader: failed to prepare console input: get console mode: The handle is invalid."
	// To fake user input use interactive.SendToProgram.
	stdin := strings.NewReader("")

	appConfig := util.AppConfig{
		Io:            util.StdIo{Out: out, Err: out, In: stdin},
		AppExecutable: appExecutable,
		Exit:          panicOnExit,
	}
	parseArguments(
		appConfig,
		flag.NewFlagSet("sd", flag.ContinueOnError),
		commandLineArgs,
	)
	slog.Debug(fmt.Sprint("***Done running arguments*** ", strings.Join(commandLineArgs, " ")))
}
