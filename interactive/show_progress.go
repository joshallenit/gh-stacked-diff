package interactive

import (
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

type progressIndicatorModel struct {
	progressBar progress.Model
	cancelled   bool
}

func (m progressIndicatorModel) Init() tea.Cmd {
	return nil
}

func (m progressIndicatorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit
		}
	case setProgressMsg:
		m.progressBar.SetPercent(msg.progress)
		// case tea.WindowSizeMsg:

	}
	var updatedProgressBar tea.Model
	updatedProgressBar, cmd = m.progressBar.Update(msg)
	m.progressBar = updatedProgressBar.(progress.Model)
	return m, cmd
}

func (m progressIndicatorModel) View() string {
	return m.progressBar.View()
}

type setProgressMsg struct {
	progress float64
}

type ProgressIndicator struct {
	program *tea.Program
}

var _ tea.Model = progressIndicatorModel{}
var _ tea.Msg = setProgressMsg{}

// Blocks until Quit is called.
func (p *ProgressIndicator) Show(appConfig util.AppConfig) {
	finalModel := runProgram(appConfig.Io, p.program)
	if !finalModel.(progressIndicatorModel).cancelled {
		appConfig.Exit(0)
	}
}

func (p *ProgressIndicator) SetProgress(progress float64) {
	p.program.Send(setProgressMsg{progress: progress})
}

func (p *ProgressIndicator) Quit() {
	p.program.Send(tea.Quit())
}

func NewProgressIndicator(stdIo util.StdIo) *ProgressIndicator {
	initialModel := progressIndicatorModel{}
	return &ProgressIndicator{
		program: newProgram(initialModel, stdIo),
	}
}
