package game

import (
	"encoding/json"
	"sort"
	"time"
)

// AchievementID uniquely identifies an achievement.
type AchievementID int

const (
	GettingWood       AchievementID = 1
	TimeToMine        AchievementID = 2
	HotTopic          AchievementID = 3
	MonsterHunter     AchievementID = 4
	Diamonds          AchievementID = 5
	CowTipper         AchievementID = 6
	BakeBread         AchievementID = 7
	GettingAnUpgrade  AchievementID = 8
	AcquireHardware   AchievementID = 9
	IntoFire          AchievementID = 10
)

// Achievement represents a single in-game achievement.
type Achievement struct {
	ID          AchievementID `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Unlocked    bool          `json:"unlocked"`
	UnlockedAt  time.Time     `json:"unlocked_at"`
}

// AchievementManager tracks the state of all achievements.
type AchievementManager struct {
	achievements map[AchievementID]*Achievement

	// OnUnlock is called when an achievement is newly unlocked.
	OnUnlock func(Achievement)
}

// achievementDef holds the static definition used to register achievements.
type achievementDef struct {
	ID          AchievementID
	Name        string
	Description string
}

var defaultAchievements = []achievementDef{
	{GettingWood, "Getting Wood", "Punch a tree until a block of wood pops out"},
	{TimeToMine, "Time to Mine!", "Use planks and sticks to make a pickaxe"},
	{HotTopic, "Hot Topic", "Construct a furnace out of eight cobblestone blocks"},
	{MonsterHunter, "Monster Hunter", "Attack and destroy a monster"},
	{Diamonds, "DIAMONDS!", "Acquire diamonds with your iron tools"},
	{CowTipper, "Cow Tipper", "Harvest some leather"},
	{BakeBread, "Bake Bread", "Turn wheat into bread"},
	{GettingAnUpgrade, "Getting an Upgrade", "Construct a better pickaxe"},
	{AcquireHardware, "Acquire Hardware", "Smelt an iron ingot"},
	{IntoFire, "Into Fire", "Relieve a Blaze of its rod"},
}

// NewAchievementManager creates an AchievementManager with all achievements
// registered in their locked state.
func NewAchievementManager() *AchievementManager {
	m := &AchievementManager{
		achievements: make(map[AchievementID]*Achievement, len(defaultAchievements)),
	}
	for _, def := range defaultAchievements {
		m.achievements[def.ID] = &Achievement{
			ID:          def.ID,
			Name:        def.Name,
			Description: def.Description,
		}
	}
	return m
}

// Unlock marks the achievement as unlocked and returns true if it was newly
// unlocked. Returns false if the achievement was already unlocked or the ID
// is unknown.
func (m *AchievementManager) Unlock(id AchievementID) bool {
	a, ok := m.achievements[id]
	if !ok {
		return false
	}
	if a.Unlocked {
		return false
	}
	a.Unlocked = true
	a.UnlockedAt = time.Now()
	if m.OnUnlock != nil {
		m.OnUnlock(*a)
	}
	return true
}

// IsUnlocked reports whether the given achievement has been unlocked.
func (m *AchievementManager) IsUnlocked(id AchievementID) bool {
	a, ok := m.achievements[id]
	if !ok {
		return false
	}
	return a.Unlocked
}

// Progress returns the number of unlocked achievements and the total count.
func (m *AchievementManager) Progress() (unlocked, total int) {
	total = len(m.achievements)
	for _, a := range m.achievements {
		if a.Unlocked {
			unlocked++
		}
	}
	return unlocked, total
}

// AllAchievements returns a snapshot of all achievements sorted by ID.
func (m *AchievementManager) AllAchievements() []Achievement {
	result := make([]Achievement, 0, len(m.achievements))
	for _, a := range m.achievements {
		result = append(result, *a)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

// achievementJSON is the serialization format for save/load.
type achievementJSON struct {
	ID         AchievementID `json:"id"`
	Unlocked   bool          `json:"unlocked"`
	UnlockedAt time.Time     `json:"unlocked_at"`
}

// MarshalJSON serializes only the unlock state of each achievement.
func (m *AchievementManager) MarshalJSON() ([]byte, error) {
	entries := make([]achievementJSON, 0, len(m.achievements))
	for _, a := range m.achievements {
		if a.Unlocked {
			entries = append(entries, achievementJSON{
				ID:         a.ID,
				Unlocked:   true,
				UnlockedAt: a.UnlockedAt,
			})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ID < entries[j].ID
	})
	return json.Marshal(entries)
}

// UnmarshalJSON restores unlock state from saved data, preserving the
// registered achievement definitions.
func (m *AchievementManager) UnmarshalJSON(data []byte) error {
	var entries []achievementJSON
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}
	for _, e := range entries {
		if a, ok := m.achievements[e.ID]; ok && e.Unlocked {
			a.Unlocked = true
			a.UnlockedAt = e.UnlockedAt
		}
	}
	return nil
}
