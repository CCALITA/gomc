package ui

import (
	"testing"

	"github.com/fanxiyao/gomc/internal/input"
	"github.com/fanxiyao/gomc/internal/inventory"
	"github.com/fanxiyao/gomc/internal/item"
	"github.com/stretchr/testify/assert"
)

// ---------- UIManager tests ----------

func TestUIManager_NewEmpty(t *testing.T) {
	m := NewUIManager()
	assert.NotNil(t, m)
	assert.Equal(t, 0, m.ScreenCount())
	assert.Nil(t, m.CurrentScreen())
	assert.False(t, m.IsBlockingInput())
}

func TestUIManager_PushPopScreen(t *testing.T) {
	m := NewUIManager()
	hud := NewHUD(inventory.NewInventory(9))
	pause := NewPauseMenu(nil, nil)

	m.PushScreen(hud)
	assert.Equal(t, 1, m.ScreenCount())
	assert.Equal(t, hud, m.CurrentScreen())

	m.PushScreen(pause)
	assert.Equal(t, 2, m.ScreenCount())
	assert.Equal(t, pause, m.CurrentScreen())

	popped := m.PopScreen()
	assert.Equal(t, pause, popped)
	assert.Equal(t, 1, m.ScreenCount())
	assert.Equal(t, hud, m.CurrentScreen())
}

func TestUIManager_PopEmptyStack(t *testing.T) {
	m := NewUIManager()
	assert.Nil(t, m.PopScreen())
}

func TestUIManager_PushNilScreen(t *testing.T) {
	m := NewUIManager()
	m.PushScreen(nil)
	assert.Equal(t, 0, m.ScreenCount())
}

func TestUIManager_IsBlockingInput_HUD(t *testing.T) {
	m := NewUIManager()
	hud := NewHUD(inventory.NewInventory(9))
	m.PushScreen(hud)
	assert.False(t, m.IsBlockingInput(), "HUD should not block input")
}

func TestUIManager_IsBlockingInput_PauseMenu(t *testing.T) {
	m := NewUIManager()
	hud := NewHUD(inventory.NewInventory(9))
	pause := NewPauseMenu(nil, nil)
	m.PushScreen(hud)
	m.PushScreen(pause)
	assert.True(t, m.IsBlockingInput(), "Pause menu should block input")
}

func TestUIManager_IsBlockingInput_InventoryScreen(t *testing.T) {
	m := NewUIManager()
	hud := NewHUD(inventory.NewInventory(9))
	inv := NewInventoryScreen(inventory.NewInventory(36), &inventory.CraftingGrid{}, nil)
	m.PushScreen(hud)
	m.PushScreen(inv)
	assert.True(t, m.IsBlockingInput(), "Inventory screen should block input")
}

func TestUIManager_Update_DelegatesToCurrentScreen(t *testing.T) {
	m := NewUIManager()
	hud := NewHUD(inventory.NewInventory(9))
	hud.Health = 20
	m.PushScreen(hud)

	inp := input.NewManager()
	// Simulate pressing F3 (toggle FPS).
	inp.KeyCallback(input.KeyF3, 0, input.ActionPress, 0)
	m.Update(inp, 0.016)
	assert.True(t, hud.ShowFPS)
}

func TestUIManager_Update_EmptyStack(t *testing.T) {
	m := NewUIManager()
	inp := input.NewManager()
	// Should not panic.
	m.Update(inp, 0.016)
}

func TestUIManager_Draw_OverlayChain(t *testing.T) {
	m := NewUIManager()
	r := NewUIRenderer(800, 600)

	hud := NewHUD(inventory.NewInventory(9))
	pause := NewPauseMenu(nil, nil)

	m.PushScreen(hud)
	m.PushScreen(pause)

	m.Draw(r)
	// Both HUD and pause menu should have drawn commands.
	assert.Greater(t, r.CommandCount(), 0)
}

func TestUIManager_Draw_NonOverlayHidesBelow(t *testing.T) {
	m := NewUIManager()
	r := NewUIRenderer(800, 600)

	hud := NewHUD(inventory.NewInventory(9))
	mainMenu := NewMainMenu(nil)

	m.PushScreen(hud)
	m.PushScreen(mainMenu)

	m.Draw(r)
	count := r.CommandCount()
	assert.Greater(t, count, 0)

	// Compare: draw only main menu (non-overlay should replace below).
	r2 := NewUIRenderer(800, 600)
	mainMenu.Draw(r2)
	// Main menu is non-overlay, so it should be the only one drawn.
	assert.Equal(t, r2.CommandCount(), count)
}

// ---------- UIRenderer tests ----------

func TestUIRenderer_New(t *testing.T) {
	r := NewUIRenderer(1920, 1080)
	assert.Equal(t, float32(1920), r.ScreenWidth)
	assert.Equal(t, float32(1080), r.ScreenHeight)
	assert.Equal(t, 0, r.CommandCount())
}

func TestUIRenderer_DrawRect(t *testing.T) {
	r := NewUIRenderer(800, 600)
	r.DrawRect(10, 20, 100, 50, 1, 0, 0, 1)
	assert.Equal(t, 1, r.CommandCount())

	cmd := r.Commands()[0]
	assert.Equal(t, DrawCmdRect, cmd.Type)
	assert.Equal(t, float32(10), cmd.X)
	assert.Equal(t, float32(20), cmd.Y)
	assert.Equal(t, float32(100), cmd.W)
	assert.Equal(t, float32(50), cmd.H)
	assert.Equal(t, float32(1), cmd.R)
	assert.Equal(t, float32(0), cmd.G)
	assert.Equal(t, float32(0), cmd.B)
	assert.Equal(t, float32(1), cmd.A)
}

func TestUIRenderer_DrawTexturedRect(t *testing.T) {
	r := NewUIRenderer(800, 600)
	r.DrawTexturedRect(0, 0, 64, 64, 0.25, 0.5, 0.125, 0.125)
	assert.Equal(t, 1, r.CommandCount())

	cmd := r.Commands()[0]
	assert.Equal(t, DrawCmdTexturedRect, cmd.Type)
	assert.Equal(t, float32(0.25), cmd.U)
	assert.Equal(t, float32(0.5), cmd.V)
}

func TestUIRenderer_DrawText(t *testing.T) {
	r := NewUIRenderer(800, 600)
	r.DrawText(10, 10, "Hello", 1.5, 1, 1, 1)
	assert.Equal(t, 1, r.CommandCount())

	cmd := r.Commands()[0]
	assert.Equal(t, DrawCmdText, cmd.Type)
	assert.Equal(t, "Hello", cmd.Text)
	assert.Equal(t, float32(1.5), cmd.Scale)
	assert.Equal(t, float32(1.0), cmd.A) // alpha defaults to 1.0 for text
}

func TestUIRenderer_DrawItemSlot(t *testing.T) {
	r := NewUIRenderer(800, 600)
	stack := item.ItemStack{ItemID: item.Stone, Count: 32}
	r.DrawItemSlot(100, 200, stack)
	assert.Equal(t, 1, r.CommandCount())

	cmd := r.Commands()[0]
	assert.Equal(t, DrawCmdItemSlot, cmd.Type)
	assert.Equal(t, item.Stone, cmd.Stack.ItemID)
	assert.Equal(t, 32, cmd.Stack.Count)
}

func TestUIRenderer_Clear(t *testing.T) {
	r := NewUIRenderer(800, 600)
	r.DrawRect(0, 0, 10, 10, 1, 1, 1, 1)
	r.DrawText(0, 0, "test", 1, 1, 1, 1)
	assert.Equal(t, 2, r.CommandCount())

	r.Clear()
	assert.Equal(t, 0, r.CommandCount())
}

func TestUIRenderer_String(t *testing.T) {
	r := NewUIRenderer(800, 600)
	r.DrawRect(0, 0, 1, 1, 1, 1, 1, 1)
	s := r.String()
	assert.Contains(t, s, "800x600")
	assert.Contains(t, s, "1 cmds")
}

func TestUIRenderer_MultipleCommands(t *testing.T) {
	r := NewUIRenderer(800, 600)
	for i := 0; i < 100; i++ {
		r.DrawRect(0, 0, 1, 1, 1, 1, 1, 1)
	}
	assert.Equal(t, 100, r.CommandCount())
}

// ---------- HUD tests ----------

func TestHUD_NewDefaults(t *testing.T) {
	hotbar := inventory.NewInventory(9)
	hud := NewHUD(hotbar)
	assert.Equal(t, hotbar, hud.Hotbar)
	assert.Equal(t, 0, hud.SelectedSlot)
	assert.Equal(t, 20, hud.Health)
	assert.Equal(t, 20, hud.MaxHealth)
	assert.Equal(t, 20, hud.Hunger)
	assert.Equal(t, 20, hud.MaxHunger)
	assert.False(t, hud.ShowFPS)
}

func TestHUD_IsOverlay(t *testing.T) {
	hud := NewHUD(inventory.NewInventory(9))
	assert.False(t, hud.IsOverlay())
}

func TestHUD_HotbarSelection(t *testing.T) {
	hud := NewHUD(inventory.NewInventory(9))
	inp := input.NewManager()

	// Select slot 3 (key '4').
	inp.KeyCallback(input.Key4, 0, input.ActionPress, 0)
	hud.Update(inp, 0.016)
	assert.Equal(t, 3, hud.SelectedSlot)

	// Advance frame, then select slot 8 (key '9').
	inp.Update()
	inp.KeyCallback(input.Key9, 0, input.ActionPress, 0)
	hud.Update(inp, 0.016)
	assert.Equal(t, 8, hud.SelectedSlot)
}

func TestHUD_FPSToggle(t *testing.T) {
	hud := NewHUD(inventory.NewInventory(9))
	inp := input.NewManager()

	inp.KeyCallback(input.KeyF3, 0, input.ActionPress, 0)
	hud.Update(inp, 0.016)
	assert.True(t, hud.ShowFPS)

	// Advance frame, toggle again.
	inp.Update()
	inp.KeyCallback(input.KeyF3, 0, input.ActionRelease, 0)
	inp.Update()
	inp.KeyCallback(input.KeyF3, 0, input.ActionPress, 0)
	hud.Update(inp, 0.016)
	assert.False(t, hud.ShowFPS)
}

func TestHUD_DrawCrosshair(t *testing.T) {
	hud := NewHUD(inventory.NewInventory(9))
	r := NewUIRenderer(800, 600)
	hud.Draw(r)

	// At minimum: 2 crosshair rects + 9 hotbar slots + health + hunger.
	assert.Greater(t, r.CommandCount(), 10)
}

func TestHUD_DrawWithFPS(t *testing.T) {
	hud := NewHUD(inventory.NewInventory(9))
	hud.ShowFPS = true
	hud.FPS = 60

	r := NewUIRenderer(800, 600)
	hud.Draw(r)

	// Check that at least one text command contains FPS.
	foundFPS := false
	for _, cmd := range r.Commands() {
		if cmd.Type == DrawCmdText && cmd.Text == "FPS: 60" {
			foundFPS = true
			break
		}
	}
	assert.True(t, foundFPS, "FPS text should be drawn when ShowFPS is true")
}

func TestHUD_DrawHotbarWithItems(t *testing.T) {
	hotbar := inventory.NewInventory(9)
	hotbar.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 64})
	hotbar.SetSlot(4, item.ItemStack{ItemID: item.OakLog, Count: 16})

	hud := NewHUD(hotbar)
	r := NewUIRenderer(800, 600)
	hud.Draw(r)

	// Count item slot commands.
	itemSlotCount := 0
	for _, cmd := range r.Commands() {
		if cmd.Type == DrawCmdItemSlot {
			itemSlotCount++
		}
	}
	assert.Equal(t, 2, itemSlotCount, "Should have 2 item slot draw commands")
}

func TestHUD_GetSelectedItem(t *testing.T) {
	hotbar := inventory.NewInventory(9)
	hotbar.SetSlot(2, item.ItemStack{ItemID: item.DiamondPickaxe, Count: 1})

	hud := NewHUD(hotbar)
	hud.SelectedSlot = 2

	selected := hud.GetSelectedItem()
	assert.Equal(t, item.DiamondPickaxe, selected.ItemID)
	assert.Equal(t, 1, selected.Count)
}

func TestHUD_GetSelectedItem_EmptySlot(t *testing.T) {
	hotbar := inventory.NewInventory(9)
	hud := NewHUD(hotbar)
	selected := hud.GetSelectedItem()
	assert.True(t, selected.IsEmpty())
}

func TestHUD_GetSelectedItem_NilHotbar(t *testing.T) {
	hud := NewHUD(nil)
	selected := hud.GetSelectedItem()
	assert.True(t, selected.IsEmpty())
}

func TestHUD_HandleKey_NoOp(t *testing.T) {
	hud := NewHUD(inventory.NewInventory(9))
	// Should not panic.
	hud.HandleKey(input.KeyEscape)
}

func TestHUD_HealthBar_HalfHeart(t *testing.T) {
	hud := NewHUD(inventory.NewInventory(9))
	hud.Health = 7 // 3 full hearts + 1 half

	r := NewUIRenderer(800, 600)
	hud.Draw(r)
	assert.Greater(t, r.CommandCount(), 0)
}

func TestHUD_HungerBar_HalfDrumstick(t *testing.T) {
	hud := NewHUD(inventory.NewInventory(9))
	hud.Hunger = 13 // 6 full + 1 half

	r := NewUIRenderer(800, 600)
	hud.Draw(r)
	assert.Greater(t, r.CommandCount(), 0)
}

// ---------- InventoryScreen tests ----------

func TestInventoryScreen_New(t *testing.T) {
	inv := inventory.NewInventory(36)
	grid := &inventory.CraftingGrid{}
	closed := false
	screen := NewInventoryScreen(inv, grid, func() { closed = true })

	assert.NotNil(t, screen)
	assert.False(t, screen.IsClosed())
	assert.True(t, screen.HeldItem.IsEmpty())
	assert.True(t, screen.IsOverlay())
	assert.False(t, closed)
}

func TestInventoryScreen_CloseOnEscape(t *testing.T) {
	inv := inventory.NewInventory(36)
	grid := &inventory.CraftingGrid{}
	closed := false
	screen := NewInventoryScreen(inv, grid, func() { closed = true })

	inp := input.NewManager()
	inp.KeyCallback(input.KeyEscape, 0, input.ActionPress, 0)
	screen.Update(inp, 0.016)

	assert.True(t, screen.IsClosed())
	assert.True(t, closed)
}

func TestInventoryScreen_CloseOnE(t *testing.T) {
	inv := inventory.NewInventory(36)
	grid := &inventory.CraftingGrid{}
	closed := false
	screen := NewInventoryScreen(inv, grid, func() { closed = true })

	inp := input.NewManager()
	inp.KeyCallback(input.KeyE, 0, input.ActionPress, 0)
	screen.Update(inp, 0.016)

	assert.True(t, screen.IsClosed())
	assert.True(t, closed)
}

func TestInventoryScreen_CloseReturnsHeldItem(t *testing.T) {
	inv := inventory.NewInventory(36)
	grid := &inventory.CraftingGrid{}
	screen := NewInventoryScreen(inv, grid, nil)

	screen.HeldItem = item.ItemStack{ItemID: item.Stone, Count: 10}
	screen.Close()

	assert.True(t, screen.HeldItem.IsEmpty(), "Held item should be returned to inventory")
	slot := inv.GetSlot(0)
	assert.Equal(t, item.Stone, slot.ItemID)
	assert.Equal(t, 10, slot.Count)
}

func TestInventoryScreen_CloseIdempotent(t *testing.T) {
	inv := inventory.NewInventory(36)
	grid := &inventory.CraftingGrid{}
	callCount := 0
	screen := NewInventoryScreen(inv, grid, func() { callCount++ })

	screen.Close()
	screen.Close()
	assert.Equal(t, 1, callCount, "onClose should only be called once")
}

func TestInventoryScreen_HandleKey(t *testing.T) {
	inv := inventory.NewInventory(36)
	grid := &inventory.CraftingGrid{}
	closed := false
	screen := NewInventoryScreen(inv, grid, func() { closed = true })

	screen.HandleKey(input.KeyE)
	assert.True(t, closed)
}

func TestInventoryScreen_SwapWithSlot_PickUp(t *testing.T) {
	inv := inventory.NewInventory(36)
	inv.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 32})
	grid := &inventory.CraftingGrid{}
	screen := NewInventoryScreen(inv, grid, nil)

	// Pick up from slot 0.
	screen.swapWithSlot(0)
	assert.Equal(t, item.Stone, screen.HeldItem.ItemID)
	assert.Equal(t, 32, screen.HeldItem.Count)
	assert.True(t, inv.GetSlot(0).IsEmpty())
}

func TestInventoryScreen_SwapWithSlot_PlaceDown(t *testing.T) {
	inv := inventory.NewInventory(36)
	grid := &inventory.CraftingGrid{}
	screen := NewInventoryScreen(inv, grid, nil)

	screen.HeldItem = item.ItemStack{ItemID: item.Dirt, Count: 16}
	screen.swapWithSlot(5)

	assert.True(t, screen.HeldItem.IsEmpty())
	slot := inv.GetSlot(5)
	assert.Equal(t, item.Dirt, slot.ItemID)
	assert.Equal(t, 16, slot.Count)
}

func TestInventoryScreen_SwapWithSlot_Swap(t *testing.T) {
	inv := inventory.NewInventory(36)
	inv.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 10})
	grid := &inventory.CraftingGrid{}
	screen := NewInventoryScreen(inv, grid, nil)

	screen.HeldItem = item.ItemStack{ItemID: item.Dirt, Count: 5}
	screen.swapWithSlot(0)

	assert.Equal(t, item.Stone, screen.HeldItem.ItemID)
	assert.Equal(t, 10, screen.HeldItem.Count)
	slot := inv.GetSlot(0)
	assert.Equal(t, item.Dirt, slot.ItemID)
	assert.Equal(t, 5, slot.Count)
}

func TestInventoryScreen_SwapWithSlot_NilInventory(t *testing.T) {
	screen := NewInventoryScreen(nil, nil, nil)
	screen.HeldItem = item.ItemStack{ItemID: item.Stone, Count: 1}
	// Should not panic.
	screen.swapWithSlot(0)
	assert.Equal(t, 1, screen.HeldItem.Count)
}

func TestInventoryScreen_SwapWithCraftSlot(t *testing.T) {
	inv := inventory.NewInventory(36)
	grid := &inventory.CraftingGrid{}
	screen := NewInventoryScreen(inv, grid, nil)

	screen.HeldItem = item.ItemStack{ItemID: item.OakPlanks, Count: 4}
	screen.swapWithCraftSlot(0, 0)

	assert.True(t, screen.HeldItem.IsEmpty())
	placed := grid.GetSlot(0, 0)
	assert.Equal(t, item.OakPlanks, placed.ItemID)
	assert.Equal(t, 4, placed.Count)
}

func TestInventoryScreen_SwapWithCraftSlot_PickUp(t *testing.T) {
	inv := inventory.NewInventory(36)
	grid := &inventory.CraftingGrid{}
	grid.SetSlot(1, 1, item.ItemStack{ItemID: item.Stick, Count: 2})
	screen := NewInventoryScreen(inv, grid, nil)

	screen.swapWithCraftSlot(1, 1)
	assert.Equal(t, item.Stick, screen.HeldItem.ItemID)
	assert.True(t, grid.GetSlot(1, 1).IsEmpty())
}

func TestInventoryScreen_SwapWithCraftSlot_NilGrid(t *testing.T) {
	screen := NewInventoryScreen(inventory.NewInventory(36), nil, nil)
	screen.HeldItem = item.ItemStack{ItemID: item.Stone, Count: 1}
	// Should not panic.
	screen.swapWithCraftSlot(0, 0)
}

func TestInventoryScreen_TakeCraftResult_EmptyHands(t *testing.T) {
	inv := inventory.NewInventory(36)
	grid := &inventory.CraftingGrid{}
	screen := NewInventoryScreen(inv, grid, nil)

	// No recipe set, so Craft() returns false.
	screen.takeCraftResult()
	assert.True(t, screen.HeldItem.IsEmpty())
}

func TestInventoryScreen_TakeCraftResult_HandsFull(t *testing.T) {
	inv := inventory.NewInventory(36)
	grid := &inventory.CraftingGrid{}
	screen := NewInventoryScreen(inv, grid, nil)

	screen.HeldItem = item.ItemStack{ItemID: item.Stone, Count: 1}
	screen.takeCraftResult()
	// Should not take result when holding something.
	assert.Equal(t, item.Stone, screen.HeldItem.ItemID)
}

func TestInventoryScreen_TakeCraftResult_NilGrid(t *testing.T) {
	screen := NewInventoryScreen(inventory.NewInventory(36), nil, nil)
	// Should not panic.
	screen.takeCraftResult()
}

func TestInventoryScreen_Draw(t *testing.T) {
	inv := inventory.NewInventory(36)
	inv.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 64})
	grid := &inventory.CraftingGrid{}
	screen := NewInventoryScreen(inv, grid, nil)

	r := NewUIRenderer(800, 600)
	screen.Draw(r)
	assert.Greater(t, r.CommandCount(), 0)
}

func TestInventoryScreen_SetScreenSize(t *testing.T) {
	screen := NewInventoryScreen(inventory.NewInventory(36), &inventory.CraftingGrid{}, nil)
	screen.SetScreenSize(1920, 1080)
	assert.Equal(t, float32(1920), screen.screenWidth)
	assert.Equal(t, float32(1080), screen.screenHeight)
}

// ---------- PauseMenu tests ----------

func TestPauseMenu_New(t *testing.T) {
	menu := NewPauseMenu(nil, nil)
	assert.NotNil(t, menu)
	assert.Equal(t, 3, len(menu.Buttons))
	assert.Equal(t, "Resume", menu.Buttons[0].Label)
	assert.Equal(t, "Options", menu.Buttons[1].Label)
	assert.Equal(t, "Quit", menu.Buttons[2].Label)
	assert.False(t, menu.IsClosed())
	assert.True(t, menu.IsOverlay())
}

func TestPauseMenu_CloseOnEscape(t *testing.T) {
	closed := false
	menu := NewPauseMenu(func() { closed = true }, nil)

	inp := input.NewManager()
	inp.KeyCallback(input.KeyEscape, 0, input.ActionPress, 0)
	menu.Update(inp, 0.016)

	assert.True(t, menu.IsClosed())
	assert.True(t, closed)
}

func TestPauseMenu_CloseIdempotent(t *testing.T) {
	callCount := 0
	menu := NewPauseMenu(func() { callCount++ }, nil)
	menu.Close()
	menu.Close()
	assert.Equal(t, 1, callCount)
}

func TestPauseMenu_HandleKey(t *testing.T) {
	closed := false
	menu := NewPauseMenu(func() { closed = true }, nil)
	menu.HandleKey(input.KeyEscape)
	assert.True(t, closed)
}

func TestPauseMenu_ResumeButton(t *testing.T) {
	closed := false
	menu := NewPauseMenu(func() { closed = true }, nil)
	menu.executeButton(0) // Resume
	assert.True(t, closed)
}

func TestPauseMenu_QuitButton(t *testing.T) {
	quit := false
	menu := NewPauseMenu(nil, func() { quit = true })
	menu.executeButton(2) // Quit
	assert.True(t, quit)
}

func TestPauseMenu_OptionsButton(t *testing.T) {
	menu := NewPauseMenu(nil, nil)
	// Options is a placeholder; should not panic.
	menu.executeButton(1)
	assert.False(t, menu.IsClosed())
}

func TestPauseMenu_InvalidButtonIndex(t *testing.T) {
	menu := NewPauseMenu(nil, nil)
	// Should not panic.
	menu.executeButton(-1)
	menu.executeButton(99)
}

func TestPauseMenu_HoverDetection(t *testing.T) {
	menu := NewPauseMenu(nil, nil)
	menu.SetScreenSize(800, 600)

	// Button layout: centered at x=300 (800-200)/2, width=200
	// The first button should be near the vertical center.
	totalHeight := float32(3)*(pauseBtnHeight+pauseBtnSpacing) - pauseBtnSpacing
	startY := (600 - totalHeight) / 2
	btnX := float32(300)

	// Hit the first button.
	idx := menu.hitTestButton(btnX+50, startY+10)
	assert.Equal(t, 0, idx)

	// Hit the second button.
	idx = menu.hitTestButton(btnX+50, startY+pauseBtnHeight+pauseBtnSpacing+10)
	assert.Equal(t, 1, idx)

	// Miss all buttons.
	idx = menu.hitTestButton(10, 10)
	assert.Equal(t, -1, idx)
}

func TestPauseMenu_Draw(t *testing.T) {
	menu := NewPauseMenu(nil, nil)
	r := NewUIRenderer(800, 600)
	menu.Draw(r)

	// Overlay bg + title text + 3 buttons (rect + text each) = at least 8.
	assert.GreaterOrEqual(t, r.CommandCount(), 8)
}

func TestPauseMenu_HoveredIndex(t *testing.T) {
	menu := NewPauseMenu(nil, nil)
	assert.Equal(t, -1, menu.HoveredIndex())
}

func TestPauseMenu_SetScreenSize(t *testing.T) {
	menu := NewPauseMenu(nil, nil)
	menu.SetScreenSize(1024, 768)
	assert.Equal(t, float32(1024), menu.screenWidth)
	assert.Equal(t, float32(768), menu.screenHeight)
}

// ---------- MainMenu tests ----------

func TestMainMenu_New(t *testing.T) {
	menu := NewMainMenu(nil)
	assert.NotNil(t, menu)
	assert.Equal(t, "GoMC", menu.Title)
	assert.Equal(t, 4, len(menu.Buttons))
	assert.Equal(t, "Singleplayer", menu.Buttons[0].Label)
	assert.Equal(t, "Multiplayer", menu.Buttons[1].Label)
	assert.Equal(t, "Options", menu.Buttons[2].Label)
	assert.Equal(t, "Quit", menu.Buttons[3].Label)
	assert.False(t, menu.IsOverlay())
}

func TestMainMenu_ButtonAction(t *testing.T) {
	var receivedAction MainMenuAction
	menu := NewMainMenu(func(a MainMenuAction) { receivedAction = a })

	menu.executeButton(0)
	assert.Equal(t, MainMenuSingleplayer, receivedAction)

	menu.executeButton(3)
	assert.Equal(t, MainMenuQuit, receivedAction)
}

func TestMainMenu_InvalidButtonIndex(t *testing.T) {
	menu := NewMainMenu(nil)
	// Should not panic.
	menu.executeButton(-1)
	menu.executeButton(99)
}

func TestMainMenu_HandleKey_NoOp(t *testing.T) {
	menu := NewMainMenu(nil)
	// Should not panic.
	menu.HandleKey(input.KeyEscape)
}

func TestMainMenu_HoverDetection(t *testing.T) {
	menu := NewMainMenu(nil)
	menu.SetScreenSize(800, 600)

	// Miss all buttons with a point in the top-left corner.
	idx := menu.hitTestButton(10, 10)
	assert.Equal(t, -1, idx)
}

func TestMainMenu_Draw(t *testing.T) {
	menu := NewMainMenu(nil)
	r := NewUIRenderer(800, 600)
	menu.Draw(r)

	// Background + title + subtitle + 4 buttons (rect + text each) = at least 11.
	assert.GreaterOrEqual(t, r.CommandCount(), 11)
}

func TestMainMenu_HoveredIndex(t *testing.T) {
	menu := NewMainMenu(nil)
	assert.Equal(t, -1, menu.HoveredIndex())
}

func TestMainMenu_SetScreenSize(t *testing.T) {
	menu := NewMainMenu(nil)
	menu.SetScreenSize(1920, 1080)
	assert.Equal(t, float32(1920), menu.screenWidth)
	assert.Equal(t, float32(1080), menu.screenHeight)
}

func TestMainMenu_NilAction(t *testing.T) {
	menu := NewMainMenu(nil)
	// Should not panic with nil callback.
	menu.executeButton(0)
}

// ---------- Integration tests ----------

func TestIntegration_FullScreenFlow(t *testing.T) {
	m := NewUIManager()
	hotbar := inventory.NewInventory(9)
	inv := inventory.NewInventory(36)
	grid := &inventory.CraftingGrid{}

	// Start with main menu.
	mainMenu := NewMainMenu(nil)
	m.PushScreen(mainMenu)
	assert.True(t, m.IsBlockingInput())

	// Transition to HUD.
	m.PopScreen()
	hud := NewHUD(hotbar)
	m.PushScreen(hud)
	assert.False(t, m.IsBlockingInput())

	// Open inventory.
	invScreen := NewInventoryScreen(inv, grid, func() { m.PopScreen() })
	m.PushScreen(invScreen)
	assert.True(t, m.IsBlockingInput())
	assert.Equal(t, 2, m.ScreenCount())

	// Close inventory with key.
	inp := input.NewManager()
	inp.KeyCallback(input.KeyE, 0, input.ActionPress, 0)
	m.Update(inp, 0.016)
	assert.Equal(t, 1, m.ScreenCount())
	assert.False(t, m.IsBlockingInput())

	// Open pause menu.
	inp.Update()
	pauseMenu := NewPauseMenu(func() { m.PopScreen() }, nil)
	m.PushScreen(pauseMenu)
	assert.True(t, m.IsBlockingInput())

	// Close pause with Escape.
	inp.KeyCallback(input.KeyEscape, 0, input.ActionPress, 0)
	m.Update(inp, 0.016)
	assert.Equal(t, 1, m.ScreenCount())
	assert.False(t, m.IsBlockingInput())
}

func TestIntegration_DrawFullStack(t *testing.T) {
	m := NewUIManager()
	hotbar := inventory.NewInventory(9)
	hotbar.SetSlot(0, item.ItemStack{ItemID: item.Stone, Count: 64})

	hud := NewHUD(hotbar)
	hud.ShowFPS = true
	hud.FPS = 144

	m.PushScreen(hud)

	invScreen := NewInventoryScreen(inventory.NewInventory(36), &inventory.CraftingGrid{}, nil)
	m.PushScreen(invScreen)

	r := NewUIRenderer(800, 600)
	m.Draw(r)

	// Both HUD and inventory should have emitted commands.
	assert.Greater(t, r.CommandCount(), 20)
}
