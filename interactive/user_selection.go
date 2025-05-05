package interactive

import (
	"regexp"
	"slices"
	"strings"

	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

const USER_HISTORY_FILE = "user-selection-history"

type userSelectionModel struct {
	textInput     textinput.Model
	history       []string
	suggestions   []string
	breakingChars []rune
	historyIndex  int
	confirmed     bool
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
			m.onKeyUp()
			return m, nil
		case tea.KeyDown:
			m.onKeyDown()
			return m, nil
		}
	case setSuggestionsMsg:
		m.suggestions = msg.suggestions
		m.setSuggestions()
		return m, nil
	}

	m.setSuggestions()

	previousValue := m.textInput.Value()
	m.textInput, cmd = m.textInput.Update(msg)
	updatedValue := m.textInput.Value()
	if previousValue != updatedValue {
		m.historyIndex = -1
	}
	return m, cmd
}

func (m userSelectionModel) View() string {
	if m.confirmed {
		return ""
	}
	return promptStyle.Render("Reviewers to add when checks pass?") + "\n" +
		m.textInput.View() + "\n" +
		"\n" +
		"Controls:\n" +
		"   up/down   history\n" +
		"   tab       select auto-complete\n" +
		"   enter     confirm\n" +
		"   esc       quit\n" + fmt.Sprint(m.textInput.AvailableSuggestions())
}

func (m *userSelectionModel) setSuggestions() {
	lastBreakingChar := -1
	valueRunes := []rune(m.textInput.Value())
	for i := len(valueRunes) - 1; i >= 0; i-- {
		if slices.Contains(m.breakingChars, valueRunes[i]) {
			lastBreakingChar = i
			break
		}
	}
	if lastBreakingChar != -1 {
		selectedFields := strings.FieldsFunc(m.textInput.Value(), func(next rune) bool {
			return slices.Contains(m.breakingChars, next)
		})
		nonSelectedSuggestions := util.FilterSlice(m.suggestions, func(next string) bool {
			return !slices.Contains(selectedFields, next)
		})
		m.textInput.SetSuggestions(util.MapSlice(nonSelectedSuggestions, func(next string) string {
			return string(valueRunes[0:lastBreakingChar+1]) + next
		}))
	} else {
		m.textInput.SetSuggestions(m.suggestions)
	}
}

func (m *userSelectionModel) onKeyUp() {
	appendToHistory := ""
	if m.historyIndex == -1 && m.textInput.Value() != "" && len(m.history) > 0 && m.history[len(m.history)-1] != m.textInput.Value() {
		appendToHistory = m.textInput.Value()
	}
	if m.historyIndex == -1 && len(m.history) > 0 {
		m.historyIndex = len(m.history) - 1
		m.textInput.SetValue(m.history[m.historyIndex])
		m.textInput.SetCursor(len(m.textInput.Value()))
	} else if m.historyIndex > 0 {
		m.historyIndex--
		m.textInput.SetValue(m.history[m.historyIndex])
		m.textInput.SetCursor(len(m.textInput.Value()))
	}
	if appendToHistory != "" {
		m.history = append(m.history, appendToHistory)
	}
}

func (m *userSelectionModel) onKeyDown() {
	if m.historyIndex != -1 {
		if m.historyIndex < len(m.history)-1 {
			m.historyIndex++
			m.textInput.SetValue(m.history[m.historyIndex])
			m.textInput.SetCursor(len(m.textInput.Value()))
		} else {
			m.historyIndex = -1
			m.textInput.SetValue("")
			m.textInput.SetCursor(len(m.textInput.Value()))
		}
	}
}

type setSuggestionsMsg struct {
	suggestions []string
}

var _ tea.Msg = setSuggestionsMsg{}

func UserSelection(appConfig util.AppConfig) string {
	input := textinput.New()
	input.Focus()
	input.Width = 100
	input.Placeholder = "None"
	input.ShowSuggestions = true
	history := util.ReadHistory(appConfig, USER_HISTORY_FILE)
	suggestions := allUsersFromHistory(history)
	input.SetSuggestions(suggestions)
	initialModel := userSelectionModel{
		history:       history,
		historyIndex:  -1,
		textInput:     input,
		confirmed:     false,
		suggestions:   suggestions,
		breakingChars: []rune{',', ' '},
	}
	program := NewProgram(
		initialModel,
		tea.WithInput(appConfig.Io.In),
		tea.WithOutput(appConfig.Io.Out),
	)
	go updateSuggestions(program)
	finalModel, err := program.Run()
	if err != nil {
		panic(err)
	}
	finalSelectionModel := finalModel.(userSelectionModel)
	if !finalSelectionModel.confirmed {
		appConfig.Exit(0)
	}
	selected := finalSelectionModel.textInput.Value()
	if selected != "" {
		util.SetHistory(appConfig, USER_HISTORY_FILE, util.AddToHistory(history, selected))
	}
	return normalizeReviewers(selected)
}

func normalizeReviewers(selected string) string {
	selected = strings.ReplaceAll(selected, " ", "#")
	selected = strings.ReplaceAll(selected, ",", "#")
	expression := regexp.MustCompile("#+")
	selected = expression.ReplaceAllString(selected, ",")
	return selected
}

func updateSuggestions(program *tea.Program) {
	// Add recent reviewers first so they are auto-completed first.
	suggestions := getRecentReviewers()
	program.Send(setSuggestionsMsg{suggestions: suggestions})
	suggestions = slices.AppendSeq(suggestions, slices.Values(getAllCollaborators()))
	program.Send(setSuggestionsMsg{suggestions: suggestions})
}

func allUsersFromHistory(history []string) []string {
	allUsers := make([]string, 0, len(history))
	for _, next := range history {
		users := strings.FieldsFunc(next, func(next rune) bool {
			return slices.Contains(getBreakingChars(), next)
		})
		allUsers = slices.AppendSeq(allUsers, slices.Values(users))
	}
	slices.Sort(allUsers)
	return slices.Compact(allUsers)
}

func getBreakingChars() []rune {
	return []rune{' ', ','}
}
