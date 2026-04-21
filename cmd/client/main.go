package main

import (
	"log"
	"runtime"

	"github.com/fanxiyao/gomc/internal/config"
	"github.com/fanxiyao/gomc/internal/game"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("Using default config: %v", err)
		cfg = config.Default()
	}

	g := &game.Game{}
	if err := g.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	defer g.Cleanup()

	g.StartSingleplayer()
	g.Run()
}
