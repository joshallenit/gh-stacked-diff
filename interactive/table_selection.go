package interactive

import (
	bubbletable "github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	lipglosstable "github.com/charmbracelet/lipgloss/table"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

var baseStyle = lipgloss.NewStyle()

var highlightEnabledStyle = baseStyle.
	Foreground(lipgloss.Color("229")).
	Background(lipgloss.Color("57"))

var highlightDisabledStyle = baseStyle.
	Foreground(lipgloss.Color("240")).
	Background(lipgloss.Color("244"))

var enabledRow = baseStyle

var disabledRow = baseStyle.
	Foreground(lipgloss.Color("240"))

type model struct {
	table       bubbletable.Model
	selectedRow int
	windowWidth int
	multiselect bool
	rowEnabled  func(row int) bool
	maxRowWidth int
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			m.selectedRow = m.table.Cursor()
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		return m, nil
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	// Use bubbletea.table to support up/down to change selected row, not used for rendering.
	// Use lipgloss.table to support StyleFunc (which has not been ported to bubbletea.table yet)
	if m.selectedRow != -1 {
		return ""
	}
	columns := util.MapSlice(m.table.Columns(), func(column bubbletable.Column) string {
		return column.Title
	})
	rows := util.MapSlice(m.table.Rows(), func(row bubbletable.Row) []string {
		return row
	})
	renderTable := lipglosstable.New().Headers(columns...).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row < 0 || row >= len(rows):
				// < 0 is the header row
				// >= len can happen on resize
				return baseStyle
			case row == m.table.Cursor():
				if m.rowEnabled(row) {
					return highlightEnabledStyle
				} else {
					return highlightDisabledStyle
				}
			default:
				if m.rowEnabled(row) {
					return enabledRow
				} else {
					return disabledRow
				}
			}
		}).Width(min(m.maxRowWidth, m.windowWidth-2))
	return renderTable.Render() + "\n"
}

// Returns -1 if the user cancelled.
func GetTableSelection(prompt string, columns []string, rows [][]string, multiselect bool, rowEnabled func(row int) bool) int {
	tableColumns := make([]bubbletable.Column, len(columns))
	for i, columnName := range columns {
		tableColumns[i] = bubbletable.Column{Title: columnName, Width: len(columnName)}
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
	)
	if firstEnabledRow != -1 {
		t.SetCursor(firstEnabledRow)
	}

	initialModel := model{
		table:       t,
		selectedRow: -1,
		multiselect: multiselect,
		rowEnabled:  rowEnabled,
		maxRowWidth: totalWidth(t.Columns()),
	}
	finalModel, err := tea.NewProgram(initialModel).Run()
	if err != nil {
		panic(err)
	}
	return finalModel.(model).selectedRow
}

func totalWidth(columns []bubbletable.Column) int {
	totalSize := 0
	for _, column := range columns {
		totalSize += column.Width
	}
	// Each column has a border, plus 1 for the last border.
	return totalSize + len(columns) + 1
}
