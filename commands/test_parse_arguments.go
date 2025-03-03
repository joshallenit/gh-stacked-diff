package commands

import (
	"flag"
	"io"
	"log/slog"

	"github.com/joshallenit/gh-stacked-diff/v2/testutil"
)

// Program name
const programName string = "gh-stacked-diff"

// Calls [parseArguments] for unit tests.
func testParseArguments(commandLineArgs ...string) string {
	createPanicOnExit := func(stdErr io.Writer, logLevelVar *slog.LevelVar) func(err any) {
		return func(err any) {
			panic(err)
		}
	}
	out := testutil.NewWriteRecorder()
	// Executable must be on PATH for tests to pass so that sequenceEditorPrefix will execute.
	// PATH is set in ../Makefile
	sequenceEditorPrefix := programName + " --log-level=INFO "
	parseArguments(out, out, flag.NewFlagSet("sd", flag.ContinueOnError), commandLineArgs, sequenceEditorPrefix, createPanicOnExit)
	return out.String()
}
