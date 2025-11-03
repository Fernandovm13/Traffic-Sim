package sim

// Tipos y constantes compartidas para la simulación.

type Dir int

const (
	North Dir = iota
	East
	South
	West
)

const (
	screenW    = 900.0
	screenH    = 700.0
	cx         = screenW / 2.0
	cy         = screenH / 2.0
	laneOffset = 80.0
	crossHalf  = 120.0
	queueGap   = 36.0
	maxOccupy  = 3
)

// SemState representa el estado del semáforo.
type SemState int

const (
	SemGreen SemState = iota
	SemYellow
	SemRed
)

// Semaphore: duración y estado
type Semaphore struct {
	GreenDur  int
	YellowDur int

	state SemState
	timer int
	Name  string
}

func (s *Semaphore) set(state SemState, timer int) {
	s.state = state
	s.timer = timer
}
func (s *Semaphore) State() SemState { return s.state }
