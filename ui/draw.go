package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2"

	"traffic-sim/sim"
)

const (
	roadW = 220.0
	cross = 240.0
)

// keep this equal to sim.laneOffset
var laneOffset = 80.0

func drawRoads(screen *ebiten.Image) {
	cx, cy := float64(ScreenW)/2, float64(ScreenH)/2
	ebitenutil.DrawRect(screen, 0, cy-roadW/2, ScreenW, roadW, color.RGBA{40, 40, 40, 255})
	ebitenutil.DrawRect(screen, cx-roadW/2, 0, roadW, ScreenH, color.RGBA{40, 40, 40, 255})
	for x := -40; x < ScreenW+40; x += 40 {
		ebitenutil.DrawRect(screen, float64(x), cy-2, 24, 4, color.RGBA{230, 230, 230, 230})
	}
	for y := -40; y < ScreenH+40; y += 40 {
		ebitenutil.DrawRect(screen, cx-2, float64(y), 4, 24, color.RGBA{230, 230, 230, 230})
	}
}

func drawCrosswalks(screen *ebiten.Image) {
	cx, cy := float64(ScreenW)/2, float64(ScreenH)/2
	stepW := 12.0
	stepH := 32.0
	for i := -3; i <= 3; i++ {
		ebitenutil.DrawRect(screen, cx+float64(i)*20-stepW/2, cy-cross/2-10, stepW, stepH, color.White)
	}
	for i := -3; i <= 3; i++ {
		ebitenutil.DrawRect(screen, cx+float64(i)*20-stepW/2, cy+cross/2-22, stepW, stepH, color.White)
	}
	for i := -3; i <= 3; i++ {
		ebitenutil.DrawRect(screen, cx-cross/2-10, cy+float64(i)*20-stepW/2, stepH, stepW, color.White)
	}
	for i := -3; i <= 3; i++ {
		ebitenutil.DrawRect(screen, cx+cross/2-22, cy+float64(i)*20-stepW/2, stepH, stepW, color.White)
	}
}

// draw two semaphores placed in different lanes and show yellow visually
func drawSemaphores(screen *ebiten.Image, light sim.SnapshotLight) {
	cx, cy := float64(ScreenW)/2, float64(ScreenH)/2

	// NS semaphore on left lane (vertical road)
	semN_x := cx - laneOffset
	semN_y := cy - cross/2 - 70
	drawSemWithState(screen, semN_x, semN_y, light.NSState)

	// EW semaphore on bottom lane (horizontal road)
	semE_x := cx + cross/2 + 70
	semE_y := cy + laneOffset
	drawSemWithState(screen, semE_x, semE_y, light.EWState)
}

func drawSemWithState(screen *ebiten.Image, x, y float64, state sim.SemState) {
	// post
	ebitenutil.DrawRect(screen, x-6, y-8, 12, 56, color.RGBA{30, 30, 30, 255})
	boxW := 28.0
	boxH := 48.0
	ebitenutil.DrawRect(screen, x-boxW/2, y-8, boxW, boxH, color.RGBA{20, 20, 20, 255})

	lightW := 10.0
	// draw each light with correct color
	// top = red
	switch state {
	case sim.SemGreen:
		ebitenutil.DrawRect(screen, x-6-lightW/2, y+2, lightW, lightW, color.RGBA{80, 80, 80, 200})
		ebitenutil.DrawRect(screen, x-6-lightW/2, y+2+14, lightW, lightW, color.RGBA{100, 100, 100, 200})
		ebitenutil.DrawRect(screen, x-6-lightW/2, y+2+28, lightW, lightW, color.RGBA{50, 180, 60, 255})
	case sim.SemYellow:
		ebitenutil.DrawRect(screen, x-6-lightW/2, y+2, lightW, lightW, color.RGBA{200, 40, 40, 255})
		ebitenutil.DrawRect(screen, x-6-lightW/2, y+2+14, lightW, lightW, color.RGBA{240, 200, 40, 255})
		ebitenutil.DrawRect(screen, x-6-lightW/2, y+2+28, lightW, lightW, color.RGBA{80, 80, 80, 200})
	case sim.SemRed:
		ebitenutil.DrawRect(screen, x-6-lightW/2, y+2, lightW, lightW, color.RGBA{200, 40, 40, 255})
		ebitenutil.DrawRect(screen, x-6-lightW/2, y+2+14, lightW, lightW, color.RGBA{100, 100, 100, 200})
		ebitenutil.DrawRect(screen, x-6-lightW/2, y+2+28, lightW, lightW, color.RGBA{80, 80, 80, 200})
	}
}

func drawCar(screen *ebiten.Image, c *sim.Car) {
	w := 44.0
	h := 20.0
	x := c.X
	y := c.Y
	clr := color.RGBA{R: uint8((c.Color>>16)&0xFF), G: uint8((c.Color>>8)&0xFF), B: uint8(c.Color&0xFF), A: 255}
	sx := x - w/2
	sy := y - h/2
	ebitenutil.DrawRect(screen, sx, sy, w, h, clr)
	ebitenutil.DrawRect(screen, sx+6, sy+h-4, 12, 4, color.Black)
	ebitenutil.DrawRect(screen, sx+w-18, sy+h-4, 12, 4, color.Black)
	ebitenutil.DrawRect(screen, sx+6, sy+4, 8, 3, color.RGBA{255, 255, 255, 80})
	switch c.Dir {
	case sim.North:
		ebitenutil.DrawRect(screen, sx+w/2-4, sy-6, 8, 6, color.White)
	case sim.South:
		ebitenutil.DrawRect(screen, sx+w/2-4, sy+h, 8, 6, color.White)
	case sim.West:
		ebitenutil.DrawRect(screen, sx-6, sy+h/2-4, 6, 8, color.White)
	case sim.East:
		ebitenutil.DrawRect(screen, sx+w, sy+h/2-4, 6, 8, color.White)
	}
}
