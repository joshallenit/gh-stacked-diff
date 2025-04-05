package interactive

import (
	"testing"

	"slices"

	tea "github.com/charmbracelet/bubbletea"
)

var programListeners = make([]*func(*tea.Program), 0)
var hasFakeProgramMessages bool = false

// Call instead of [tea.NewProgram] to support testing hook [SendToProgram].
func NewProgram(model tea.Model, opts ...tea.ProgramOption) *tea.Program {
	program := tea.NewProgram(model, opts...)
	for _, listener := range programListeners {
		(*listener)(program) //.onNewProgram(program)
	}
	return program
}

// Sends messages to the program. Each time [NewProgram] is called after [SendToProgram]
// programIndex is incremented.
// This is used instead of using stdin to avoid having to (somehow?) fake keyboard scan codes.
func SendToProgram(t *testing.T, programIndex int, messages ...tea.Msg) {
	hasFakeProgramMessages = true
	currentProgramNumber := 0
	listener := func(program *tea.Program) {
		if currentProgramNumber == programIndex {
			go func() {
				for _, msg := range messages {
					program.Send(msg)
				}
			}()
		}
		currentProgramNumber++
	}
	AddNewProgramListener(t, listener)
	t.Cleanup(func() {
		hasFakeProgramMessages = false
	})
}

func AddNewProgramListener(t *testing.T, listener func(*tea.Program)) {
	programListeners = append(programListeners, &listener)
	t.Cleanup(func() {
		programListeners = slices.DeleteFunc(programListeners, func(next *func(*tea.Program)) bool {
			return next == &listener
		})
	})
}

// Convienience method for creating a message for when user typed a key.
func NewMessageRune(r rune) tea.KeyMsg {
	return tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{r}})
}

// Convienience method for creating a message for when user hits a non-rune key like enter or up/down.
func NewMessageKey(keyType tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg(tea.Key{Type: keyType})
}

func HasFakeProgramMessages() bool {
	return hasFakeProgramMessages
}
