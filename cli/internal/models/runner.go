package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spotdemo4/trevstack/cli/internal/apps"
	"github.com/spotdemo4/trevstack/cli/internal/utils"
)

type runner struct {
	width  *int
	height *int

	prefix    lipgloss.Style
	checkmark string
	xmark     string

	header  *Header
	cbox    *Cbox
	help    *Help
	spinner spinner.Model

	msgChan chan apps.Msg
	msgs    []apps.Msg
	apps    []*apps.App
}

func NewRunner(msgChan chan apps.Msg, applications []*apps.App) *runner {

	prefix := lipgloss.NewStyle().
		Padding(0, 1, 0, 1).
		Margin(0, 1, 0, 1).
		Background(lipgloss.Color("#89dceb")).
		Foreground(lipgloss.Color("#11111b"))

	checkmark := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a6e3a1")).
		Bold(true).
		Render("✓")

	xmark := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f38ba8")).
		Bold(true).
		Render("✕")

	mpl := 0
	for _, app := range applications {
		if len(app.Name) > mpl {
			mpl = len(app.Name)
		}
	}

	return &runner{
		width:  nil,
		height: nil,

		prefix:    prefix,
		checkmark: checkmark,
		xmark:     xmark,

		header:  NewHeader(),
		cbox:    NewCbox(mpl + 1),
		help:    NewHelp(),
		spinner: spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#a6adc8")))),

		msgChan: msgChan,
		msgs:    []apps.Msg{},
		apps:    applications,
	}
}

func (m runner) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			return apps.Msg(<-m.msgChan)
		},
	)
}

func (m runner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case apps.Msg:
		// Remove old message with the same key
		if msg.Key != nil && msg.Loading != nil && !*msg.Loading {
			for i, prev := range m.msgs {
				if prev.Key != nil && prev.Loading != nil && *prev.Key == *msg.Key && *prev.Loading {
					m.msgs = append(m.msgs[:i], m.msgs[i+1:]...)
					break
				}
			}
		}

		// Set current state
		if msg.Loading != nil {
			if *msg.Loading {
				msg.App.Loading = nil
			} else {
				msg.App.Loading = msg.Success
			}
		}

		// Append new message
		m.msgs = append(m.msgs, msg)

		return m, func() tea.Msg {
			return apps.Msg(<-m.msgChan)
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = utils.IntPointer(msg.Width)
		m.height = utils.IntPointer(msg.Height)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.help.keys.Help):
			m.help.Toggle()

		case key.Matches(msg, m.help.keys.Quit):
			return m, tea.Quit
		}

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelDown:
			m.cbox.Viewport.LineDown(1)

		case tea.MouseButtonWheelUp:
			m.cbox.Viewport.LineUp(1)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m runner) term() (rows []string) {
	if m.width == nil {
		return rows
	}

	for _, msg := range m.msgs {
		item := []string{}

		if msg.Loading != nil && *msg.Loading {
			item = append(item, m.spinner.View())
		}

		if msg.Success != nil {
			if *msg.Success {
				item = append(item, m.checkmark)
			} else {
				item = append(item, m.xmark)
			}
		}

		item = append(item, msg.Text)
		itemStr := strings.Join(item, " ")

		// Render the row
		rows = append(rows, m.cbox.GenItem(msg.Time, msg.App.Name, itemStr, msg.App.Color, *m.width))
	}

	return rows
}

func (m runner) head() (items []string) {
	for _, app := range m.apps {
		item := []string{}

		if app.Loading == nil {
			item = append(item, m.spinner.View())
		} else if *app.Loading {
			item = append(item, m.checkmark)
		} else {
			item = append(item, m.xmark)
		}

		item = append(item, app.Name)
		items = append(items, m.header.GenItem(strings.Join(item, " ")))
	}

	return items
}

func (m runner) View() string {
	if m.width == nil || m.height == nil {
		return fmt.Sprintf("\n %s Loading...", m.spinner.View())
	}

	// Generate the UI
	header := m.header.Gen(*m.width, m.head()...)
	footer := m.help.Gen(*m.width)
	main := m.cbox.Gen(
		strings.Join(m.term(), "\n"),
		*m.width,
		*m.height,
		lipgloss.Height(header),
		lipgloss.Height(footer),
	)

	s := header
	s += main
	s += footer

	// Send the UI for rendering
	return s
}
