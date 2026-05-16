package navigator

import (
	"strings"

	"github.com/brunoluiz/xpdig/internal/bubbles/component/table"
	"github.com/charmbracelet/bubbles/key"

	tea "github.com/charmbracelet/bubbletea"
)

type EventItemCopied struct {
	ID   string
	Data any
}

type EventItemDescribe struct {
	ID   string
	Data any
}

type EventItemFocused struct {
	ID   string
	Data any
}

type EventItemGet struct {
	ID   string
	Data any
}

type EventItemEdit struct {
	ID   string
	Data any
}

type EventItemDelete struct {
	ID   string
	Data any
}

type EventItemRemoveFinalizers struct {
	ID   string
	Data any
}

type EventItemPause struct {
	ID   string
	Data any
}

type EventItemUnpause struct {
	ID   string
	Data any
}

type EventQuitted struct{}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmd = m.onResize(msg)
	case tea.KeyMsg:
		cmd = m.onKey(msg)
	case table.EventCursorUpdated:
		cmd = m.onCursorUpdated(msg)
	}

	switch m.searchMode {
	case searchModeInput:
		var searchCmd tea.Cmd
		m.searchInput, searchCmd = m.searchInput.Update(msg)
		return m, searchCmd
	case searchModeInit:
		m.searchMode = searchModeInput
	}

	var tableCmd tea.Cmd
	m.table, tableCmd = m.table.Update(msg)

	return m, tea.Batch(cmd, tableCmd)
}

func (m *Model) onResize(msg tea.WindowSizeMsg) tea.Cmd {
	m.setSize(msg.Width, msg.Height)
	m.table.SetWidth(msg.Width)
	m.table.SetHeight(msg.Height)
	m.SetColumns(m.table.Columns())
	return nil
}

func (m *Model) onCursorUpdated(msg table.EventCursorUpdated) tea.Cmd {
	m.cursor = msg.Current
	m.doLoadTable()
	return func() tea.Msg {
		return EventItemFocused{ID: m.Current().ID, Data: m.Current().Data}
	}
}

func (m *Model) onSearch(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.KeyMap.SearchConfirm):
		m.searchResult = m.searchInput.Value()
		m.searchInput.Blur()
		m.searchMode = searchModeFilter
		m.doSearch()
		m.doLoadTable()
	case key.Matches(msg, m.KeyMap.SearchQuit):
		m.searchInput.Blur()
		m.searchMode = searchModeOff
		m.searchResult = ""
		m.searchInput.Reset()
	}
	return nil
}

func (m *Model) doSearch() {
	searchTerm := strings.ToLower(m.searchInput.Value())
	m.cursorBySearchCursor = map[int]int{}
	m.searchCursorByCursor = map[int]int{}

	match := 0
	for pos, v := range m.data {
		if strings.Contains(strings.ToLower(v.ID), searchTerm) {
			m.cursorBySearchCursor[match] = pos
			m.searchCursorByCursor[pos] = match
			match++
		}
	}

	if len(m.cursorBySearchCursor) > 0 {
		m.searchCursor = 0
		m.cursor = m.cursorBySearchCursor[0]
		m.table.SetCursor(m.cursor)
	}
}

func (m *Model) onSearchInit() {
	m.searchMode = searchModeInit
	m.searchInput.Reset()
	m.searchInput.Focus()
	m.searchResult = ""
}

func (m *Model) onSearchQuit() {
	if m.searchMode == searchModeOff {
		return
	}
	m.searchInput.Blur()
	m.searchInput.Reset()
	m.searchMode = searchModeOff
	m.searchResult = ""
	m.searchCursor = 0
	m.cursorBySearchCursor = map[int]int{}
	m.searchCursorByCursor = map[int]int{}
	m.doLoadTable()
}

func (m *Model) onSearchNext() tea.Cmd {
	if len(m.cursorBySearchCursor) == 0 {
		return nil
	}

	switch {
	// Behaviour within boundaries of search highlighted range
	// If m.cursor is within the highlighted range, resets the position to the cursor itself.
	// This will be the point of reference to be used.
	case m.cursorBySearchCursor[0] <= m.cursor && m.cursor <= m.cursorBySearchCursor[len(m.cursorBySearchCursor)-1]:
		m.searchCursor = m.searchCursorByCursor[m.cursor]
		m.searchCursor++
		// Behaviour for out of left boundary
	case m.cursor < m.cursorBySearchCursor[0]:
		m.searchCursor = 0
		// Behaviour for out of right boundary
	case m.cursor >= m.cursorBySearchCursor[len(m.cursorBySearchCursor)-1]:
		m.searchCursor = 0
	default:
		m.searchCursor++
	}

	if m.searchCursor >= len(m.cursorBySearchCursor) {
		m.searchCursor = 0 // Wrap around to the first result
	}

	m.cursor = m.cursorBySearchCursor[m.searchCursor]
	m.table.SetCursor(m.cursor)
	return func() tea.Msg {
		return EventItemFocused{ID: m.Current().ID, Data: m.Current().Data}
	}
}

func (m *Model) onSearchPrev() tea.Cmd {
	if len(m.cursorBySearchCursor) == 0 {
		return nil
	}

	switch {
	// Behaviour within boundaries of search highlighted range
	// If m.cursor is within the highlighted range, resets the position to the cursor itself.
	// This will be the point of reference to be used.
	case (m.cursorBySearchCursor[0] <= m.cursor && m.cursor <= m.cursorBySearchCursor[len(m.cursorBySearchCursor)-1]):
		m.searchCursor = m.searchCursorByCursor[m.cursor]
		m.searchCursor--
		// Behaviour for out of left boundary
	case m.cursor < m.cursorBySearchCursor[0]:
		m.searchCursor = 0
		// Behaviour for out of right boundary
	case m.cursor >= m.cursorBySearchCursor[len(m.cursorBySearchCursor)-1]:
		m.searchCursor = len(m.cursorBySearchCursor) - 1
	default:
		m.searchCursor--
	}

	if m.searchCursor < 0 {
		m.searchCursor = len(m.cursorBySearchCursor) - 1 // Wrap around to the last result
	}

	m.cursor = m.cursorBySearchCursor[m.searchCursor]
	m.table.SetCursor(m.cursor)
	return func() tea.Msg {
		return EventItemFocused{ID: m.Current().ID, Data: m.Current().Data}
	}
}

func (m *Model) onKey(msg tea.KeyMsg) tea.Cmd {
	if m.searchMode == searchModeInput {
		return m.onSearch(msg)
	}

	switch {
	case key.Matches(msg, m.KeyMap.Search):
		m.onSearchInit()
	case key.Matches(msg, m.KeyMap.Help):
		m.showHelp = !m.showHelp
		// m.Help.ShowAll = !m.Help.ShowAll
	case key.Matches(msg, m.KeyMap.Copy):
		return func() tea.Msg {
			return EventItemCopied{ID: m.Current().ID, Data: m.Current().Data}
		}
	case key.Matches(msg, m.KeyMap.Get):
		return func() tea.Msg {
			return EventItemGet{ID: m.Current().ID, Data: m.Current().Data}
		}
	case key.Matches(msg, m.KeyMap.Delete):
		return func() tea.Msg {
			return EventItemDelete{ID: m.Current().ID, Data: m.Current().Data}
		}
	case key.Matches(msg, m.KeyMap.RemoveFinalizers):
		return func() tea.Msg {
			return EventItemRemoveFinalizers{ID: m.Current().ID, Data: m.Current().Data}
		}
	case key.Matches(msg, m.KeyMap.Edit):
		return func() tea.Msg {
			return EventItemEdit{ID: m.Current().ID, Data: m.Current().Data}
		}
	case key.Matches(msg, m.KeyMap.SearchQuit):
		m.onSearchQuit()
	case key.Matches(msg, m.KeyMap.Describe):
		return func() tea.Msg {
			return EventItemDescribe{ID: m.Current().ID, Data: m.Current().Data}
		}
	case key.Matches(msg, m.KeyMap.Pause):
		return func() tea.Msg {
			return EventItemPause{ID: m.Current().ID, Data: m.Current().Data}
		}
	case key.Matches(msg, m.KeyMap.Unpause):
		return func() tea.Msg {
			return EventItemUnpause{ID: m.Current().ID, Data: m.Current().Data}
		}
	case key.Matches(msg, m.KeyMap.Quit):
		return func() tea.Msg {
			return EventQuitted{}
		}
	case key.Matches(msg, m.KeyMap.SearchNext):
		return m.onSearchNext()
	case key.Matches(msg, m.KeyMap.SearchPrevious):
		return m.onSearchPrev()
	}
	return nil
}
