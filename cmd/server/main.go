package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fanxiyao/gomc/internal/config"
	"github.com/fanxiyao/gomc/internal/mcmath"
	"github.com/fanxiyao/gomc/internal/network"
	"github.com/fanxiyao/gomc/internal/protocol"
	"github.com/fanxiyao/gomc/internal/world"
)

func main() {
	cfg, err := config.Load("config.json")
	if err != nil {
		cfg = config.Default()
	}

	log.Printf("Starting GoMC Server on %s:%d", cfg.Server.Address, cfg.Server.Port)

	gameWorld := world.NewWorld(42)

	for dx := int32(-4); dx <= 4; dx++ {
		for dz := int32(-4); dz <= 4; dz++ {
			gameWorld.LoadChunk(mcmath.ChunkPos{X: dx, Z: dz})
		}
	}
	log.Printf("Pre-generated %d chunks", gameWorld.LoadedChunkCount())

	_ = protocol.DefaultRegistry()

	srv := network.NewServer(cfg.Server.Address, cfg.Server.Port)
	srv.OnConnect(func(conn *network.Connection) {
		log.Printf("Client connected: %s", conn.RemoteAddr())
	})
	srv.OnDisconnect(func(conn *network.Connection) {
		log.Printf("Client disconnected: %s", conn.RemoteAddr())
	})
	srv.OnPacket(func(conn *network.Connection, pkt protocol.Packet) {
		log.Printf("Received packet %T from %s", pkt, conn.RemoteAddr())
	})

	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Printf("Server listening on %s:%d", cfg.Server.Address, cfg.Server.Port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	fmt.Printf("\nReceived %v, shutting down...\n", sig)

	if err := srv.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}
	log.Println("Server stopped.")
}
