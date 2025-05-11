package interactive

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshallenit/gh-stacked-diff/v2/templates"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

type dashboardRow struct {
	index        string
	pr           bool
	checksPassed *bool
	approved     []string
	log          templates.GitLog
}

type dashboardModel struct {
	spinner spinner.Model
	table   table.Model
	rows    []dashboardRow
}

func (m dashboardModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "Q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	case updateDashboardRowMsg:
		m.rows[msg.index] = msg.row
		return m, nil
	}
	var tableCmd tea.Cmd
	m.table, tableCmd = m.table.Update(msg)
	var spinnerCmd tea.Cmd
	m.spinner, spinnerCmd = m.spinner.Update(msg)
	return m, tea.Batch(tableCmd, spinnerCmd)
}

func (m dashboardModel) View() string {
	m.table.SetRows(m.getTableRows())
	return m.spinner.View() + " " + m.spinner.View() + "\n" + m.table.View() + "\n"
}

func (m dashboardModel) getTableRows() []table.Row {
	tableRows := make([]table.Row, len(m.rows))
	for i, row := range m.rows {
		var pr string = ""
		if row.pr {
			pr = "pr"
		}
		var checksPassed string
		if row.checksPassed == nil {
			if row.pr {
				checksPassed = m.spinner.View()
			} else {
				checksPassed = "-"
			}
		} else {
			if *row.checksPassed {
				checksPassed = "passed"
			} else {
				checksPassed = "failed"
			}
		}
		var approved string
		if row.approved == nil {
			if row.pr {
				approved = m.spinner.View()
			} else {
				approved = "-"
			}
		} else {
			approved = strings.Join(row.approved, " ")
		}
		tableRows[i] = table.Row{row.index, pr, checksPassed, approved, row.log.Commit, row.log.Subject + "\nnext 2"}
	}
	return tableRows
}

type updateDashboardRowMsg struct {
	index int
	row   dashboardRow
}

var _ tea.Model = dashboardModel{}
var _ tea.Msg = updateDashboardRowMsg{}

func ShowDashboard(asyncConfig util.AsyncAppConfig) {

	columns := []string{"Index", "PR", "Checks", "Approved", "Commit", "Summary"}
	newCommits := templates.GetNewCommits("HEAD")
	gitBranchArgs := make([]string, 0, len(newCommits)+2)
	gitBranchArgs = append(gitBranchArgs, "branch", "-l")
	for _, log := range newCommits {
		gitBranchArgs = append(gitBranchArgs, log.Branch)
	}
	prBranches := strings.Fields(util.ExecuteOrDie(util.ExecuteOptions{}, "git", gitBranchArgs...))

	rows := make([]dashboardRow, len(newCommits))
	for i, log := range newCommits {
		hasLocalBranch := slices.Contains(prBranches, log.Branch)
		indexString := fmt.Sprint(i + 1)
		paddingLen := len(fmt.Sprint(len(newCommits))) - len(indexString)
		indexString = strings.Repeat(" ", paddingLen) + indexString
		rows[i] = dashboardRow{
			index: indexString, pr: hasLocalBranch, checksPassed: nil, approved: nil, log: log,
		}
	}

	tableColumns := util.MapSlice(columns, func(columnName string) table.Column {
		return table.Column{Title: columnName}
	})
	t := table.New(
		table.WithColumns(tableColumns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithWrapCursor(true),
	)
	initialModel := dashboardModel{
		spinner: spinner.New(),
		table:   t,
		rows:    rows,
	}
	initialModel.spinner.Spinner = spinner.Dot
	program := newProgram(initialModel, asyncConfig.App.Io)
	go updateDashboardData(asyncConfig, program, rows)
	finalModel := runProgram(asyncConfig.App.Io, program)
	finalDashboardModel := finalModel.(dashboardModel)
	println("finalDashboardModel", fmt.Sprint(finalDashboardModel))
}

func updateDashboardData(asyncConfig util.AsyncAppConfig, program *tea.Program, rows []dashboardRow) {
	defer asyncConfig.GracefulRecover()
	println("here   len ", len(rows))
	for i, row := range rows {
		if row.pr {
			row.approved = util.GetAllApprovingUsers(row.log.Branch)
			slog.Warn("hi" + fmt.Sprint(row.approved))
			tea.Println("Approved value " + fmt.Sprint(row.approved))
			program.Send(updateDashboardRowMsg{index: i, row: row})
		} else {
			slog.Warn("h3i")
			tea.Println("not PR " + fmt.Sprint(row.approved))
		}
	}
}
