package interactive

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

/*
Do you want to add reviewers? y/n
Enter, but how
New reviewer:
Previous reviewers:

Reviewers to add when checks pass:
*Enter comma delimited names here and/or select below*
jallen,ankit,danm
previous picks, plus preview picks broken down, by most recently used and then alphabetically
jallen
ankit
hossain
dan

so that means that we cannot unselect the first issue

I could also do tab for auto-complete of names
*/
type userSelectionModel struct {
	textInput    textinput.Model
	history      []string
	historyIndex int
	confirmed    bool
	err          error
}

func (m userSelectionModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m userSelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			m.confirmed = true
			return m, tea.Quit
		case tea.KeyUp:
			appendToHistory := ""
			if m.historyIndex == -1 && m.textInput.Value() != "" && len(m.history) > 0 && m.history[len(m.history)-1] != m.textInput.Value() {
				appendToHistory = m.textInput.Value()
			}
			if m.historyIndex == -1 && len(m.history) > 0 {
				m.historyIndex = len(m.history) - 1
				m.textInput.SetValue(m.history[m.historyIndex])
			} else if m.historyIndex > 0 {
				m.historyIndex--
				m.textInput.SetValue(m.history[m.historyIndex])
			}
			if appendToHistory != "" {
				m.history = append(m.history, appendToHistory)
			}
			return m, nil
		case tea.KeyDown:
			if m.historyIndex != -1 {
				if m.historyIndex < len(m.history)-1 {
					m.historyIndex++
					m.textInput.SetValue(m.history[m.historyIndex])
				} else {
					m.historyIndex = -1
					m.textInput.SetValue("")
				}
			}
			return m, nil
		}

	case error:
		m.err = msg
		return m, tea.Quit
	}

	previousValue := m.textInput.Value()
	m.textInput, cmd = m.textInput.Update(msg)
	updatedValue := m.textInput.Value()
	if previousValue != updatedValue {
		m.historyIndex = -1
	}
	return m, cmd
}

func (m userSelectionModel) View() string {
	return fmt.Sprintf("Reviewers to add when checks pass?\n"+
		"%s\n"+
		"(up/down for history, tab to select auto-complete, enter to confirm, esc to quit)\n",
		m.textInput.View())
}

func UserSelection(appConfig util.AppConfig) string {
	input := textinput.New()
	input.Focus()
	input.Width = 100
	input.Placeholder = "None"
	input.ShowSuggestions = true
	history := []string{
		"danm200,ankit223",
		"slack-jallen",
		"ankit299",
		"danm",
		"slack-jallen",
		"ankit299",
		"danm",
	}
	input.SetSuggestions(history)
	initialModel := userSelectionModel{
		history:      history,
		historyIndex: -1,
		textInput:    input,
		confirmed:    false,
		err:          nil,
	}
	finalModel, err := NewProgram(
		initialModel,
		tea.WithInput(appConfig.Io.In),
		tea.WithOutput(appConfig.Io.Out),
	).Run()
	if err != nil {
		panic(err)
	}
	finalSelectionModel := finalModel.(userSelectionModel)
	if finalSelectionModel.err != nil {
		panic(finalSelectionModel.err)
	}
	if !finalSelectionModel.confirmed {
		appConfig.Exit(0)
	}
	return finalSelectionModel.textInput.Value()
}
