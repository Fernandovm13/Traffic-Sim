package sim

import "math"

func queuePosFor(d Dir, idx int) (float64, float64) {
	switch d {
	case North:
		return cx - laneOffset, cy - crossHalf - 12 - float64(idx)*queueGap
	case South:
		return cx + laneOffset, cy + crossHalf + 12 + float64(idx)*queueGap
	case West:
		return cx - crossHalf - 12 - float64(idx)*queueGap, cy - laneOffset
	case East:
		return cx + crossHalf + 12 + float64(idx)*queueGap, cy + laneOffset
	}
	return cx, cy
}

func crossingPointFor(d Dir) (float64, float64) {
	switch d {
	case North:
		return cx - laneOffset, cy + 8
	case South:
		return cx + laneOffset, cy - 8
	case West:
		return cx + 8, cy - laneOffset
	case East:
		return cx - 8, cy + laneOffset
	}
	return cx, cy
}

func exitTargetFor(d Dir) (float64, float64) {
	switch d {
	case North:
		return cx - laneOffset, screenH + 400
	case South:
		return cx + laneOffset, -400
	case West:
		return screenW + 400, cy - laneOffset
	case East:
		return -400, cy + laneOffset
	}
	return cx, cy
}

func hypot(x, y float64) float64 { return math.Hypot(x, y) }
