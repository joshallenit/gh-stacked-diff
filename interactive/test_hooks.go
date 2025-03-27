package interactive

import (
	"fmt"
	"testing"

	"slices"

	tea "github.com/charmbracelet/bubbletea"
)

var programListeners = make([]*newProgramListener, 0)

type newProgramListener struct {
	messages             []tea.Msg
	numSent              int
	targetProgram        int
	currentProgramNumber int
}

func (l *newProgramListener) onNewProgram(program *tea.Program) {
	l.currentProgramNumber++
	if l.targetProgram == l.currentProgramNumber {
		go func() {
			for _, msg := range l.messages {
				program.Send(msg)
				l.numSent++
			}
		}()
	}
}

// Call instead of [tea.NewProgram] to support testing hook [SendToProgram].
func NewProgram(model tea.Model, opts ...tea.ProgramOption) *tea.Program {
	program := tea.NewProgram(model, opts...)
	for _, listener := range programListeners {
		listener.onNewProgram(program)
	}
	return program
}

// Sends messages to the program. Each time [NewProgram] is called after [SendToProgram]
// programIndex is incremented.
func SendToProgram(t *testing.T, programIndex int, messages ...tea.Msg) {
	programListener := &newProgramListener{messages: messages, currentProgramNumber: -1, targetProgram: programIndex}
	programListeners = append(programListeners, programListener)
	t.Cleanup(func() {
		programListeners = slices.DeleteFunc(programListeners, func(next *newProgramListener) bool {
			return next == programListener
		})
		if programListener.numSent != len(programListener.messages) {
			panic("Did not use the all of the desired input: " + fmt.Sprint(*programListener))
		}
	})
}

// Returns whether [SendToProgram] has been setup.
func HasProgramMessagesSet() bool {
	return len(programListeners) > 0
}

// Convienience method for creating a message for when user typed a key.
func NewMessageRune(r rune) tea.KeyMsg {
	return tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{r}})
}

// Convienience method for creating a message for when user hits a non-rune key like enter or up/down.
func NewMessageKey(keyType tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg(tea.Key{Type: keyType})
}
