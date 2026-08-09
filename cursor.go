package main

import "math"

// trackGlobalCursor records desktop coordinate pointer movement
func (g *Game) trackGlobalCursor(cursorPos Point) {
	if !g.cursorTrackingReady {
		g.lastGlobalCursorPos = cursorPos
		g.cursorTrackingReady = true
		return
	}

	if cursorPos != g.lastGlobalCursorPos {
		g.lastGlobalCursorPos = cursorPos
		g.cursorLastMovedTick = g.tick
	}
}

// makes Daisy look toward the cursor when it is near.
func (g *Game) handleCursorProximity(cursorPos Point) {
	switch g.currentAction.Type {
	case ActionTypeIdle, ActionTypeLookAtCursor:
	default:
		return
	}

	isMouseActive := g.cursorLastMovedTick != 0 &&
		g.tick-g.cursorLastMovedTick < g.tuning.MaxInactivityTicks
	distance := g.distanceFromCatBounds(cursorPos)
	radius := g.tuning.ProximityRadius
	if g.currentAction.Type == ActionTypeLookAtCursor {
		radius += g.tuning.ProximityHysteresis
	}

	if distance <= radius && isMouseActive {
		if g.currentAction.Type != ActionTypeLookAtCursor {
			g.updateCurrentAction(ActionTypeLookAtCursor)
		}
	} else if g.currentAction.Type == ActionTypeLookAtCursor {
		g.updateCurrentAction(ActionTypeIdle)
	}
}


func (g *Game) distanceFromCatBounds(cursorPos Point) float64 {
	var dx, dy float64

	if cursorPos.X < 0 {
		dx = float64(-cursorPos.X)
	} else if cursorPos.X > g.windowDimension.Width {
		dx = float64(cursorPos.X - g.windowDimension.Width)
	}

	if cursorPos.Y < 0 {
		dy = float64(-cursorPos.Y)
	} else if cursorPos.Y > g.windowDimension.Height {
		dy = float64(cursorPos.Y - g.windowDimension.Height)
	}

	return math.Hypot(dx, dy)
}


func (g *Game) getLookAtCursorFrameIndex(cursorPos Point) int {
	const margin = 10

	belowBody := cursorPos.Y > g.windowDimension.Height-margin
	aboveBody := cursorPos.Y < margin
	leftOfBody := cursorPos.X < margin
	rightOfBody := cursorPos.X > g.windowDimension.Width-margin

	if belowBody && !aboveBody {
		return 4
	}
	if aboveBody {
		return 3
	}
	if leftOfBody {
		return 1
	}
	if rightOfBody {
		return 2
	}
	return 0
}
