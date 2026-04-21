package game

import (
	"encoding/json"
	"testing"
)

func TestAllAchievementsRegistered(t *testing.T) {
	m := NewAchievementManager()
	all := m.AllAchievements()
	if len(all) != 10 {
		t.Fatalf("expected 10 achievements, got %d", len(all))
	}

	ids := map[AchievementID]bool{}
	for _, a := range all {
		ids[a.ID] = true
		if a.Name == "" {
			t.Errorf("achievement %d has empty name", a.ID)
		}
		if a.Description == "" {
			t.Errorf("achievement %d has empty description", a.ID)
		}
	}

	for id := AchievementID(1); id <= 10; id++ {
		if !ids[id] {
			t.Errorf("achievement ID %d not registered", id)
		}
	}
}

func TestUnlockNew(t *testing.T) {
	m := NewAchievementManager()
	if !m.Unlock(GettingWood) {
		t.Fatal("expected Unlock to return true for new unlock")
	}
	if !m.IsUnlocked(GettingWood) {
		t.Fatal("expected GettingWood to be unlocked")
	}
}

func TestUnlockDuplicateReturnsFalse(t *testing.T) {
	m := NewAchievementManager()
	m.Unlock(Diamonds)
	if m.Unlock(Diamonds) {
		t.Fatal("expected Unlock to return false for duplicate unlock")
	}
}

func TestUnlockUnknownID(t *testing.T) {
	m := NewAchievementManager()
	if m.Unlock(999) {
		t.Fatal("expected Unlock to return false for unknown ID")
	}
}

func TestProgressCount(t *testing.T) {
	m := NewAchievementManager()

	unlocked, total := m.Progress()
	if unlocked != 0 || total != 10 {
		t.Fatalf("expected 0/10, got %d/%d", unlocked, total)
	}

	m.Unlock(GettingWood)
	m.Unlock(TimeToMine)
	m.Unlock(HotTopic)

	unlocked, total = m.Progress()
	if unlocked != 3 || total != 10 {
		t.Fatalf("expected 3/10, got %d/%d", unlocked, total)
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	m := NewAchievementManager()
	m.Unlock(GettingWood)
	m.Unlock(Diamonds)
	m.Unlock(IntoFire)

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	m2 := NewAchievementManager()
	if err := json.Unmarshal(data, m2); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !m2.IsUnlocked(GettingWood) {
		t.Error("expected GettingWood to be unlocked after load")
	}
	if !m2.IsUnlocked(Diamonds) {
		t.Error("expected Diamonds to be unlocked after load")
	}
	if !m2.IsUnlocked(IntoFire) {
		t.Error("expected IntoFire to be unlocked after load")
	}
	if m2.IsUnlocked(CowTipper) {
		t.Error("expected CowTipper to remain locked after load")
	}

	unlocked, total := m2.Progress()
	if unlocked != 3 || total != 10 {
		t.Fatalf("expected 3/10 after load, got %d/%d", unlocked, total)
	}
}

func TestOnUnlockCallback(t *testing.T) {
	m := NewAchievementManager()
	var called AchievementID
	m.OnUnlock = func(a Achievement) {
		called = a.ID
	}

	m.Unlock(MonsterHunter)
	if called != MonsterHunter {
		t.Fatalf("expected OnUnlock called with MonsterHunter, got %d", called)
	}

	// Duplicate unlock should not trigger callback again.
	called = 0
	m.Unlock(MonsterHunter)
	if called != 0 {
		t.Fatal("expected OnUnlock not called on duplicate unlock")
	}
}

func TestIsUnlockedUnknownID(t *testing.T) {
	m := NewAchievementManager()
	if m.IsUnlocked(999) {
		t.Fatal("expected IsUnlocked to return false for unknown ID")
	}
}

func TestAllAchievementsSortedByID(t *testing.T) {
	m := NewAchievementManager()
	all := m.AllAchievements()
	for i := 1; i < len(all); i++ {
		if all[i].ID <= all[i-1].ID {
			t.Fatalf("achievements not sorted: ID %d before %d", all[i-1].ID, all[i].ID)
		}
	}
}
