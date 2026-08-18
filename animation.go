package main


func (g *Game) updateDisplayImage(cursorPos Point) {
	g.displayImgTick++

	frameCount := len(g.currentAction.Images)
	if frameCount == 0 {
		g.displayImage = nil
		return
	}

	frameTicks := g.tuning.FrameTicks
	totalFramesElapsed := g.displayImgTick / frameTicks
	imgIdx := totalFramesElapsed % frameCount
	animLoopCount := totalFramesElapsed / frameCount
	sleepLoopCount := 0

	switch g.currentAction.Type {
	case ActionTypeSleep:
		if frameCount > 1 && totalFramesElapsed >= frameCount {
			loopFrameCount := frameCount - 1
			framesAfterFirstPass := totalFramesElapsed - frameCount
			imgIdx = 1 + framesAfterFirstPass%loopFrameCount
			sleepLoopCount = framesAfterFirstPass / loopFrameCount
		}

	case ActionTypeHang:
		maxIdx := frameCount - 1
		if totalFramesElapsed >= maxIdx {
			imgIdx = maxIdx
		}

	case ActionTypeLookAtCursor:
		imgIdx = g.getLookAtCursorFrameIndex(cursorPos)
		if imgIdx >= frameCount {
			imgIdx = frameCount - 1
		}
	}

	g.displayImage = g.currentAction.Images[imgIdx]

	if g.displayImgTick%frameTicks == 0 {
		switch g.currentAction.Type {
		case ActionTypeWalkingLeft:
			if !g.moveWindowHorizontally(-g.tuning.WalkSpeed) {
				g.updateCurrentAction(ActionTypeIdle)
				return
			}

		case ActionTypeWalkingRight:
			if !g.moveWindowHorizontally(g.tuning.WalkSpeed) {
				g.updateCurrentAction(ActionTypeIdle)
				return
			}

		case ActionTypeRunningLeft:
			if !g.moveWindowHorizontally(-g.tuning.RunSpeed) {
				g.updateCurrentAction(ActionTypeIdle)
				return
			}

		case ActionTypeRunningRight:
			if !g.moveWindowHorizontally(g.tuning.RunSpeed) {
				g.updateCurrentAction(ActionTypeIdle)
				return
			}
		}
	}

	switch g.currentAction.Type {
	case ActionTypeIdle:
		if imgIdx == 0 && animLoopCount >= g.idleLoopTarget {
			g.chooseNextNaturalAction()
			return
		}

	case ActionTypeSleep:
		if frameCount == 1 {
			if animLoopCount >= g.sleepLoopTarget {
				g.updateCurrentAction(ActionTypeIdle)
				return
			}
		} else if sleepLoopCount >= g.sleepLoopTarget {
			g.updateCurrentAction(ActionTypeIdle)
			return
		}

	case ActionTypeWalkingLeft, ActionTypeWalkingRight:
		if imgIdx == 0 && animLoopCount > 2 {
			g.updateCurrentAction(ActionTypeIdle)
			return
		}

	case ActionTypeRunningLeft, ActionTypeRunningRight:
		if imgIdx == 0 && animLoopCount >= g.randomRunLoopTarget {
			g.updateCurrentAction(ActionTypeIdle)
			return
		}
	}
}
