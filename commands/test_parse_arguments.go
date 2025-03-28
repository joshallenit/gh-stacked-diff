package commands

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"time"

	"strings"

	"github.com/joshallenit/gh-stacked-diff/v2/interactive"
	"github.com/joshallenit/gh-stacked-diff/v2/testutil"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

// Program name
const programName string = "gh-stacked-diff"

// io.Reader where Read just sleeps if input is setup to come from [interactive.SendToProgram]
// or panics (to indicate that the test should not use stdin)
type unsupportedReader struct{}

func (r unsupportedReader) Read([]byte) (int, error) {
	if interactive.HasProgramMessagesSet() {
		time.Sleep(10 * time.Minute)
		panic("timeout, test should have cancelled read")
	} else {
		panic("Use of stdin not expected for this test")
	}
}

var _ io.Reader = unsupportedReader{}

// Calls [parseArguments] for unit tests.
func testParseArguments(commandLineArgs ...string) string {
	slog.Debug(fmt.Sprint("***Testing parse arguments***", strings.Join(commandLineArgs, " ")))
	createPanicOnExit := func(stdErr io.Writer, logLeveler slog.Leveler) func(err any) {
		return func(err any) {
			if err == nil {
				panic("User cancelled")
			}
			panic(err)
		}
	}
	out := testutil.NewWriteRecorder()
	// Executable must be on PATH for tests to pass so that sequenceEditorPrefix will execute.
	// PATH is set in ../Makefile
	appExecutable := programName + " --log-level=INFO "

	// Set stdin in unit tests to avoid error with bubbletea:
	// "error creating cancelreader: failed to prepare console input: get console mode: The handle is invalid."
	stdin := unsupportedReader{}

	parseArguments(
		util.StdIo{Out: out, Err: out, In: stdin},
		flag.NewFlagSet("sd", flag.ContinueOnError),
		commandLineArgs,
		appExecutable,
		createPanicOnExit,
	)
	slog.Debug(fmt.Sprint("***Done running arguments***", strings.Join(commandLineArgs, " ")))
	return out.String()
}
