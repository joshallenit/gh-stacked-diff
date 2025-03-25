package interactive

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var programListeners = make([]func(program *tea.Program), 0)

// Call instead of [tea.NewProgram] to support testing hook [SendToProgram].
func NewProgram(model tea.Model, opts ...tea.ProgramOption) *tea.Program {
	program := tea.NewProgram(model, opts...)
	for _, listener := range programListeners {
		listener(program)
	}
	return program
}

// Sends messages to the program. Each time [NewProgram] is called after [SendToProgram]
// programIndex is incremented.
func SendToProgram(t *testing.T, programIndex int, messages ...tea.Msg) {
	addNewProgramListener(t, func(currentProgramIndex int, program *tea.Program) {
		if currentProgramIndex == programIndex {
			go func() {
				for _, msg := range messages {
					program.Send(msg)
				}
			}()
		}
	})
}

// Returns whether [SendToProgram] has been setup.
func HasProgramMessagesSet() bool {
	return len(programListeners) > 0
}

func addNewProgramListener(t *testing.T, onNewProgram func(programIndex int, program *tea.Program)) {
	programIndex := 0
	programListener := func(program *tea.Program) {
		onNewProgram(programIndex, program)
		programIndex++
	}
	programListeners = append(programListeners, programListener)
	t.Cleanup(func() {
		// clear all. Removing individually would require using an ID as you cannot compare functions in golang.
		programListeners = make([]func(program *tea.Program), 0)
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
