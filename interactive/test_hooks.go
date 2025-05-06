package interactive

import (
	"fmt"
	"slices"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

var sendMessageProgramListener func(program *tea.Program)
var fakeMessages = map[int][]tea.Msg{}

// Call instead of [tea.NewProgram] to support testing hook [SendToProgram].
func NewProgram(model tea.Model, stdIo util.StdIo) *tea.Program {
	program := tea.NewProgram(
		model,
		tea.WithInput(stdIo.In),
		tea.WithOutput(stdIo.Out),
	)
	if sendMessageProgramListener != nil {
		go sendMessageProgramListener(program)
	}
	return program
}

// Sends messages to the program. Each time [NewProgram] is called after [SendToProgram]
// programIndex is incremented.
// This is used instead of using stdin to avoid having to (somehow?) fake keyboard scan codes.
func SendToProgram(programIndex int, messages ...tea.Msg) {
	if sendMessageProgramListener == nil {
		panic("RequireInput must be called by test init")
	}
	programMessages := fakeMessages[programIndex]
	if programMessages == nil {
		programMessages = []tea.Msg{}
	}
	programMessages = slices.AppendSeq(programMessages, slices.Values(messages))
	fakeMessages[programIndex] = programMessages
}

func RequireInput(t *testing.T) {
	currentProgramIndex := 0
	if sendMessageProgramListener != nil {
		panic("RequireInput already called for this test")
	}
	sendMessageProgramListener = func(program *tea.Program) {
		programMessages := fakeMessages[currentProgramIndex]
		if len(programMessages) == 0 {
			panic(fmt.Sprint(
				"no input setup for interactive ui program number ",
				currentProgramIndex, ", use interactive.SendToProgram"))
		}
		currentProgramIndex++
		for _, msg := range programMessages {
			program.Send(msg)
		}
	}
	t.Cleanup(func() {
		sendMessageProgramListener = nil
		fakeMessages = map[int][]tea.Msg{}
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
