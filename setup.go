package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("78"))
	selStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
)

// setupModel asks for the default prefix on first launch.
type setupModel struct {
	input     textinput.Model
	err       string
	prefix    string
	cancelled bool
}

func newSetupModel() setupModel {
	ti := textinput.New()
	ti.Prompt = "❯ "
	ti.SetValue("~/src")
	ti.CursorEnd()
	ti.Focus()
	return setupModel{input: ti}
}

func (m setupModel) Init() tea.Cmd { return textinput.Blink }

func (m setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			v := strings.TrimSpace(m.input.Value())
			if v == "" {
				m.err = "enter a directory"
				return m, nil
			}
			if !isDir(expandHome(v)) {
				m.err = fmt.Sprintf("%s is not a directory", v)
				return m, nil
			}
			m.prefix = v
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.err = ""
	return m, cmd
}

func (m setupModel) View() string {
	if m.prefix != "" || m.cancelled {
		return ""
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render("cc-launch setup") + "\n")
	b.WriteString("What should your default prefix be? " + dimStyle.Render("(typing / in the launcher browses under it)") + "\n")
	b.WriteString(m.input.View() + "\n")
	if m.err != "" {
		b.WriteString(errStyle.Render(m.err) + "\n")
	} else {
		b.WriteString(dimStyle.Render("enter to save · esc to cancel") + "\n")
	}
	return b.String()
}

// runSetup returns the chosen prefix, or "" if the user cancelled.
func runSetup() (string, error) {
	final, err := tea.NewProgram(newSetupModel()).Run()
	if err != nil {
		return "", err
	}
	return final.(setupModel).prefix, nil
}
