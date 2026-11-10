package main

// The start screen, shown when el_poblador runs with no command (as a
// launcher such as av starts it): choose a new game for three or four
// players and name them, or continue a saved game from game.SaveDir().

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"git.sr.ht/~vicho/el_poblador/game"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type savedGame struct {
	path    string
	when    time.Time
	players string
}

type setupResult struct {
	names    []string // a new game with these players
	loadPath string   // or this saved game
}

type setupScreen int

const (
	screenMenu setupScreen = iota
	screenNames
	screenSaves
)

type setupModel struct {
	screen  setupScreen
	cursor  int
	saves   []savedGame
	count   int // 3 or 4 players
	inputs  []textinput.Model
	focus   int // 0..count-1 names, count = Start
	result  *setupResult
	width   int
	height  int
	problem string
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	pickStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
)

const banner = "E L   P O B L A D O R"

func findSaves() []savedGame {
	var saves []savedGame
	matches, _ := filepath.Glob(filepath.Join(game.SaveDir(), "game_save_*.gob"))
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		s := savedGame{path: path, when: info.ModTime()}
		if g, err := loadGameState(path); err == nil {
			names := make([]string, len(g.Players))
			for i, p := range g.Players {
				names[i] = p.Name
			}
			s.players = strings.Join(names, ", ")
		} else {
			continue // not a game this version can open
		}
		saves = append(saves, s)
	}
	sort.Slice(saves, func(i, j int) bool { return saves[i].when.After(saves[j].when) })
	return saves
}

func newSetup() *setupModel {
	m := &setupModel{saves: findSaves(), count: 3}
	for i := 0; i < 4; i++ {
		ti := textinput.New()
		ti.Placeholder = fmt.Sprintf("Player %d", i+1)
		ti.CharLimit = 16
		ti.Width = 18
		ti.Prompt = ""
		m.inputs = append(m.inputs, ti)
	}
	return m
}

func (m *setupModel) menuItems() []string {
	items := []string{"New game"}
	if len(m.saves) > 0 {
		items = append(items, "Continue a saved game")
	}
	return append(items, "Quit")
}

func (m *setupModel) Init() tea.Cmd { return nil }

func (m *setupModel) focusName(i int) tea.Cmd {
	m.focus = i
	for j := range m.inputs {
		m.inputs[j].Blur()
	}
	if i < m.count {
		return m.inputs[i].Focus()
	}
	return nil
}

func (m *setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.screen {
		case screenMenu:
			return m.updateMenu(msg)
		case screenSaves:
			return m.updateSaves(msg)
		case screenNames:
			return m.updateNames(msg)
		}
	}
	return m, nil
}

func (m *setupModel) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := m.menuItems()
	switch msg.String() {
	case "up", "k":
		m.cursor = (m.cursor + len(items) - 1) % len(items)
	case "down", "j":
		m.cursor = (m.cursor + 1) % len(items)
	case "q", "esc":
		return m, tea.Quit
	case "enter", " ":
		switch items[m.cursor] {
		case "New game":
			m.screen = screenNames
			return m, m.focusName(0)
		case "Continue a saved game":
			m.screen, m.cursor = screenSaves, 0
		case "Quit":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *setupModel) updateSaves(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.cursor = (m.cursor + len(m.saves) - 1) % len(m.saves)
	case "down", "j":
		m.cursor = (m.cursor + 1) % len(m.saves)
	case "esc", "q":
		m.screen, m.cursor = screenMenu, 1
	case "enter", " ":
		m.result = &setupResult{loadPath: m.saves[m.cursor].path}
		return m, tea.Quit
	}
	return m, nil
}

func (m *setupModel) updateNames(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.problem = ""
	switch msg.String() {
	case "esc":
		m.screen, m.cursor = screenMenu, 0
		return m, m.focusName(m.count)
	case "up", "shift+tab":
		return m, m.focusName((m.focus + m.count) % (m.count + 1))
	case "down", "tab":
		return m, m.focusName((m.focus + 1) % (m.count + 1))
	case "left", "right":
		if m.focus == m.count { // on Start, left and right choose 3 or 4 players
			m.count = 7 - m.count
			m.focus = m.count
			return m, nil
		}
	case "enter":
		if m.focus < m.count {
			return m, m.focusName(m.focus + 1)
		}
		names := make([]string, m.count)
		seen := map[string]bool{}
		for i := 0; i < m.count; i++ {
			names[i] = strings.TrimSpace(m.inputs[i].Value())
			if names[i] == "" {
				names[i] = m.inputs[i].Placeholder
			}
			if seen[strings.ToLower(names[i])] {
				m.problem = "Every player needs a different name."
				return m, nil
			}
			seen[strings.ToLower(names[i])] = true
		}
		m.result = &setupResult{names: names}
		return m, tea.Quit
	}
	if m.focus < m.count {
		var cmd tea.Cmd
		m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *setupModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(banner) + "\n\n")
	b.WriteString(dimStyle.Render("Settle the island, trade, build, and be first to 10 points.") + "\n\n")

	line := func(selected bool, text string) {
		if selected {
			b.WriteString(pickStyle.Render("> "+text) + "\n")
		} else {
			b.WriteString("  " + text + "\n")
		}
	}

	switch m.screen {
	case screenMenu:
		for i, item := range m.menuItems() {
			line(i == m.cursor, item)
		}
		b.WriteString("\n" + dimStyle.Render("Up and down to choose, Enter to pick."))
	case screenSaves:
		b.WriteString("Saved games\n\n")
		for i, s := range m.saves {
			line(i == m.cursor, fmt.Sprintf("%s   %s", s.when.Format("2006-01-02 15:04"), s.players))
		}
		b.WriteString("\n" + dimStyle.Render("Enter to continue, Esc to go back."))
	case screenNames:
		b.WriteString(fmt.Sprintf("New game for %d players, all at this keyboard\n\n", m.count))
		for i := 0; i < m.count; i++ {
			mark := "  "
			if i == m.focus {
				mark = pickStyle.Render("> ")
			}
			b.WriteString(fmt.Sprintf("%sPlayer %d  %s\n", mark, i+1, m.inputs[i].View()))
		}
		b.WriteString("\n")
		start := fmt.Sprintf("Start   < %d players >", m.count)
		line(m.focus == m.count, start)
		if m.problem != "" {
			b.WriteString("\n" + pickStyle.Render(m.problem) + "\n")
		}
		b.WriteString("\n" + dimStyle.Render("Type the names (empty keeps the grey one), Tab or arrows to move,\nleft and right on Start for 3 or 4 players, Enter to start, Esc to go back.\nIn the game, 1 to 4 shows that player's hand, 0 follows the turn."))
	}

	out := b.String()
	out = lipgloss.NewStyle().Width(lipgloss.Width(out)).Align(lipgloss.Left).Render(out)
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, out)
	}
	return out
}

// runSetup shows the start screen and returns what was chosen, or nil to quit.
func runSetup() (*setupResult, error) {
	m := newSetup()
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		return nil, err
	}
	return m.result, nil
}
