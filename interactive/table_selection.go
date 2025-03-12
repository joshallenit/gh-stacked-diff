package interactive

import (
	"slices"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table       table.Model
	selectedRow int
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
		newColumns := resizeColumns(msg.Width-10, m.table.Columns(), m.table.Rows())
		m.table.SetColumns(newColumns)
		return m, nil
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func resizeColumns(maxTotalWidth int, columns []table.Column, rows []table.Row) []table.Column {
	resizedColumns := make([]table.Column, 0, len(columns))
	for _, column := range columns {
		resizedColumns = append(resizedColumns, table.Column{Title: column.Title, Width: len(column.Title)})
	}
	for _, row := range rows {
		for i, cell := range row {
			resizedColumns[i].Width = max(len(cell), resizedColumns[i].Width)
		}
	}
	for totalWidth(resizedColumns) > maxTotalWidth {
		resized := false
		for i, _ := range slices.Backward(resizedColumns) {
			if resizedColumns[i].Width > 1 {
				resizedColumns[i].Width = resizedColumns[i].Width - 1
				resized = true
				break
			}
		}
		if !resized {
			break
		}
	}
	return resizedColumns
}

func totalWidth(columns []table.Column) int {
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
	return baseStyle.Render(m.table.View()) + "\n"
}

// Returns empty array if the user cancelled.
func GetTableSelection(columns []string, rows [][]string) int {
	tableColumns := make([]table.Column, len(columns))
	for i, columnName := range columns {
		tableColumns[i] = table.Column{Title: columnName, Width: 1}
	}

	tableRows := make([]table.Row, len(rows))
	for i, rowData := range rows {
		tableRows[i] = rowData
	}

	t := table.New(
		table.WithColumns(tableColumns),
		table.WithRows(tableRows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	initialModel := model{table: t, selectedRow: -1}
	finalModel, err := tea.NewProgram(initialModel).Run()
	if err != nil {
		panic(err)
	}
	return finalModel.(model).selectedRow
}
