package interactive

import (
	"regexp"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

const REVIEWERS_HISTORY_FILE = "reviewers.history"
const all_collaborators_file = "all-collaborators.cache"

type userSelectionModel struct {
	textInput     textinput.Model
	history       []string
	suggestions   []string
	breakingChars []rune
	historyIndex  int
	confirmed     bool
	windowWidth   int
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
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
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
	const USER_PREFIX = "   users     "
	users := strings.Join(m.getMatchingSuggestions(), " ")
	if len(users)+len(USER_PREFIX) > m.windowWidth {
		users = users[0:min(max(0, m.windowWidth-len(USER_PREFIX)), len(users))]
	}
	users = USER_PREFIX + users + "\n"
	return promptStyle.Render("Reviewers to add when checks pass?") + "\n" +
		m.textInput.View() + "\n" +
		"\n" +
		"Controls:\n" +
		"   up/down   history\n" +
		"   tab       select auto-complete\n" +
		"   enter     confirm\n" +
		"   esc       quit\n" +
		users
}

// Returns the suggestions that match the current wip text.
func (m *userSelectionModel) getMatchingSuggestions() []string {
	csv, wipText := m.splitInput()
	matchingSuggestions := util.FilterSlice(m.suggestions, func(next string) bool {
		// more lenient than m.textInput.MatchingSuggestions
		return strings.Contains(strings.ToUpper(next), strings.ToUpper(wipText))
	})
	selectedFields := strings.FieldsFunc(csv, func(next rune) bool {
		return slices.Contains(m.breakingChars, next)
	})
	return slices.DeleteFunc(matchingSuggestions, func(next string) bool {
		return slices.Contains(selectedFields, next)
	})
}

// Sets suggestions so users can be added to an existing comma delimited string.
func (m *userSelectionModel) setSuggestions() {
	csv, _ := m.splitInput()
	if csv != "" {
		selectedFields := strings.FieldsFunc(m.textInput.Value(), func(next rune) bool {
			return slices.Contains(m.breakingChars, next)
		})
		nonSelectedSuggestions := util.FilterSlice(m.suggestions, func(next string) bool {
			return !slices.Contains(selectedFields, next)
		})
		m.textInput.SetSuggestions(util.MapSlice(nonSelectedSuggestions, func(next string) string {
			return csv + next
		}))
	} else {
		m.textInput.SetSuggestions(m.suggestions)
	}
}

// Returns the current text input split between the CSV portion and the wip text.
func (m *userSelectionModel) splitInput() (string, string) {
	valueRunes := []rune(m.textInput.Value())
	for i := len(valueRunes) - 1; i >= 0; i-- {
		if slices.Contains(m.breakingChars, valueRunes[i]) {
			return string(valueRunes[0 : i+1]), string(valueRunes[i+1:])
		}
	}
	return "", m.textInput.Value()
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

func UserSelection(asyncConfig util.AsyncAppConfig) string {
	input := textinput.New()
	input.Focus()
	input.Width = 100
	input.Placeholder = "None"
	input.ShowSuggestions = true
	history := util.ReadHistory(asyncConfig.App, REVIEWERS_HISTORY_FILE)
	suggestions := util.ReadHistory(asyncConfig.App, all_collaborators_file)
	input.SetSuggestions(suggestions)
	initialModel := userSelectionModel{
		history:       history,
		historyIndex:  -1,
		textInput:     input,
		confirmed:     false,
		suggestions:   suggestions,
		breakingChars: []rune{',', ' '},
	}
	program := newProgram(initialModel, asyncConfig.App.Io)
	go updateSuggestions(asyncConfig, program)
	finalModel, err := runProgram(asyncConfig.App.Io, program)
	if err != nil {
		panic(err)
	}
	finalSelectionModel := finalModel.(userSelectionModel)
	if !finalSelectionModel.confirmed {
		asyncConfig.App.Exit(0)
	}
	selected := finalSelectionModel.textInput.Value()
	if selected != "" {
		util.SetHistory(asyncConfig.App, REVIEWERS_HISTORY_FILE, util.AddToHistory(history, selected))
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

// Updates suggestions with results from API collaborators call.
func updateSuggestions(asyncConfig util.AsyncAppConfig, program *tea.Program) {
	defer asyncConfig.GracefulRecover()
	allCollaborators := getAllCollaborators()
	program.Send(setSuggestionsMsg{suggestions: allCollaborators})
	util.SetHistory(asyncConfig.App, all_collaborators_file, allCollaborators)
}
