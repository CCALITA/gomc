package world

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fanxiyao/gomc/internal/mcmath"
)

func TestSignManager_SetGetText(t *testing.T) {
	m := NewSignManager()
	pos := mcmath.BlockPos{X: 10, Y: 64, Z: -5}
	lines := [4]string{"Hello", "World", "", ""}

	m.SetSignText(pos, lines)

	got, ok := m.GetSignText(pos)
	if !ok {
		t.Fatal("expected sign to exist after SetSignText")
	}
	if got.Lines != lines {
		t.Errorf("got lines %v, want %v", got.Lines, lines)
	}
}

func TestSignManager_GetMissing(t *testing.T) {
	m := NewSignManager()
	pos := mcmath.BlockPos{X: 0, Y: 0, Z: 0}

	_, ok := m.GetSignText(pos)
	if ok {
		t.Error("expected ok=false for missing sign")
	}
}

func TestSignManager_RemoveSign(t *testing.T) {
	m := NewSignManager()
	pos := mcmath.BlockPos{X: 1, Y: 2, Z: 3}
	m.SetSignText(pos, [4]string{"A", "", "", ""})

	m.RemoveSign(pos)

	_, ok := m.GetSignText(pos)
	if ok {
		t.Error("expected sign to be removed")
	}
}

func TestSignManager_RemoveNonexistent(t *testing.T) {
	m := NewSignManager()
	// Should not panic.
	m.RemoveSign(mcmath.BlockPos{X: 99, Y: 99, Z: 99})
}

func TestSignManager_TruncateLongLines(t *testing.T) {
	m := NewSignManager()
	pos := mcmath.BlockPos{X: 5, Y: 5, Z: 5}
	long := "1234567890123456789" // 19 chars, should truncate to 15
	m.SetSignText(pos, [4]string{long, "", "", ""})

	got, ok := m.GetSignText(pos)
	if !ok {
		t.Fatal("expected sign to exist")
	}
	if len([]rune(got.Lines[0])) != maxSignLineLength {
		t.Errorf("line length = %d, want %d", len([]rune(got.Lines[0])), maxSignLineLength)
	}
	if got.Lines[0] != "123456789012345" {
		t.Errorf("truncated line = %q, want %q", got.Lines[0], "123456789012345")
	}
}

func TestSignManager_ExactMaxLength(t *testing.T) {
	m := NewSignManager()
	pos := mcmath.BlockPos{X: 0, Y: 0, Z: 0}
	exact := "123456789012345" // exactly 15 chars
	m.SetSignText(pos, [4]string{exact, "", "", ""})

	got, _ := m.GetSignText(pos)
	if got.Lines[0] != exact {
		t.Errorf("line = %q, want %q", got.Lines[0], exact)
	}
}

func TestSignManager_AllSigns(t *testing.T) {
	m := NewSignManager()
	pos1 := mcmath.BlockPos{X: 1, Y: 1, Z: 1}
	pos2 := mcmath.BlockPos{X: 2, Y: 2, Z: 2}
	m.SetSignText(pos1, [4]string{"A", "", "", ""})
	m.SetSignText(pos2, [4]string{"B", "", "", ""})

	all := m.AllSigns()
	if len(all) != 2 {
		t.Fatalf("expected 2 signs, got %d", len(all))
	}
	if all[pos1].Lines[0] != "A" || all[pos2].Lines[0] != "B" {
		t.Error("sign data mismatch in AllSigns snapshot")
	}
}

func TestSignManager_Overwrite(t *testing.T) {
	m := NewSignManager()
	pos := mcmath.BlockPos{X: 0, Y: 64, Z: 0}
	m.SetSignText(pos, [4]string{"First", "", "", ""})
	m.SetSignText(pos, [4]string{"Second", "", "", ""})

	got, _ := m.GetSignText(pos)
	if got.Lines[0] != "Second" {
		t.Errorf("expected overwritten text %q, got %q", "Second", got.Lines[0])
	}
}

func TestStorage_SaveLoadSigns(t *testing.T) {
	dir := t.TempDir()
	storage, err := NewStorage(dir)
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}

	signs := map[mcmath.BlockPos]SignData{
		{X: 10, Y: 64, Z: -5}: {Lines: [4]string{"Hello", "World", "", "Line4"}},
		{X: 0, Y: 0, Z: 0}:   {Lines: [4]string{"", "", "", ""}},
	}

	if err := storage.SaveSigns(signs); err != nil {
		t.Fatalf("SaveSigns: %v", err)
	}

	// Verify the file exists.
	path := filepath.Join(dir, "signs.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("signs.json not created: %v", err)
	}

	loaded, err := storage.LoadSigns()
	if err != nil {
		t.Fatalf("LoadSigns: %v", err)
	}

	if len(loaded) != len(signs) {
		t.Fatalf("loaded %d signs, want %d", len(loaded), len(signs))
	}

	for pos, expected := range signs {
		got, ok := loaded[pos]
		if !ok {
			t.Errorf("missing sign at %v", pos)
			continue
		}
		if got.Lines != expected.Lines {
			t.Errorf("at %v: got %v, want %v", pos, got.Lines, expected.Lines)
		}
	}
}

func TestStorage_LoadSigns_NoFile(t *testing.T) {
	dir := t.TempDir()
	storage, err := NewStorage(dir)
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}

	// Loading when no signs.json exists should return empty map, no error.
	signs, err := storage.LoadSigns()
	if err != nil {
		t.Fatalf("LoadSigns: unexpected error: %v", err)
	}
	if len(signs) != 0 {
		t.Errorf("expected empty map, got %d entries", len(signs))
	}
}

func TestSignManager_SaveLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	storage, err := NewStorage(dir)
	if err != nil {
		t.Fatalf("NewStorage: %v", err)
	}

	m := NewSignManager()
	pos := mcmath.BlockPos{X: -100, Y: 128, Z: 200}
	lines := [4]string{"Line 1", "Line 2", "Line 3", "Line 4"}
	m.SetSignText(pos, lines)

	// Save via storage.
	if err := storage.SaveSigns(m.AllSigns()); err != nil {
		t.Fatalf("SaveSigns: %v", err)
	}

	// Load into a fresh manager.
	loaded, err := storage.LoadSigns()
	if err != nil {
		t.Fatalf("LoadSigns: %v", err)
	}

	m2 := NewSignManager()
	m2.LoadSigns(loaded)

	got, ok := m2.GetSignText(pos)
	if !ok {
		t.Fatal("expected sign after roundtrip")
	}
	if got.Lines != lines {
		t.Errorf("roundtrip lines = %v, want %v", got.Lines, lines)
	}
}
