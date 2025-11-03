package ui

import (
	"fmt"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"traffic-sim/sim"
)

const (
	ScreenW = 900
	ScreenH = 700
)

type Game struct {
	engine   *sim.Engine
	frame    int
	lastSnap sim.Snapshot
}

func NewGame(engine *sim.Engine) *Game {
	return &Game{engine: engine}
}

func (g *Game) Update() error {
	g.frame++
	// leer snapshot si hay una nueva (no bloqueante)
	select {
	case s := <-g.engine.SnapshotChan():
		g.lastSnap = s
	default:
	}
	// spawn on space (non-blocking)
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		select {
		case g.engine.SpawnCh <- sim.North:
		default:
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// background
	screen.Fill(color.RGBA{R: 180, G: 220, B: 200, A: 255})

	// draw map and sems (use lastSnap.Light)
	drawRoads(screen)
	drawCrosswalks(screen)
	drawSemaphores(screen, g.lastSnap.Light)

	// draw cars from snapshot (sorted)
	cars := g.lastSnap.Cars
	sort.SliceStable(cars, func(i, j int) bool { return cars[i].Y < cars[j].Y })
	for i := range cars {
		drawCar(screen, &cars[i])
	}

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Space: spawn | Cars: %d", len(cars)), 8, 8)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) { return ScreenW, ScreenH }
