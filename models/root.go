package models

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type view uint
type progressTick struct{}

var (
	focusedModelStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color("69"))
	modelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("69"))
	progressPercentPerTick = 0.10
	progressTickDuration   = 100 * time.Millisecond
)

const (
	treeView view = iota
	detailsView
)

type rootKeymap struct {
	tab, quit, refresh, debug key.Binding
}

type viewportKeymaps struct {
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
}

func newViewportKeymaps() viewportKeymaps {
	return viewportKeymaps{
		HalfPageUp: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "page down"),
		),
	}
}

func (v viewportKeymaps) reassignViewportKeymap(m *viewport.KeyMap) {
	m.HalfPageUp = v.HalfPageUp
	m.HalfPageDown = v.HalfPageDown
}

type RootModel struct {
	Tree            *TreeModel
	treeViewport    viewport.Model
	treeStyle       lipgloss.Style
	detailsViewport viewport.Model
	detailsStyle    lipgloss.Style
	view            view
	height          int
	width           int
	help            help.Model
	keymap          rootKeymap
	viewportKeymap  viewportKeymaps
	progress        tea.Model
	showProgress    bool
}

func NewRootModel() *RootModel {
	details := viewport.New(0, 0)
	tree := viewport.New(0, 0)

	viewportKeymap := newViewportKeymaps()
	viewportKeymap.reassignViewportKeymap(&tree.KeyMap)
	viewportKeymap.reassignViewportKeymap(&details.KeyMap)

	m := &RootModel{
		Tree:            NewTreeModel(),
		treeViewport:    tree,
		detailsViewport: details,
		view:            treeView,
		help:            help.New(),
		keymap: rootKeymap{
			tab: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "next window"),
			),
			quit: key.NewBinding(
				key.WithKeys("esc", "ctrl+c"),
				key.WithHelp("esc", "quit"),
			),
			refresh: key.NewBinding(
				key.WithKeys("r"),
				key.WithHelp("r", "refresh"),
			),
			debug: key.NewBinding(
				key.WithKeys("d"),
				key.WithHelp("d", "debug"),
			),
		},
		viewportKeymap: viewportKeymap,
		progress:       progress.New(progress.WithDefaultScaledGradient()),
		showProgress:   true,
	}
	return m
}

func (m *RootModel) Init() tea.Cmd {
	return tea.Batch(
		m.Tree.Init(),
		m.detailsViewport.Init(),
		m.progress.Init(),
		func() tea.Msg {
			return progressTick{}
		},
	)
}

func (m *RootModel) updateViewports(msg tea.Msg) tea.Cmd {
	if m.view == treeView {
		newViewport, cmd := m.treeViewport.Update(msg)
		m.treeViewport = newViewport
		return cmd
	}
	newDetailsViewport, cmd := m.detailsViewport.Update(msg)
	m.detailsViewport = newDetailsViewport
	return cmd
}

func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit
		case key.Matches(msg, m.keymap.tab):
			if m.view == treeView {
				m.view = detailsView
			} else {
				m.view = treeView
			}
			m.resizeElements()
		case key.Matches(msg, m.viewportKeymap.HalfPageUp):
			cmds = append(cmds, m.updateViewports(msg))
		case key.Matches(msg, m.viewportKeymap.HalfPageDown):
			cmds = append(cmds, m.updateViewports(msg))
		case key.Matches(msg, m.keymap.refresh):
			cmds = append(cmds, m.Tree.CurNode.RefreshDetail())
		}

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		if m.showProgress {
			p := m.progress.(progress.Model)
			p.Width = m.width
			m.progress = p
		}
		m.resizeElements()
	case progressTick:
		p := m.progress.(progress.Model)
		if p.Percent() >= 1.0 {
			m.showProgress = false
		} else {
			cmds = append(cmds, p.IncrPercent(progressPercentPerTick))
			m.progress = p
			cmds = append(cmds, func() tea.Msg {
				time.Sleep(progressTickDuration)
				return progressTick{}
			})
		}
		return m, tea.Batch(cmds...)
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		cmds = append(cmds, cmd)
		m.progress = progressModel.(progress.Model)
		return m, tea.Batch(cmds...)
	}

	if !m.showProgress && m.view == treeView {
		newTree, cmd := m.Tree.Update(msg)
		m.Tree = newTree.(*TreeModel)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *RootModel) resizeElements() {
	var (
		height              int = m.height - 4
		width               int = m.width - 4
		detailsHeightOffset int = 1
		treeHeightOffset    int = 1
		treeStyle           lipgloss.Style
		detailsStyle        lipgloss.Style
	)

	if m.view == treeView {
		treeStyle = focusedModelStyle
		detailsStyle = modelStyle
	} else {
		treeStyle = modelStyle
		detailsStyle = focusedModelStyle
	}

	m.treeStyle = treeStyle.
		Height(height - treeHeightOffset).
		Width(width / 2)
	m.treeViewport.Height = m.treeStyle.GetHeight()
	m.treeViewport.Width = m.treeStyle.GetWidth()

	m.detailsStyle = detailsStyle.
		Height(height - detailsHeightOffset).
		Width((width - m.treeStyle.GetWidth()))
	m.detailsViewport.Height = m.detailsStyle.GetHeight()
	m.detailsViewport.Width = m.detailsStyle.GetWidth()
}

func (m *RootModel) View() string {
	var s string
	if m.showProgress {
		return m.progress.View()
	}
	m.detailsViewport.SetContent(m.Tree.CurNode.GetDetail())
	m.treeViewport.SetContent(m.Tree.View())

	help := m.help.ShortHelpView([]key.Binding{
		m.keymap.tab,
		m.keymap.quit,
		m.keymap.refresh,
		m.keymap.debug,
	})
	treeHelp := m.help.ShortHelpView([]key.Binding{
		m.treeViewport.KeyMap.HalfPageUp,
		m.treeViewport.KeyMap.HalfPageDown,
		m.Tree.Keymap.Left,
		m.Tree.Keymap.Right,
		m.Tree.Keymap.Up,
		m.Tree.Keymap.Down,
	})
	detailsHelp := m.help.ShortHelpView([]key.Binding{
		m.detailsViewport.KeyMap.HalfPageUp,
		m.detailsViewport.KeyMap.HalfPageDown,
	})
	treeViewport := lipgloss.JoinVertical(lipgloss.Top, m.treeViewport.View(), treeHelp)
	detailsViewport := lipgloss.JoinVertical(lipgloss.Top, m.detailsViewport.View(), detailsHelp)
	s += lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.treeStyle.Render(treeViewport),
		m.detailsStyle.Render(detailsViewport),
	)
	s += "\n\n" + help
	return s
}
