package main

import (
	"math/rand"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func (g *Game) chooseNextNaturalAction() {
	roll := rng.Intn(100)
	nextAction := ActionTypeIdle

	if g.tick >= g.nextSleepAllowedTick && roll < 20 {
		nextAction = ActionTypeSleep
	} else {
		switch {
		case roll < 40:
			nextAction = ActionTypeIdle
		case roll < 55:
			nextAction = ActionTypeWalkingLeft
		case roll < 70:
			nextAction = ActionTypeWalkingRight
		case roll < 85:
			nextAction = ActionTypeRunningLeft
		default:
			nextAction = ActionTypeRunningRight
		}
	}

	// Do not select an outward movement when Daisy is already at an edge.
	maxX := g.screenDimension.Width - g.windowDimension.Width
	if g.windowPos.X <= 0 {
		switch nextAction {
		case ActionTypeWalkingLeft:
			nextAction = ActionTypeWalkingRight
		case ActionTypeRunningLeft:
			nextAction = ActionTypeRunningRight
		}
	} else if g.windowPos.X >= maxX {
		switch nextAction {
		case ActionTypeWalkingRight:
			nextAction = ActionTypeWalkingLeft
		case ActionTypeRunningRight:
			nextAction = ActionTypeRunningLeft
		}
	}

	g.updateCurrentAction(nextAction)
}

func (g *Game) updateCurrentAction(actionType ActionType) {
	action := g.actionIdle

	switch actionType {
	case ActionTypeSleep:
		action = g.actionSleep
	case ActionTypeWalkingLeft:
		action = g.actionWalkingLeft
	case ActionTypeWalkingRight:
		action = g.actionWalkingRight
	case ActionTypeRunningLeft:
		action = g.actionRunningLeft
	case ActionTypeRunningRight:
		action = g.actionRunningRight
	case ActionTypeHang:
		action = g.actionHang
	case ActionTypeLookAtCursor:
		action = g.actionLookAtCursor
	}

	if g.currentAction != nil &&
		g.currentAction.Type == ActionTypeSleep &&
		actionType != ActionTypeSleep {
		g.nextSleepAllowedTick = g.tick + randomLoopTarget(
			g.tuning.SleepCooldownMinTicks,
			g.tuning.SleepCooldownMaxTicks,
		)
	}

	switch actionType {
	case ActionTypeIdle:
		g.idleLoopTarget = randomLoopTarget(g.tuning.IdleLoopMin, g.tuning.IdleLoopMax)
	case ActionTypeSleep:
		g.sleepLoopTarget = randomLoopTarget(g.tuning.SleepLoopMin, g.tuning.SleepLoopMax)
	case ActionTypeRunningLeft, ActionTypeRunningRight:
		g.randomRunLoopTarget = randomLoopTarget(2, 3)
	}

	g.currentAction = action
	g.displayImgTick = 0
	if len(action.Images) > 0 {
		g.displayImage = action.Images[0]
	}
}

func randomLoopTarget(minimum, maximum int) int {
	if maximum <= minimum {
		return minimum
	}
	return minimum + rng.Intn(maximum-minimum+1)
}
