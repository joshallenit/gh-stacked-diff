package interactive

import (
	"fmt"

	bubbletable "github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	lipglosstable "github.com/charmbracelet/lipgloss/table"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

var baseStyle = lipgloss.NewStyle()

var highlightedStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("229")).
	Background(lipgloss.Color("57")).
	Bold(false)

type model struct {
	table       bubbletable.Model
	selectedRow int
	windowWidth int
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

func totalWidth(columns []bubbletable.Column) int {
	totalSize := 0
	for _, column := range columns {
		totalSize += column.Width
	}
	return totalSize
}

func (m model) View() string {
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
			case row == m.table.Cursor():
				return highlightedStyle
			default:
				return baseStyle
			}
		}).Width(m.windowWidth - 2)
	return renderTable.Render() + "\n"
}

// Returns empty array if the user cancelled.
func GetTableSelection(prompt string, columns []string, rows [][]string) int {
	tableColumns := make([]bubbletable.Column, len(columns))
	for i, columnName := range columns {
		tableColumns[i] = bubbletable.Column{Title: (columnName + columnName + columnName + columnName + columnName + columnName), Width: 1}
	}

	tableRows := make([]bubbletable.Row, len(rows))
	for i, rowData := range rows {
		tableRows[i] = rowData
	}

	for i, _ := range rows {
		tableRows = append(tableRows, bubbletable.Row{"fake data " + fmt.Sprint(i), "fake data 2", "fake data 3"})
	}

	t := bubbletable.New(
		bubbletable.WithColumns(tableColumns),
		bubbletable.WithRows(tableRows),
		bubbletable.WithFocused(true),
		bubbletable.WithHeight(min(len(rows)+3, 10)),
	)

	s := bubbletable.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true)
	t.SetStyles(s)

	initialModel := model{table: t, selectedRow: -1}
	finalModel, err := tea.NewProgram(initialModel).Run()
	if err != nil {
		panic(err)
	}
	return finalModel.(model).selectedRow
}
