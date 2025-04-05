package interactive

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshallenit/gh-stacked-diff/v2/util"
)

type confirmModel struct {
	prompt    string
	confirmed bool
}

var _ tea.Model = confirmModel{}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "n", "N", "ctrl+c":
			return m, tea.Quit
		case "y", "Y":
			m.confirmed = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	return promptStyle.Render(m.prompt) + " (y/n): "
}

func ConfirmOrDie(appConfig util.AppConfig, prompt string) {
	initialModel := confirmModel{prompt: prompt}
	finalModel, err := NewProgram(initialModel, tea.WithInput(appConfig.Io.In)).Run()
	if err != nil {
		panic(err)
	}
	if !finalModel.(confirmModel).confirmed {
		appConfig.Exit(0)
	}
}
