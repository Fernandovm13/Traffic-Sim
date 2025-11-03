package sim

import (
	"math"
	"time"
)

func loop(e *Engine) {
	ticker := time.NewTicker(60 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.mu.Lock()
			e.phaseTimer--
			if e.phaseTimer <= 0 {
				e.phaseIdx = (e.phaseIdx + 1) % 4
				switch e.phaseIdx {
				case 0:
					e.semNS.set(SemGreen, e.semNS.GreenDur)
					e.semEW.set(SemRed, e.semNS.GreenDur+e.semNS.YellowDur)
					e.phaseTimer = e.semNS.GreenDur
				case 1:
					e.semNS.set(SemYellow, e.semNS.YellowDur)
					e.semEW.set(SemRed, e.semNS.YellowDur)
					e.phaseTimer = e.semNS.YellowDur
				case 2:
					e.semEW.set(SemGreen, e.semEW.GreenDur)
					e.semNS.set(SemRed, e.semEW.GreenDur+e.semEW.YellowDur)
					e.phaseTimer = e.semEW.GreenDur
				case 3:
					e.semEW.set(SemYellow, e.semEW.YellowDur)
					e.semNS.set(SemRed, e.semEW.YellowDur)
					e.phaseTimer = e.semEW.YellowDur
				}
			}

			snap := Snapshot{
				Cars:  make([]Car, len(e.cars)),
				Light: SnapshotLight{NSState: e.semNS.State(), EWState: e.semEW.State()},
			}
			for i, c := range e.cars {
				snap.Cars[i] = *c
			}
			select {
			case e.snapshotCh <- snap:
			default:
			}

			// construir copia para workers y liberar lock pronto
			carsCopy := make([]Car, len(e.cars))
			for i, c := range e.cars {
				carsCopy[i] = *c
			}
			e.mu.Unlock()

			// fan-out jobs (intentar enviar, no bloquear)
			for _, c := range carsCopy {
				select {
				case e.jobs <- job{car: c}:
				default:
				}
			}

			// fan-in results (consumir lo disponible)
			collected := 0
			max := len(carsCopy)
			for collected < max {
				select {
				case res := <-e.results:
					e.mu.Lock()
					for _, mc := range e.cars {
						if mc.ID == res.id {
							mc.X = res.x
							mc.Y = res.y
							break
						}
					}
					e.mu.Unlock()
					collected++
				default:
					collected = max
				}
			}

			// CONTROL: axis-based granting
			e.mu.Lock()
			tryAllowAxis := func(dirs []Dir, sem *Semaphore) {
				for {
					if e.occupiedCount >= maxOccupy {
						return
					}
					allowedOne := false
					for _, d := range dirs {
						q := e.queues[d]
						if len(q) == 0 {
							continue
						}
						front := q[0]
						sx, sy := queuePosFor(d, 0)
						if math.Hypot(front.X-sx, front.Y-sy) > 6.0 {
							continue
						}
						if front.Waiting && sem.State() == SemGreen && e.occupiedCount < maxOccupy {
							if e.occupiedCount == 0 {
								e.intersectionOwner = d
							}
							e.occupiedCount++
							front.Occupying = true
							front.Waiting = false
							cxp, cyp := crossingPointFor(d)
							front.Tx = cxp
							front.Ty = cyp
							if len(e.queues[d]) > 0 {
								e.queues[d] = e.queues[d][1:]
							}
							for i, rem := range e.queues[d] {
								rem.QIdx = i
								rem.Tx, rem.Ty = queuePosFor(d, i)
							}
							allowedOne = true
						}
					}
					if !allowedOne {
						return
					}
				}
			}

			tryAllowAxis([]Dir{North, South}, e.semNS)
			tryAllowAxis([]Dir{East, West}, e.semEW)

			// cuando coche llega al crossing point, asignar exit
			for _, c := range e.cars {
				if c.Occupying && !c.PassedCross {
					cxp, cyp := crossingPointFor(c.Dir)
					if math.Hypot(c.X-cxp, c.Y-cyp) < 6.0 {
						c.PassedCross = true
						ex, ey := exitTargetFor(c.Dir)
						c.Tx = ex
						c.Ty = ey
					}
				}
			}

			// liberar ocupación cuando salga del área central
			for _, c := range e.cars {
				if c.Occupying {
					switch c.Dir {
					case North:
						if c.Y > cy+crossHalf/2 {
							c.Occupying = false
							if e.occupiedCount > 0 {
								e.occupiedCount--
							}
							if e.occupiedCount == 0 {
								e.intersectionOwner = -1
							}
						}
					case South:
						if c.Y < cy-crossHalf/2 {
							c.Occupying = false
							if e.occupiedCount > 0 {
								e.occupiedCount--
							}
							if e.occupiedCount == 0 {
								e.intersectionOwner = -1
							}
						}
					case West:
						if c.X > cx+crossHalf/2 {
							c.Occupying = false
							if e.occupiedCount > 0 {
								e.occupiedCount--
							}
							if e.occupiedCount == 0 {
								e.intersectionOwner = -1
							}
						}
					case East:
						if c.X < cx-crossHalf/2 {
							c.Occupying = false
							if e.occupiedCount > 0 {
								e.occupiedCount--
							}
							if e.occupiedCount == 0 {
								e.intersectionOwner = -1
							}
						}
					}
				}
			}

			// limpiar
			live := e.cars[:0]
			for _, c := range e.cars {
				if c.X < -600 || c.X > 1600 || c.Y < -600 || c.Y > 1600 {
					continue
				}
				live = append(live, c)
			}
			e.cars = live
			e.mu.Unlock()
		}
	}
}
