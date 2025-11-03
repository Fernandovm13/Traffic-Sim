package sim

import (
	"math"
)

type job struct{ car Car }
type jobResult struct{ id int; x, y float64 }

func worker(e *Engine) {
	for {
		select {
		case <-e.ctx.Done():
			return
		case j := <-e.jobs:
			cx0 := j.car.X
			cy0 := j.car.Y
			dx := j.car.Tx - cx0
			dy := j.car.Ty - cy0
			d := math.Hypot(dx, dy)
			if d > 0.5 {
				step := 5.0
				nx := cx0 + (dx/d)*step
				ny := cy0 + (dy/d)*step
				select {
				case e.results <- jobResult{id: j.car.ID, x: nx, y: ny}:
				case <-e.ctx.Done():
					return
				}
			} else {
				select {
				case e.results <- jobResult{id: j.car.ID, x: j.car.Tx, y: j.car.Ty}:
				case <-e.ctx.Done():
					return
				}
			}
		}
	}
}
