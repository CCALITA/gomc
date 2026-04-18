package game

import (
	"fmt"
	"log"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"

	"github.com/fanxiyao/gomc/internal/audio"
	"github.com/fanxiyao/gomc/internal/block"
	"github.com/fanxiyao/gomc/internal/config"
	"github.com/fanxiyao/gomc/internal/ecs"
	"github.com/fanxiyao/gomc/internal/entity"
	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/player"
	"github.com/fanxiyao/gomc/internal/render"
	"github.com/fanxiyao/gomc/internal/ui"
	"github.com/fanxiyao/gomc/internal/world"
)

const (
	tickRate     = 20.0
	tickInterval = 1.0 / tickRate
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
	Mode        *ModeManager
	Running     bool
}

func (g *Game) Init(cfg *config.Config) error {
	g.Config = cfg
	g.Running = true
	g.State = NewStateManager()
	g.Mode = NewModeManager()

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

	g.State.OnEnter(StatePlaying, func() {
		g.Renderer.Window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	})
	g.State.OnExit(StatePlaying, func() {
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
	g.Scheduler.Add(&entity.AISystem{})
	g.Scheduler.Add(&entity.LifetimeSystem{})
	g.Scheduler.Add(&entity.DamageSystem{})
	g.Scheduler.Add(&entity.HealthSystem{})
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

		for accumulator >= tickInterval {
			g.tick(tickInterval)
			accumulator -= tickInterval
		}

		g.render()
	}
}

func (g *Game) handleGlobalInput() {
	pauseKey := g.KeyMap.GetKey(input.Pause)
	if g.Input.IsKeyJustPressed(pauseKey) {
		switch g.State.CurrentState() {
		case StatePlaying:
			g.State.SetState(StatePaused)
			g.UI.PushScreen(ui.NewPauseMenu(
				func() {
					g.UI.PopScreen()
					g.State.SetState(StatePlaying)
				},
				func() {
					g.Running = false
				},
			))
		case StatePaused:
			g.UI.PopScreen()
			g.State.SetState(StatePlaying)
		}
	}
}

func (g *Game) tick(dt float64) {
	g.UI.Update(g.Input, dt)

	if g.State.CurrentState() != StatePlaying {
		return
	}

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
	}

	g.Scheduler.Update(g.ECSWorld, dt)
}

func (g *Game) render() {
	imageIndex, cmdBuf, err := g.Renderer.BeginFrame()
	if err != nil {
		return
	}

	if g.State.CurrentState() == StatePlaying && g.Player != nil {
		g.Renderer.DrawChunks(cmdBuf, g.Player.Camera)
	}

	vk.CmdEndRenderPass(cmdBuf)

	if err := g.Renderer.EndFrame(imageIndex); err != nil {
		log.Printf("end frame error: %v", err)
	}
}

func (g *Game) StartSingleplayer() {
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

	cam := render.NewCamera(spawnPos.Add(mcmath.Vec3{Y: player.EyeOffset}))
	cam.FOV = g.Config.Render.FOV

	playerEntity := entity.SpawnPlayer(g.ECSWorld, "Player", spawnPos)
	g.Player = player.NewController(playerEntity, g.ECSWorld, cam, g.KeyMap)
	g.Player.Mode = g.Mode

	g.State.SetState(StatePlaying)
}

func (g *Game) Cleanup() {
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
