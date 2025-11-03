package sim

// Car: estructura por valor (para snapshot seguro)
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
