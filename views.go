package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle      = lipgloss.NewStyle().Margin(1, 2)
	quitTextStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	selected      = map[string]bool{}
)

func (i choice) Title() string {

	if selected[i.name] {
		return "[x] " + i.name
	}
	return "[ ] " + i.name
}
func (i choice) String() string      { return i.name }
func (i choice) Description() string { return "    " + i.desc }
func (i choice) FilterValue() string { return i.name }

type listKeyMap struct {
	togglechoice  key.Binding
	finishchoices key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		togglechoice: key.NewBinding(
			key.WithKeys(""),
			key.WithHelp("<space>", "toggle choice"),
		),
		finishchoices: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("<enter>", "execute simulations"),
		),
	}
}

type model struct {
	list list.Model
	keys *listKeyMap
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == " " {
			i, ok := m.list.SelectedItem().(choice)
			if ok {
				if selected[i.name] {
					selected[i.name] = false
				} else {
					selected[i.name] = true
				}
			}
			return m, nil
		}

		if msg.String() == "enter" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		top, right, bottom, left := docStyle.GetMargin()
		m.list.SetSize(msg.Width-left-right, msg.Height-top-bottom)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func showChoices(_ context.Context, choices []choice) ([]string, error) {
	items := []list.Item{}
	for _, c := range choices {
		items = append(items, c)
		selected[c.name] = true
	}

	m := model{
		list: list.New(items, list.NewDefaultDelegate(), 0, 0),
	}

	m.list.Title = "ioc-bench"

	p := tea.NewProgram(m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	names := []string{}
	for name, enabled := range selected {
		if enabled {
			names = append(names, name)
		}
	}

	sort.Strings(names)
	return names, nil
}

func announce(title string) {
	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		MarginTop(1).
		PaddingLeft(4).
		PaddingRight(4)

	fmt.Println(style.Render(title))
}

type errMsg error

type spinModel struct {
	text    string
	spinner spinner.Model
	stop    bool
	err     error
}

func createSpinModel() spinModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return spinModel{spinner: s}
}

func (m spinModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	quitMsg := tea.Quit()

	switch msg := msg.(type) {
	case errMsg:
		m.err = msg
		return m, nil
	default:
		if msg == quitMsg {
			m.stop = true
			return m, tea.Quit
		}

		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

}

func (m spinModel) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	str := fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), m.text)
	if m.stop {
		return str + "\n"
	}
	return str
}

func createSpinner(text string) *tea.Program {
	m := createSpinModel()
	m.text = text
	p := tea.NewProgram(m)

	go func() {
		if err := p.Start(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	return p
}
