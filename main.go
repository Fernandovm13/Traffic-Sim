package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"traffic-sim/sim"
	"traffic-sim/ui"
)

func main() {
	engine := sim.NewEngine()

	engine.Start()

	game := ui.NewGame(engine)

	ebiten.SetWindowSize(ui.ScreenW, ui.ScreenH)
	ebiten.SetWindowTitle("Traffic - Pixel Art (Clean Arch)")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}

	engine.Stop()
}
