package sim

type Car struct {
	ID         int
	Dir        Dir
	X, Y       float64
	Tx, Ty     float64
	Color      uint32
	Waiting    bool
	QIdx       int
	Occupying  bool
	PassedCross bool
}
