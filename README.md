# GoMC

A Minecraft clone written in Go with a Vulkan renderer.

## Features

**World**
- Procedural terrain generation with 8 biomes (Plains, Forest, Desert, Taiga, Jungle, Swamp, Mountains, Ocean)
- Cave carving, ore veins, tree placement
- Nether dimension with netherrack caves, lava oceans, glowstone clusters
- Village and dungeon structure generation
- Day/night cycle, weather (rain, thunder, lightning)
- Fluid simulation (water/lava flow)
- Fire spread to flammable blocks

**Blocks (107 types)**
- Full blocks: stone, dirt, ores, wood variants, wool colors, glass, bricks
- Functional: crafting table, furnace, chest, anvil, brewing stand, enchanting table, jukebox, bed, signs
- Redstone: wire, torch, lever, button with signal propagation (0-15 strength)
- Pistons (push up to 12 blocks, sticky variant)
- Stairs (6 types) and slabs (6 types) with partial-block collision
- Doors and trapdoors with open/close toggle
- Crops: farmland, wheat with 8 growth stages

**Entities (17 types)**
- Player, Zombie, Skeleton, Creeper (explodes), Spider (wall climbing), Enderman (teleport)
- Passive mobs: Cow, Pig, Sheep, Chicken (breeding system)
- Villagers with 5 professions and trading
- Vehicles: Boat (water riding), Minecart (rail riding)
- Projectiles: Arrow, Fishing Bobber
- Items, XP Orbs, Paintings, Item Frames

**Combat & Survival**
- Health, hunger, saturation system
- Melee combat with knockback
- Armor with damage reduction
- Food eating (8+ food types with unique values)
- Drowning, lava damage, fall damage
- Death/respawn with bed spawn points

**Items & Crafting**
- 50+ crafting recipes (shaped grid matching)
- Furnace smelting (7+ recipes)
- Brewing stand (5 potion recipes)
- Tool durability (wood/stone/iron/diamond tiers)
- Enchanting system (6 enchantments: Sharpness, Protection, Efficiency, Unbreaking, Fortune, Silk Touch)
- Anvil: repair, rename, enchantment combining
- Functional items: compass, clock, flint & steel, fishing rod

**Potion Effects**
- 10 status effects: Speed, Slowness, Haste, Strength, Regeneration, Poison, Night Vision, Invisibility, Fire Resistance, Jump Boost

**AI & Pathfinding**
- A* pathfinding for hostile mobs
- AI state machine: Idle, Wander, Chase, Attack, Flee
- Creeper fuse mechanic, Spider wall climbing, Enderman teleportation
- Mob spawning with day/night hostility

**Multiplayer**
- TCP server with packet framing
- Entity synchronization (spawn/move/remove)
- Player position relay
- Chat system with commands

**Commands**
- `/gamemode`, `/tp`, `/give`, `/time`, `/weather`, `/kill`, `/seed`, `/help`

**UI**
- HUD: health hearts, hunger bars, XP bar, hotbar
- Inventory, chest, furnace, crafting table screens
- F3 debug overlay (FPS, coords, biome, memory)
- Chat, pause menu, death screen
- Achievement toast notifications

**Audio**
- Procedural sound generation
- Music playback system
- Jukebox with music discs

**Rendering**
- Vulkan graphics pipeline
- Greedy meshing for chunks
- Entity rendering
- Block break animation, particles

## Architecture

```
cmd/
  client/     Game client (Vulkan window)
  server/     Multiplayer server

internal/
  audio/      Procedural audio, music player
  biome/      Whittaker biome classification
  block/      Block types, properties, drops, redstone
  chunk/      16x256x16 chunks, greedy meshing
  config/     Game configuration with TOML/JSON
  ecs/        Entity Component System
  entity/     Mobs, AI, physics, breeding, enchanting, effects
  game/       Game loop, state, commands, achievements
  input/      Keyboard/mouse input management
  inventory/  Crafting, smelting, brewing, anvil
  item/       Item types, properties, durability
  light/      Light propagation
  mcmath/     Vectors, AABB, raycasting, noise
  network/    TCP server, entity sync
  noise/      Perlin/Simplex noise generators
  particle/   Particle system
  physics/    Collision detection, sweep
  player/     Controller, interaction, swimming, eating, fishing
  protocol/   Network packet definitions
  registry/   Type-safe ID registry
  render/     Vulkan renderer, pipelines, textures
  texgen/     Procedural texture generation
  tick/       Random ticks, scheduled ticks, redstone, crops, pistons
  ui/         Screens, HUD, debug overlay
  world/      World, terrain gen, structures, Nether, signs, storage
```

## Building

### Prerequisites

- Go 1.26+
- Vulkan SDK / MoltenVK (macOS: `brew install molten-vk`)
- GLFW 3.3+ (`brew install glfw`)

### Build & Run

```bash
# Client (requires Vulkan)
go build -o gomc ./cmd/client
./gomc

# Server (multiplayer, no GPU needed)
go build -o gomc-server ./cmd/server
./gomc-server
```

### Configuration

Create `config.json` in the working directory (optional — defaults are used otherwise):

```json
{
  "window": {"width": 1280, "height": 720, "title": "GoMC"},
  "audio": {"masterVolume": 0.8},
  "gameplay": {"tickRate": 20, "autoSaveIntervalTicks": 6000}
}
```

### Tests

```bash
# Run all tests (no GPU required)
go test ./internal/... -count=1

# With race detector
go test -race ./internal/... -count=1

# Specific package
go test ./internal/entity/ -v -cover
```

## Stats

| Metric | Value |
|--------|-------|
| Lines of Go | 42,178 |
| Test functions | 1,270 |
| Block types | 107 |
| Entity types | 17 |
| Packages | 25 |
| Commits | 80 |

## License

MIT
