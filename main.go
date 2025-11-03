package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"traffic-sim/sim"
	"traffic-sim/ui"
)

func main() {
	// Create simulation engine
	engine := sim.NewEngine()

	// Start sim engine goroutines
	engine.Start()

	// Create the Ebiten UI and pass a reference to the engine for snapshots
	game := ui.NewGame(engine)

	ebiten.SetWindowSize(ui.ScreenW, ui.ScreenH)
	ebiten.SetWindowTitle("Traffic - Pixel Art (Clean Arch)")

	// Run Ebiten (this will call ui.Update/Draw on main goroutine)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}

	// stop engine after window is closed
	engine.Stop()
}
