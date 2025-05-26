package interactive

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshallenit/gh-stacked-diff/v2/templates"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

type dashboardRow struct {
	index  string
	pr     bool
	status *util.PullRequestStatus
	log    templates.GitLog
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
		checksPassed := ""
		approved := ""
		if row.pr {
			pr = "pr"
			if row.status != nil {
				if row.status.Checks.IsSuccess() {
					checksPassed = "passesd"
				} else if row.status.Checks.IsFailing() {
					checksPassed = "failed"
				} else {
					checksPassed = fmt.Sprint(row.status.Checks.PercentageComplete())
				}
				approved = strings.Join(row.status.Approvers, "\n")
			} else {
				checksPassed = m.spinner.View()
				approved = m.spinner.View()
			}
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

func ShowDashboard(asyncConfig util.AsyncAppConfig, minChecks int) {

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
			index: indexString, pr: hasLocalBranch, log: log, status: nil,
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
	go updateDashboardData(asyncConfig, program, rows, minChecks)
	runProgram(asyncConfig.App.Io, program)
	// finalModel := runProgram(asyncConfig.App.Io, program)
	// finalDashboardModel := finalModel.(dashboardModel)
	// println("finalDashboardModel", fmt.Sprint(finalDashboardModel))
}

func updateDashboardData(asyncConfig util.AsyncAppConfig, program *tea.Program, rows []dashboardRow, minChecks int) {
	defer asyncConfig.GracefulRecover()
	for i, row := range rows {
		if row.pr {
			status := util.GetPullRequestStatus(row.log.Branch, minChecks)
			row.status = &status
			program.Send(updateDashboardRowMsg{index: i, row: row})

		}
	}
}
