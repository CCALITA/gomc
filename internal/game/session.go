package game

import (
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/player"
	"github.com/fanxiyao/gomc/internal/world"
)

// GameSession holds all per-world state that exists only while a world is
// loaded. It is nil when the player is at the main menu.
type GameSession struct {
	World       *world.World
	Player      *player.Controller
	ECSWorld    *ecs.World
	Scheduler   *ecs.Scheduler
	Spawner     *entity.Spawner
	Storage     *world.Storage
	ChunkLoader *world.ChunkLoader
	Inventory   *inventory.Inventory
	SpawnPoint  mcmath.Vec3
}
