package interactive

import (
	"slices"

	"io"

	table "github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

var highlightEnabledStyle = baseStyle.
	Foreground(lipgloss.Color("229")).
	Background(lipgloss.Color("57"))

var highlightDisabledStyle = baseStyle.
	Foreground(lipgloss.Color("240")).
	Background(lipgloss.Color("244"))

var enabledRowStyle = baseStyle

var disabledRowStyle = baseStyle.
	Foreground(lipgloss.Color("240"))

var selectedRowStyle = baseStyle.Bold(true)

var selectedHighlightRowStyle = highlightEnabledStyle.Bold(true)

type model struct {
	table        table.Model
	selectedRows []int
	multiselect  bool
	rowEnabled   func(row int) bool
	completed    bool
	prompt       string
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.table.SetStyleFunc(m.createStyleFunc())
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "Q", "ctrl+c":
			m.selectedRows = []int{}
			return m, tea.Quit
		case " ":
			if !m.multiselect || !m.rowEnabled(m.table.Cursor()) {
				break
			}
			existingIndex := slices.Index(m.selectedRows, m.table.Cursor())
			if existingIndex != -1 {
				m.selectedRows = slices.Delete(m.selectedRows, existingIndex, existingIndex+1)
			} else {
				m.selectedRows = append(m.selectedRows, m.table.Cursor())
			}
			return m, nil
		case "enter":
			if !m.rowEnabled(m.table.Cursor()) {
				break
			}
			if !slices.Contains(m.selectedRows, m.table.Cursor()) {
				m.selectedRows = append(m.selectedRows, m.table.Cursor())
			}
			m.completed = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.table.SetHeight(min(max(msg.Height-10, 5), 20))
		return m, nil
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// This needs to be recreated everytime the model changes so that the model reference is updated.
func (m model) createStyleFunc() func(tableModel table.Model, row int, col int) lipgloss.Style {
	return func(tableModel table.Model, row int, col int) lipgloss.Style {
		switch {
		case row < 0 || row >= len(tableModel.Rows()):
			// < 0 is the header row
			// >= len can happen on resize
			return baseStyle
		case row == tableModel.Cursor():
			if m.rowEnabled(row) {
				if slices.Contains(m.selectedRows, row) {
					return selectedHighlightRowStyle
				} else {
					return highlightEnabledStyle
				}
			} else {
				return highlightDisabledStyle
			}
		case slices.Contains(m.selectedRows, row):
			return selectedRowStyle
		default:
			if m.rowEnabled(row) {
				return enabledRowStyle
			} else {
				return disabledRowStyle
			}
		}
	}
}

func (m model) View() string {
	if m.completed {
		return ""
	}
	return promptStyle.Render(m.prompt) + "\n" + m.table.View() + "\n"
}

// Returns empty selection if the user cancelled.
func GetTableSelection(prompt string, columns []string, rows [][]string, multiselect bool, stdIn io.Reader, rowEnabled func(row int) bool) []int {
	tableColumns := util.MapSlice(columns, func(columnName string) table.Column {
		return table.Column{Title: columnName}
	})

	tableRows := make([]table.Row, len(rows))
	firstEnabledRow := -1
	for i, rowData := range rows {
		tableRows[i] = rowData
		if firstEnabledRow == -1 && rowEnabled(i) {
			firstEnabledRow = i
		}
	}
	t := table.New(
		table.WithColumns(tableColumns),
		table.WithRows(tableRows),
		table.WithFocused(true),
		table.WithWrapCursor(true),
	)
	if firstEnabledRow != -1 {
		t.SetCursor(firstEnabledRow)
	}

	initialModel := model{
		table:        t,
		selectedRows: []int{},
		multiselect:  multiselect,
		rowEnabled:   rowEnabled,
		prompt:       prompt,
	}
	finalModel, err := NewProgram(initialModel, tea.WithInput(stdIn)).Run()
	if err != nil {
		panic(err)
	}
	selected := finalModel.(model).selectedRows
	slices.Sort(selected)
	return selected
}
