package interactive

import (
	"slices"

	bubbletable "github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle()

var promptStyle = baseStyle.Bold(true)
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
	table        bubbletable.Model
	selectedRows []int
	windowWidth  int
	multiselect  bool
	rowEnabled   func(row int) bool
	maxRowWidth  int
	completed    bool
	prompt       string
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "ctrl+c":
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
		m.windowWidth = msg.Width
		return m, nil
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	// Use bubbletea.table to support up/down to change selected row, not used for rendering.
	// Use lipgloss.table to support StyleFunc (which has not been ported to bubbletea.table yet)
	if m.completed {
		return ""
	}
	// oh no I need the resize capabilities

	// columns := util.MapSlice(m.table.Columns(), func(column bubbletable.Column) string {
	// 	return column.Title
	// })
	// rows := util.MapSlice(m.table.Rows(), func(row bubbletable.Row) []string {
	// 	return row
	// })
	// renderTable := lipglosstable.New().Headers(columns...).
	// 	Rows(rows...).
	// 	StyleFunc(func(row, col int) lipgloss.Style {
	// 		switch {
	// 		case row < 0 || row >= len(rows):
	// 			// < 0 is the header row
	// 			// >= len can happen on resize
	// 			return baseStyle
	// 		case row == m.table.Cursor():
	// 			if m.rowEnabled(row) {
	// 				if slices.Contains(m.selectedRows, row) {
	// 					return selectedHighlightRowStyle
	// 				} else {
	// 					return highlightEnabledStyle
	// 				}
	// 			} else {
	// 				return highlightDisabledStyle
	// 			}
	// 		case slices.Contains(m.selectedRows, row):
	// 			return selectedRowStyle
	// 		default:
	// 			if m.rowEnabled(row) {
	// 				return enabledRowStyle
	// 			} else {
	// 				return disabledRowStyle
	// 			}
	// 		}
	// 	}).
	// 	Width(min(m.maxRowWidth, m.windowWidth-2))
	return promptStyle.Render(m.prompt) + "\n" + m.table.View() + "\n"
}

// Returns -1 if the user cancelled.
func GetTableSelection(prompt string, columns []string, rows [][]string, multiselect bool, rowEnabled func(row int) bool) []int {
	tableColumns := make([]bubbletable.Column, len(columns))
	for i, columnName := range columns {
		tableColumns[i] = bubbletable.Column{Title: columnName, Width: 1}
	}

	tableRows := make([]bubbletable.Row, len(rows))
	firstEnabledRow := -1
	for i, rowData := range rows {
		tableRows[i] = rowData
		if firstEnabledRow == -1 && rowEnabled(i) {
			firstEnabledRow = i
		}
		for i, cell := range rowData {
			tableColumns[i].Width = max(tableColumns[i].Width, len(cell))
		}
	}

	t := bubbletable.New(
		bubbletable.WithColumns(tableColumns),
		bubbletable.WithRows(tableRows),
		bubbletable.WithFocused(true),
		bubbletable.WithStyleFunc(func(m bubbletable.Model, row, col int) lipgloss.Style {
			switch {
			case row < 0 || row >= len(rows):
				// < 0 is the header row
				// >= len can happen on resize
				return baseStyle
			case row == m.Cursor():
				if rowEnabled(row) {
					// if slices.Contains(m.selectedRows, row) {
					return selectedHighlightRowStyle
					// } else {
					// 	return highlightEnabledStyle
					// }
				} else {
					return highlightDisabledStyle
				}
			// case slices.Contains(m.selectedRows, row):
			// return selectedRowStyle
			default:
				if rowEnabled(row) {
					return enabledRowStyle
				} else {
					return disabledRowStyle
				}
			}
		}),
	)
	if firstEnabledRow != -1 {
		t.SetCursor(firstEnabledRow)
	}

	initialModel := model{
		table:        t,
		selectedRows: []int{},
		multiselect:  multiselect,
		rowEnabled:   rowEnabled,
		maxRowWidth:  totalWidth(t.Columns()),
		prompt:       prompt,
	}
	finalModel, err := tea.NewProgram(initialModel).Run()
	if err != nil {
		panic(err)
	}
	return finalModel.(model).selectedRows
}

func totalWidth(columns []bubbletable.Column) int {
	totalSize := 0
	for _, column := range columns {
		totalSize += column.Width
	}
	// Each column has a border, plus 1 for the last border.
	return totalSize + len(columns) + 1
}
