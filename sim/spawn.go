package sim

import (
	"math/rand"
	"time"
)

func spawnLoop(e *Engine) {
	t := time.NewTicker(900 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-e.ctx.Done():
			return
		case d := <-e.spawnCh:
			createCar(e, d)
		case <-t.C:
			createCar(e, Dir(rand.Intn(4)))
		}
	}
}

func createCar(e *Engine, d Dir) {
	e.mu.Lock()
	defer e.mu.Unlock()
	id := e.nextID
	e.nextID++
	qidx := len(e.queues[d])
	c := &Car{ID: id, Dir: d, Color: rand.Uint32(), Waiting: true, QIdx: qidx}
	switch d {
	case North:
		c.X = cx - laneOffset
		c.Y = -80
		c.Tx, c.Ty = queuePosFor(North, qidx)
	case South:
		c.X = cx + laneOffset
		c.Y = screenH + 80
		c.Tx, c.Ty = queuePosFor(South, qidx)
	case West:
		c.X = -80
		c.Y = cy - laneOffset
		c.Tx, c.Ty = queuePosFor(West, qidx)
	case East:
		c.X = screenW + 80
		c.Y = cy + laneOffset
		c.Tx, c.Ty = queuePosFor(East, qidx)
	}
	e.cars = append(e.cars, c)
	e.queues[d] = append(e.queues[d], c)
}
