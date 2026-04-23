package game

import (
	"fmt"
	"log"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/fanxiyao/gomc/internal/audio"
	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/chunk"
	"github.com/fanxiyao/gomc/internal/config"
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/player"
	"github.com/fanxiyao/gomc/internal/render"
	"github.com/fanxiyao/gomc/internal/ui"
	"github.com/fanxiyao/gomc/internal/world"
)

const (
	defaultTickRate      = 20.0
	defaultAutoSaveTicks = 6000 // 5 minutes at 20 TPS
	defaultSavePath      = "saves/world1"
)

type Game struct {
	Renderer    *render.Renderer
	World       *world.World
	ECSWorld    *ecs.World
	Player      *player.Controller
	Input       *input.Manager
	KeyMap      *input.KeyMap
	UI          *ui.UIManager
	UIRenderer  *ui.UIRenderer
	Audio       *audio.Engine
	Config      *config.Config
	ChunkLoader *world.ChunkLoader
	Scheduler   *ecs.Scheduler
	State       *StateManager
	Inventory   *inventory.Inventory

	meshQueue   chan mcmath.ChunkPos

	Mode        *ModeManager
	Spawner     *entity.Spawner
	Time        *TimeKeeper

	SpawnPoint  mcmath.Vec3

	Storage     *world.Storage
	Running       bool
	tickCount     int64
	tickRate      float64
	tickInterval  float64
	autoSaveTicks int64
}

func (g *Game) Init(cfg *config.Config) error {
	g.Config = cfg
	g.Running = true
	g.State = NewStateManager()
	g.Mode = NewModeManager()
	g.Time = NewTimeKeeper()

	// Derive tick timing from config, falling back to compile-time defaults.
	g.tickRate = cfg.Gameplay.TickRate
	if g.tickRate <= 0 {
		g.tickRate = defaultTickRate
	}
	g.tickInterval = 1.0 / g.tickRate

	saveTicks := cfg.Gameplay.AutoSaveIntervalTicks
	if saveTicks <= 0 {
		saveTicks = defaultAutoSaveTicks
	}
	g.autoSaveTicks = int64(saveTicks)

	block.InitRegistry()

	g.Renderer = &render.Renderer{}
	if err := g.Renderer.Init(cfg.Window.Width, cfg.Window.Height, cfg.Window.Title); err != nil {
		return fmt.Errorf("renderer init: %w", err)
	}

	g.Audio = audio.NewEngine(44100)
	if err := g.Audio.Init(); err != nil {
		log.Printf("audio init failed (continuing without sound): %v", err)
	}
	g.Audio.SetMasterVolume(cfg.Audio.MasterVolume)

	g.Input = input.NewManager()
	g.KeyMap = input.NewKeyMap()
	g.setupInputCallbacks()

	g.ECSWorld = ecs.NewWorld()
	g.Scheduler = ecs.NewScheduler()
	g.setupSystems()

	g.Inventory = inventory.NewInventory(36)
	g.UIRenderer = ui.NewUIRenderer(float32(cfg.Window.Width), float32(cfg.Window.Height))
	g.UI = ui.NewUIManager()
	g.setupUI()

	g.State.OnEnter(GameStatePlaying, func() {
		g.Renderer.Window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	})
	g.State.OnExit(GameStatePlaying, func() {
		g.Renderer.Window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	})

	return nil
}

func (g *Game) setupInputCallbacks() {
	w := g.Renderer.Window
	w.SetKeyCallback(func(_ *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		g.Input.KeyCallback(int(key), scancode, int(action), int(mods))
	})
	w.SetMouseButtonCallback(func(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		g.Input.MouseButtonCallback(int(button), int(action), int(mods))
	})
	w.SetCursorPosCallback(func(_ *glfw.Window, x, y float64) {
		g.Input.CursorPosCallback(x, y)
	})
	w.SetScrollCallback(func(_ *glfw.Window, xoff, yoff float64) {
		g.Input.ScrollCallback(xoff, yoff)
	})
}

func (g *Game) setupSystems() {
	g.Scheduler.Add(&entity.PhysicsSystem{
		GetBlockAABBs: func(region mcmath.AABB) []mcmath.AABB {
			if g.World == nil {
				return nil
			}
			return g.World.GetBlockAABBs(region)
		},
	})
	g.Scheduler.Add(&entity.AISystem{
		ChaseRange:  g.Config.Gameplay.AIChaseRange,
		AttackRange: g.Config.Gameplay.AIAttackRange,
	})
	g.Scheduler.Add(&entity.LifetimeSystem{})
	g.Scheduler.Add(&entity.DamageSystem{})
	g.Scheduler.Add(&entity.HealthSystem{})
	g.Scheduler.Add(&entity.HungerSystem{})
	g.Scheduler.Add(&entity.ItemPickupSystem{
		OnPickup: func(playerEntity ecs.Entity, itemID uint16, count int) bool {
			if g.Player == nil || playerEntity != g.Player.Entity {
				return false
			}
			if g.Inventory != nil {
				remainder := g.Inventory.AddItem(item.ItemStack{ItemID: itemID, Count: count})
				return remainder.IsEmpty()
			}
			return false
		},
	})
}

func (g *Game) setupUI() {
	hud := ui.NewHUD(g.Inventory)
	g.UI.PushScreen(hud)
}

func (g *Game) Run() {
	lastTime := glfw.GetTime()
	accumulator := 0.0

	for g.Running && !g.Renderer.Window.ShouldClose() {
		currentTime := glfw.GetTime()
		dt := currentTime - lastTime
		lastTime = currentTime
		if dt > 0.25 {
			dt = 0.25
		}
		accumulator += dt

		glfw.PollEvents()
		g.Input.Update()

		g.handleGlobalInput()

		for accumulator >= g.tickInterval {
			g.tick(g.tickInterval)
			accumulator -= g.tickInterval
		}

		g.render()
	}
}

func (g *Game) handleGlobalInput() {
	pauseKey := g.KeyMap.GetKey(input.Pause)
	if g.Input.IsKeyJustPressed(pauseKey) {
		switch g.State.CurrentState() {
		case GameStatePlaying:
			g.State.SetState(GameStatePaused)
			g.UI.PushScreen(ui.NewPauseMenu(
				func() {
					g.UI.PopScreen()
					g.State.SetState(GameStatePlaying)
				},
				func() {
					g.Running = false
				},
			))
		case GameStatePaused:
			g.UI.PopScreen()
			g.State.SetState(GameStatePlaying)
		}
	}
}

func (g *Game) tick(dt float64) {
	g.UI.Update(g.Input, dt)

	if g.State.CurrentState() != GameStatePlaying {
		return
	}

	g.tickCount++

	if g.World != nil && g.Player != nil {
		transform := ecs.GetStore[entity.Transform](g.ECSWorld)
		if t, ok := transform.Get(g.Player.Entity); ok {
			centerChunk := mcmath.BlockPos{
				X: int32(t.Position.X),
				Y: int32(t.Position.Y),
				Z: int32(t.Position.Z),
			}.ToChunkPos()
			g.ChunkLoader.Update(centerChunk)
		}

		g.Player.Update(g.Input, g.World, float32(dt))
		g.Player.UpdateCombat(g.Input, g.ECSWorld, float32(dt))
	}

	g.Scheduler.Update(g.ECSWorld, dt)

	g.Time = g.Time.Advance(1)

	if g.Spawner != nil && g.Player != nil {
		transform := ecs.GetStore[entity.Transform](g.ECSWorld)
		if t, ok := transform.Get(g.Player.Entity); ok {
			g.Spawner.SpawnCycle(g.ECSWorld, dt, t.Position)
		}
	}

	g.checkPlayerDeath()

	if g.autoSaveTicks > 0 && g.tickCount%g.autoSaveTicks == 0 {
		g.autoSave()
	}
}

// checkPlayerDeath checks the player's health component. If health is
// at or below zero, it transitions to the death state and pushes the
// death screen.
func (g *Game) checkPlayerDeath() {
	if g.Player == nil {
		return
	}

	healthStore := ecs.GetStore[entity.Health](g.ECSWorld)
	h, ok := healthStore.Get(g.Player.Entity)
	if !ok {
		return
	}

	if h.Current > 0 {
		return
	}

	g.State.SetState(GameStateDead)
	g.UI.PushScreen(ui.NewDeathScreen(func() {
		g.respawnPlayer()
	}))
}

// respawnPlayer restores the player to the spawn point with full health,
// pops the death screen, and returns to the playing state.
func (g *Game) respawnPlayer() {
	if g.Player == nil {
		return
	}

	// Restore health.
	healthStore := ecs.GetStore[entity.Health](g.ECSWorld)
	healthStore.Set(g.Player.Entity, entity.Health{Current: 20, Max: 20})

	// Reset position to spawn point.
	transformStore := ecs.GetStore[entity.Transform](g.ECSWorld)
	transformStore.Set(g.Player.Entity, entity.Transform{Position: g.SpawnPoint})

	pbStore := ecs.GetStore[entity.PhysicsBody](g.ECSWorld)
	if pb, ok := pbStore.Get(g.Player.Entity); ok {
		pb.Body.Position = g.SpawnPoint
		pb.Body.Velocity = mcmath.Vec3{}
	}

	// Remove any pending damage.
	ecs.GetStore[entity.Damage](g.ECSWorld).Remove(g.Player.Entity)

	// Sync camera position.
	g.Player.Camera.SetPosition(g.SpawnPoint.Add(mcmath.Vec3{Y: player.EyeOffset}))

	// Pop death screen and return to playing.
	g.UI.PopScreen()
	g.State.SetState(GameStatePlaying)
}

func (g *Game) render() {
	imageIndex, cmdBuf, err := g.Renderer.BeginFrame()
	if err != nil {
		return
	}

	if g.State.CurrentState() == GameStatePlaying && g.Player != nil {
		if cam, ok := g.Player.Camera.(*render.Camera); ok {
			g.Renderer.DrawChunks(cmdBuf, cam)
		}
	}

	// Draw UI overlay: crosshair + hotbar
	uiVerts := g.buildUIVertices()
	g.Renderer.DrawUI(cmdBuf, uiVerts)

	if err := g.Renderer.EndFrame(imageIndex); err != nil {
		log.Printf("end frame error: %v", err)
	}

	g.processMeshQueue()
}

func (g *Game) processMeshQueue() {
	if g.meshQueue == nil || g.World == nil || g.Renderer == nil || g.Renderer.ChunkRenderer == nil {
		return
	}
	cr := g.Renderer.ChunkRenderer
	w := g.World
	for i := 0; i < 16; i++ {
		select {
		case cp := <-g.meshQueue:
			c := w.GetChunk(cp)
			if c == nil {
				continue
			}
			neighbors := [4]*chunk.Chunk{
				w.GetChunk(mcmath.ChunkPos{X: cp.X, Z: cp.Z - 1}),
				w.GetChunk(mcmath.ChunkPos{X: cp.X, Z: cp.Z + 1}),
				w.GetChunk(mcmath.ChunkPos{X: cp.X + 1, Z: cp.Z}),
				w.GetChunk(mcmath.ChunkPos{X: cp.X - 1, Z: cp.Z}),
			}
			mesh := chunk.MeshChunk(c, neighbors, block.IsSolid, block.IsTransparent)
			if err := cr.UploadMesh(cp, mesh.Vertices, mesh.Indices); err != nil {
				log.Printf("failed to mesh chunk %v: %v", cp, err)
			}
		default:
			return
		}
	}
}

func (g *Game) StartSingleplayer() {
	storage, err := world.NewStorage(defaultSavePath)
	if err != nil {
		log.Printf("failed to create storage: %v", err)
		return
	}
	g.Storage = storage

	if storage.HasSave() {
		g.loadExistingSave(storage)
	} else {
		g.startNewWorld(storage)
	}

	// Wire up block-change callback to trigger chunk re-meshing.
	if g.World != nil && g.Renderer != nil && g.Renderer.ChunkRenderer != nil {
		cr := g.Renderer.ChunkRenderer
		w := g.World

		meshChunk := func(cp mcmath.ChunkPos) {
			c := w.GetChunk(cp)
			if c == nil {
				return
			}
			neighbors := [4]*chunk.Chunk{
				w.GetChunk(mcmath.ChunkPos{X: cp.X, Z: cp.Z - 1}),
				w.GetChunk(mcmath.ChunkPos{X: cp.X, Z: cp.Z + 1}),
				w.GetChunk(mcmath.ChunkPos{X: cp.X + 1, Z: cp.Z}),
				w.GetChunk(mcmath.ChunkPos{X: cp.X - 1, Z: cp.Z}),
			}
			mesh := chunk.MeshChunk(c, neighbors, block.IsSolid, block.IsTransparent)
		if err := cr.UploadMesh(cp, mesh.Vertices, mesh.Indices); err != nil {
				log.Printf("failed to mesh chunk %v: %v", cp, err)
			}
		}

		g.World.OnBlockChange = meshChunk

		g.meshQueue = make(chan mcmath.ChunkPos, 256)
		g.ChunkLoader.OnChunkLoaded = func(cp mcmath.ChunkPos) {
			select {
			case g.meshQueue <- cp:
			default:
			}
		}

		// Queue all initially loaded chunks for meshing on the render thread.
		for _, cp := range w.LoadedChunkPositions() {
			select {
			case g.meshQueue <- cp:
			default:
			}
		}
	}

	g.State.SetState(GameStatePlaying)
}

// loadExistingSave restores the world and player from an existing save.
func (g *Game) loadExistingSave(storage *world.Storage) {
	levelData, err := storage.LoadLevel()
	if err != nil {
		log.Printf("failed to load level data, starting new world: %v", err)
		g.startNewWorld(storage)
		return
	}

	g.World = world.NewWorld(levelData.Seed)
	g.ChunkLoader = world.NewChunkLoader(g.World)

	if err := g.World.LoadAll(storage); err != nil {
		log.Printf("failed to load chunks: %v", err)
	}

	// Ensure spawn area is loaded.
	spawnChunk := mcmath.BlockPos{X: levelData.SpawnX, Y: 0, Z: levelData.SpawnZ}.ToChunkPos()
	for dx := int32(-2); dx <= 2; dx++ {
		for dz := int32(-2); dz <= 2; dz++ {
			g.World.LoadChunk(mcmath.ChunkPos{X: spawnChunk.X + dx, Z: spawnChunk.Z + dz})
		}
	}

	spawnPos := mcmath.Vec3{
		X: float32(levelData.SpawnX),
		Y: float32(levelData.SpawnY),
		Z: float32(levelData.SpawnZ),
	}
	g.SpawnPoint = spawnPos

	// Try to load player data.
	playerData, playerErr := storage.LoadPlayer()
	if playerErr == nil {
		spawnPos = mcmath.Vec3{X: playerData.X, Y: playerData.Y, Z: playerData.Z}
	}

	cam := render.NewCamera(spawnPos.Add(mcmath.Vec3{Y: player.EyeOffset}))
	cam.FOV = g.Config.Render.FOV

	playerEntity := entity.SpawnPlayer(g.ECSWorld, "Player", spawnPos)
	g.Player = player.NewController(playerEntity, g.ECSWorld, cam, g.KeyMap)
	g.Player.Mode = g.Mode
	g.Player.Inventory = g.Inventory

	g.Spawner = g.newConfiguredSpawner()
	g.applyPlayerConfig(g.Player)

	// Restore player orientation.
	if playerErr == nil {
		if cam, ok := g.Player.Camera.(*render.Camera); ok {
			cam.Yaw = playerData.Yaw
			cam.Pitch = playerData.Pitch
		}
	}
}

// startNewWorld creates a fresh world with a random seed.
func (g *Game) startNewWorld(storage *world.Storage) {
	seed := time.Now().UnixNano()
	g.World = world.NewWorld(seed)
	g.ChunkLoader = world.NewChunkLoader(g.World)

	spawnChunk := mcmath.ChunkPos{X: 0, Z: 0}
	g.World.LoadChunk(spawnChunk)
	for dx := int32(-2); dx <= 2; dx++ {
		for dz := int32(-2); dz <= 2; dz++ {
			g.World.LoadChunk(mcmath.ChunkPos{X: dx, Z: dz})
		}
	}

	spawnY := g.World.GetChunk(spawnChunk).HighestBlock(8, 8) + 2
	spawnPos := mcmath.Vec3{X: 8, Y: float32(spawnY), Z: 8}
	g.SpawnPoint = spawnPos

	cam := render.NewCamera(spawnPos.Add(mcmath.Vec3{Y: player.EyeOffset}))
	cam.FOV = g.Config.Render.FOV

	playerEntity := entity.SpawnPlayer(g.ECSWorld, "Player", spawnPos)
	g.Player = player.NewController(playerEntity, g.ECSWorld, cam, g.KeyMap)
	g.Player.Mode = g.Mode
	g.Player.Inventory = g.Inventory

	g.Spawner = g.newConfiguredSpawner()
	g.applyPlayerConfig(g.Player)

	// Save initial level data.
	levelData := world.LevelData{
		Seed:       seed,
		SpawnX:     8,
		SpawnY:     int32(spawnY),
		SpawnZ:     8,
		GameTime:   0,
		Difficulty: "normal",
	}
	if err := storage.SaveLevel(levelData); err != nil {
		log.Printf("failed to save initial level data: %v", err)
	}
}

// applyPlayerConfig overrides the player controller's default speeds and
// reach with values from the configuration, when they are positive.
func (g *Game) applyPlayerConfig(ctrl *player.Controller) {
	pc := g.Config.Player
	if pc.WalkSpeed > 0 {
		ctrl.WalkSpeed = pc.WalkSpeed
	}
	if pc.SprintSpeed > 0 {
		ctrl.SprintSpeed = pc.SprintSpeed
	}
	if pc.SneakSpeed > 0 {
		ctrl.SneakSpeed = pc.SneakSpeed
	}
	if pc.FlySpeed > 0 {
		ctrl.FlySpeed = pc.FlySpeed
	}
	if pc.JumpVelocity > 0 {
		ctrl.JumpVelocity = pc.JumpVelocity
	}
	if pc.Reach > 0 {
		ctrl.Reach = pc.Reach
	}
}

// newConfiguredSpawner creates a Spawner populated with values from the
// gameplay config.
func (g *Game) newConfiguredSpawner() *entity.Spawner {
	sp := entity.NewSpawner(nil)
	gp := g.Config.Gameplay
	sp.HostileSpawnCap = gp.HostileSpawnCap
	sp.PassiveSpawnCap = gp.PassiveSpawnCap
	sp.SpawnRadius = gp.SpawnRadius
	return sp
}

func (g *Game) Cleanup() {
	g.saveWorldAndPlayer()
	if g.ChunkLoader != nil {
		g.ChunkLoader.Stop()
	}
	if g.Audio != nil {
		g.Audio.Close()
	}
	if g.Renderer != nil {
		g.Renderer.Cleanup()
	}
}

// autoSave persists world and player data during gameplay.
func (g *Game) autoSave() {
	log.Println("auto-saving world...")
	g.saveWorldAndPlayer()
}

// saveWorldAndPlayer writes all world chunks and player state to storage.
func (g *Game) saveWorldAndPlayer() {
	if g.Storage == nil || g.World == nil {
		return
	}

	if err := g.World.SaveAll(g.Storage); err != nil {
		log.Printf("failed to save world: %v", err)
	}

	if g.Player != nil {
		pd := g.buildPlayerData()
		if err := g.Storage.SavePlayer(pd); err != nil {
			log.Printf("failed to save player: %v", err)
		}
	}

	levelData := world.LevelData{
		Seed:       g.World.Seed(),
		GameTime:   g.tickCount,
		Difficulty: "normal",
	}

	// Use player position as spawn if available.
	if g.Player != nil {
		transform := ecs.GetStore[entity.Transform](g.ECSWorld)
		if t, ok := transform.Get(g.Player.Entity); ok {
			levelData.SpawnX = int32(t.Position.X)
			levelData.SpawnY = int32(t.Position.Y)
			levelData.SpawnZ = int32(t.Position.Z)
		}
	}

	if err := g.Storage.SaveLevel(levelData); err != nil {
		log.Printf("failed to save level data: %v", err)
	}
}

// buildPlayerData constructs a PlayerData from the current player state.
func (g *Game) buildPlayerData() world.PlayerData {
	pd := world.PlayerData{
		Health: 20.0,
		Hunger: 20.0,
	}

	transform := ecs.GetStore[entity.Transform](g.ECSWorld)
	if t, ok := transform.Get(g.Player.Entity); ok {
		pd.X = t.Position.X
		pd.Y = t.Position.Y
		pd.Z = t.Position.Z
	}

	if cam, ok := g.Player.Camera.(*render.Camera); ok {
		pd.Yaw = cam.Yaw
		pd.Pitch = cam.Pitch
	}

	health := ecs.GetStore[entity.Health](g.ECSWorld)
	if h, ok := health.Get(g.Player.Entity); ok {
		pd.Health = h.Current
	}

	if g.Inventory != nil {
		slots := make([]world.InventorySlotData, g.Inventory.Size())
		for i := 0; i < g.Inventory.Size(); i++ {
			stack := g.Inventory.GetSlot(i)
			slots[i] = world.InventorySlotData{
				ItemID:     stack.ItemID,
				Count:      stack.Count,
				Durability: stack.Durability,
			}
		}
		pd.InventorySlots = slots
	}

	return pd
}

func (g *Game) buildUIVertices() []render.UIVertex {
	sw := float32(g.Config.Window.Width)
	sh := float32(g.Config.Window.Height)
	cx, cy := sw/2, sh/2

	var verts []render.UIVertex

	// Crosshair: white cross, 2px thick, 20px long
	cross := func(x, y, w, h float32) {
		verts = append(verts,
			render.UIVertex{X: x, Y: y, R: 1, G: 1, B: 1, A: 0.8},
			render.UIVertex{X: x + w, Y: y, R: 1, G: 1, B: 1, A: 0.8},
			render.UIVertex{X: x + w, Y: y + h, R: 1, G: 1, B: 1, A: 0.8},
			render.UIVertex{X: x, Y: y, R: 1, G: 1, B: 1, A: 0.8},
			render.UIVertex{X: x + w, Y: y + h, R: 1, G: 1, B: 1, A: 0.8},
			render.UIVertex{X: x, Y: y + h, R: 1, G: 1, B: 1, A: 0.8},
		)
	}
	cross(cx-10, cy-1, 20, 2)  // horizontal
	cross(cx-1, cy-10, 2, 20)  // vertical

	// Hotbar background: dark semi-transparent bar at bottom center
	barW := float32(9*40 + 8*2) // 9 slots * 40px + 8 gaps * 2px
	barH := float32(44)
	barX := cx - barW/2
	barY := sh - barH - 4

	rect := func(x, y, w, h, r, gg, b, a float32) {
		verts = append(verts,
			render.UIVertex{X: x, Y: y, R: r, G: gg, B: b, A: a},
			render.UIVertex{X: x + w, Y: y, R: r, G: gg, B: b, A: a},
			render.UIVertex{X: x + w, Y: y + h, R: r, G: gg, B: b, A: a},
			render.UIVertex{X: x, Y: y, R: r, G: gg, B: b, A: a},
			render.UIVertex{X: x + w, Y: y + h, R: r, G: gg, B: b, A: a},
			render.UIVertex{X: x, Y: y + h, R: r, G: gg, B: b, A: a},
		)
	}

	// Hotbar background
	rect(barX-2, barY-2, barW+4, barH+4, 0, 0, 0, 0.6)

	// Individual slots
	for i := 0; i < 9; i++ {
		slotX := barX + float32(i)*42
		slotY := barY
		if i == g.Player.SelectedSlot {
			rect(slotX-1, slotY-1, 42, 42, 1, 1, 1, 0.5) // selected highlight
		}
		rect(slotX, slotY, 40, 40, 0.2, 0.2, 0.2, 0.5) // slot background
	}

	// Health bar: red hearts area (top-left of hotbar)
	if g.Player != nil {
		healthStore := ecs.GetStore[entity.Health](g.ECSWorld)
		if h, ok := healthStore.Get(g.Player.Entity); ok {
			healthFrac := float32(h.Current) / float32(h.Max)
			healthW := barW * healthFrac
			rect(barX, barY-14, healthW, 8, 0.8, 0.1, 0.1, 0.8)       // red health
			rect(barX, barY-14, barW, 8, 0.3, 0.05, 0.05, 0.4)        // dark background
		}

		// Hunger bar: above health
		hungerStore := ecs.GetStore[entity.Hunger](g.ECSWorld)
		if hu, ok := hungerStore.Get(g.Player.Entity); ok {
			hungerFrac := float32(hu.FoodLevel) / float32(20)
			hungerW := barW * hungerFrac
			rect(barX, barY-26, hungerW, 8, 0.7, 0.5, 0.1, 0.8)      // orange hunger
			rect(barX, barY-26, barW, 8, 0.25, 0.15, 0.05, 0.4)       // dark background
		}
	}

	return verts
}
