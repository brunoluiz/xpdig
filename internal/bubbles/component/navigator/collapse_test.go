package navigator

import (
	"io"
	"log/slog"
	"testing"

	"github.com/brunoluiz/xpdig/internal/bubbles/component/table"
	"github.com/charmbracelet/bubbles/textinput"
)

const (
	rootID  = "root"
	rootKey = "/0:r"
)

func newTestModel() Model {
	m := New(slog.New(slog.NewTextHandler(io.Discard, nil)), table.New(), textinput.New())
	m.SetColumns([]table.Column{{Title: "OBJECT", Width: 20}})
	m.SetData([]DataRow{
		{ID: rootID, Key: rootID, Columns: []string{rootID}},
		{ID: "a", Key: "a", ParentKey: rootID, Columns: []string{"a"}},
		{ID: "a1", Key: "a1", ParentKey: "a", Columns: []string{"a1"}},
		{ID: "b", Key: "b", ParentKey: rootID, Columns: []string{"b"}},
	})
	return m
}

func TestCollapseExpand(t *testing.T) {
	m := newTestModel()
	if len(m.visible) != 4 {
		t.Fatalf("want 4 visible, got %d", len(m.visible))
	}

	m.setAllCollapsed(true)
	if len(m.visible) != 3 {
		t.Fatalf("collapse all: want 3 visible, got %d", len(m.visible))
	}

	m.setAllCollapsed(false)
	m.cursor = 1
	m.setCollapsed(true)
	if len(m.visible) != 3 || !m.collapsed["a"] {
		t.Fatalf("collapse selected failed: %d", len(m.visible))
	}
	m.toggleSelected()
	if len(m.visible) != 4 {
		t.Fatalf("toggle should expand, got %d", len(m.visible))
	}

	m.cursor = 2 // a1 is a leaf: collapsing does nothing
	m.setCollapsed(true)
	if m.Current().Key != "a1" || len(m.visible) != 4 || len(m.collapsed) != 0 {
		t.Fatalf("leaf collapse should be a no-op: cursor on %s, %d visible", m.Current().Key, len(m.visible))
	}
}

func TestDuplicateIDsCollapseIndependently(t *testing.T) {
	m := New(slog.New(slog.NewTextHandler(io.Discard, nil)), table.New(), textinput.New())
	m.SetColumns([]table.Column{{Title: "OBJECT", Width: 20}})
	m.SetData([]DataRow{
		{ID: "r", Key: rootKey, Columns: []string{"r"}},
		{ID: "x", Key: rootKey + "/0:x", ParentKey: rootKey, Columns: []string{"x"}},
		{ID: "c", Key: rootKey + "/0:x/0:c", ParentKey: rootKey + "/0:x", Columns: []string{"c"}},
		{ID: "x", Key: rootKey + "/1:x", ParentKey: rootKey, Columns: []string{"x"}},
		{ID: "c", Key: rootKey + "/1:x/0:c", ParentKey: rootKey + "/1:x", Columns: []string{"c"}},
	})
	m.cursor = 1
	m.setCollapsed(true)
	if len(m.visible) != 4 {
		t.Fatalf("only the first x should collapse, got %d visible", len(m.visible))
	}
}

func TestEmptyAndRefresh(t *testing.T) {
	m := newTestModel()
	m.cursor = 1
	m.setCollapsed(true)

	// refresh keeps collapse state and cursor
	m.SetData(m.data)
	if len(m.visible) != 3 || m.Current().Key != "a" {
		t.Fatalf("state lost on refresh: %d visible, on %s", len(m.visible), m.Current().Key)
	}

	// refresh with an empty/shrunk tree must not panic
	m.SetData(nil)
	m.setAllCollapsed(true)
	m.toggleSelected()
	_ = m.Current()
}
