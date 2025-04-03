package interactive

import (
	"io"

	"github.com/charmbracelet/x/term"
)

// that means that confirm will also have to return an error? that's not right
/*
So the question is do I keep appConfig.Exit and whether to move to using panic for all
errors with custom error types (preferred)
*/
var ErrorNotATerminal string = "not a terminal"

func IsOutputTerminal(out io.Writer) bool {
	f, isFile := out.(term.File)
	if !isFile {
		return false
	}
	return term.IsTerminal(f.Fd())
}
