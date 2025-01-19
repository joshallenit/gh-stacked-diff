package execute

import (
	"log"
	"strings"
)

type fakeResponse struct {
	out         string
	err         error
	programName string
	args        []string
}

type TestExecutor struct {
	TestLogger    *log.Logger
	fakeResponses []fakeResponse
}

// Ensure that [TestExecutor] implements [Executor].
var _ Executor = TestExecutor{}

func (t TestExecutor) Execute(options ExecuteOptions, programName string, args ...string) (string, error) {
	for _, response := range t.fakeResponses {
		if response.programName == programName {
			matched := true
			if len(response.args) <= len(args) {
				for i, arg := range response.args {
					if arg != args[i] {
						matched = false
						break
					}
				}
			} else {
				matched = false
			}
			if matched {
				return response.out, response.err
			}
		}
	}
	log.Print("Executing ", programName, " ", strings.Join(args, " "))
	out, err := (&DefaultExecutor{}).Execute(options, programName, args...)
	if out != "" {
		log.Print("\n", out)
	}
	if err != nil {
		log.Print("\nError", err)
	}
	log.Println("")
	return out, err
}

func (t TestExecutor) Logger() *log.Logger {
	return t.TestLogger
}

func (t *TestExecutor) SetResponse(out string, err error, programName string, args ...string) {
	t.fakeResponses = append(t.fakeResponses, fakeResponse{out: out, err: err, programName: programName, args: args})
}
