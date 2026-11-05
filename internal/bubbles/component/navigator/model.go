package navigator

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/brunoluiz/xpdig/internal/bubbles/component/table"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type searchMode int

type DataRow struct {
	ID        string
	Key       string // unique tree-path key, used for collapse state (ID may repeat)
	ParentKey string
	Data      any

	Columns []string
	Color   lipgloss.TerminalColor
}

const (
	searchModeOff searchMode = iota
	searchModeInit
	searchModeInput
	searchModeFilter
)

type State struct {
	LastTransitionTime time.Time
	Status             string
}

type ColorConfig struct {
	Foreground lipgloss.ANSIColor
	Background lipgloss.ANSIColor
}

type Model struct {
	KeyMap      KeyMap
	Styles      Styles
	Help        help.Model
	table       table.Model
	searchInput textinput.Model
	logger      *slog.Logger

	width  int
	height int
	cursor int

	showHelp             bool
	searchMode           searchMode
	searchResult         string
	searchCursor         int
	cursorBySearchCursor map[int]int
	searchCursorByCursor map[int]int

	data      []DataRow
	visible   []DataRow
	collapsed map[string]bool

	childCounts map[string]int
	parents     map[string]string
}

func New(
	logger *slog.Logger,
	tableModel table.Model,
	searchInputModel textinput.Model,
) Model {
	searchInputModel.Prompt = "🔍 "
	searchInputModel.Placeholder = "Search..."
	return Model{
		logger:      logger,
		table:       tableModel,
		searchInput: searchInputModel,
		KeyMap:      DefaultKeyMap(),
		Styles:      DefaultStyles(),

		width:  0,
		height: 0,

		showHelp: true,
		Help:     help.New(),

		collapsed:            map[string]bool{},
		searchCursor:         0,
		cursorBySearchCursor: map[int]int{},
		searchCursorByCursor: map[int]int{},
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {
	components := []string{}
	availableHeight := m.height

	if m.showHelp {
		help := m.helpView()
		availableHeight -= lipgloss.Height(help)
		components = append(components, help)
	}

	switch m.searchMode {
	case searchModeInit:
		fallthrough
	case searchModeInput:
		searchBar := lipgloss.NewStyle().Render(m.searchInput.View())
		availableHeight -= lipgloss.Height(searchBar)
		components = append(components, searchBar)
	case searchModeFilter:
		filterBar := lipgloss.NewStyle().Render(fmt.Sprintf("🔍 Showing results for: %s", m.searchInput.Value()))
		availableHeight -= lipgloss.Height(filterBar)
		components = append(components, filterBar)
	}

	m.table.SetHeight(availableHeight)
	tree := m.table.View()

	return lipgloss.JoinVertical(lipgloss.Left, append([]string{tree}, components...)...)
}

func (m *Model) SetData(data []DataRow) {
	m.data = data
	m.childCounts = make(map[string]int, len(data))
	m.parents = make(map[string]string, len(data))
	for _, d := range data {
		m.parents[d.Key] = d.ParentKey
		if d.ParentKey != "" {
			m.childCounts[d.ParentKey]++
		}
	}
	m.refreshVisible("")
	m.doLoadTable()
}

// refreshVisible rebuilds the visible rows from the collapsed state, keeping the
// cursor on the row identified by keepID (or the current row when empty).
func (m *Model) refreshVisible(keepKey string) {
	if keepKey == "" && m.cursor < len(m.visible) {
		keepKey = m.visible[m.cursor].Key
	}

	known := map[string]bool{}
	for _, d := range m.data {
		known[d.Key] = true
	}
	for id := range m.collapsed {
		if !known[id] {
			delete(m.collapsed, id)
		}
	}

	hidden := map[string]bool{}
	m.visible = m.visible[:0:0]
	for _, d := range m.data {
		if d.ParentKey != "" && (m.collapsed[d.ParentKey] || hidden[d.ParentKey]) {
			hidden[d.Key] = true
			continue
		}
		m.visible = append(m.visible, d)
	}

	m.cursor = 0
	for i, d := range m.visible {
		if d.Key == keepKey {
			m.cursor = i
			break
		}
	}
	m.table.SetCursor(m.cursor)
	m.searchCursor = 0
	m.cursorBySearchCursor = map[int]int{}
	m.searchCursorByCursor = map[int]int{}
	if m.searchMode == searchModeFilter {
		m.doSearchIndex()
	}
}

func (m *Model) doLoadTable() {
	rows := []table.Row{}
	searchTerm := strings.ToLower(m.searchInput.Value())
	for k, v := range m.visible {
		s := lipgloss.NewStyle()
		if m.cursor != k {
			s = s.Foreground(v.Color)
		}

		if m.searchMode == searchModeFilter && strings.Contains(strings.ToLower(v.ID), searchTerm) {
			s = s.Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("1")) // White on red
		}

		cols := []table.Cell{}
		for i, col := range v.Columns {
			if i == 0 && m.collapsed[v.Key] && m.childCounts[v.Key] > 0 {
				col = fmt.Sprintf("%s ▸ (+%d)", col, m.childCounts[v.Key])
			}
			cols = append(cols, table.Cell{Value: col, Style: s})
		}

		rows = append(rows, cols)
	}

	m.table.SetRows(rows)
	m.table.Focus()
}

func (m *Model) SetColumns(cc []table.Column) {
	m.table.SetColumns(cc)
	cols := m.table.Columns()

	// Adding `2` due to borders and all
	if len(cols) > 2 {
		w := 0
		for _, col := range cols[:len(cols)-1] {
			w += col.Width + 3
		}
		cols[len(cols)-1].Width = (m.width - w + 2)
	}
}

func (m Model) ShortHelp() []key.Binding {
	k := m.KeyMap
	return append([]key.Binding{},
		k.Up, k.Down, k.Copy,
		k.Describe, k.Get, k.Edit, k.Delete, k.RemoveFinalizers,
		k.Pause, k.Unpause,
		k.Toggle, k.CollapseAll, k.ExpandAll,
		k.Search, k.Help, k.Quit,
	)
}

func (m Model) FullHelp() [][]key.Binding {
	k := m.KeyMap
	kb := [][]key.Binding{{k.Up, k.Down, k.Copy, k.Describe}}
	return append(kb, []key.Binding{k.Quit, k.CloseFullHelp})
}

func (m Model) Current() *DataRow {
	if len(m.visible) == 0 {
		return &DataRow{}
	}
	return &m.visible[min(max(m.cursor, 0), len(m.visible)-1)]
}
func (m *Model) setSize(width, height int) { m.width = width; m.height = height }

func (m Model) helpView() string {
	m.Help.ShowAll = false
	return m.Styles.Help.Render(m.Help.View(m))
}
