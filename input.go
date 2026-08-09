package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)


func (g *Game) handleWindowClose() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		if g.lastRightClickTick != 0 &&
			g.tick-g.lastRightClickTick <= g.tuning.CloseDoubleClickTicks {
			return ebiten.Termination
		}
		g.lastRightClickTick = g.tick
	}

	return nil
}


func (g *Game) handleWakeUpKittyIfNecessary() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.leftClickTick = g.tick

		// Stop a run immediately when the user grabs Daisy
		if g.currentAction.Type == ActionTypeRunningLeft ||
			g.currentAction.Type == ActionTypeRunningRight {
			g.updateCurrentAction(ActionTypeIdle)
		}

		// Grabbing her again cancels any in-progress fall or bounce.
		g.cancelFallBounce()
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if g.leftClickTick != 0 &&
			g.tick-g.leftClickTick > g.tuning.HoldClickTicks &&
			g.currentAction.Type != ActionTypeHang {
			g.updateCurrentAction(ActionTypeHang)
			g.playMeow()
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		g.leftClickTick = 0
		if g.currentAction.Type == ActionTypeHang {
			g.updateCurrentAction(ActionTypeIdle)
			// If she was dropped above the floor, she falls and bounces.
			g.startFall()
		}
	}
}

func (g *Game) updateWindowPosOnLeftClick(cursorPos Point) {
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		return
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.lastLeftClickPos = cursorPos
		return
	}

	newX := g.windowPos.X + (cursorPos.X - g.lastLeftClickPos.X)
	newY := g.windowPos.Y + (cursorPos.Y - g.lastLeftClickPos.Y)
	maxX := g.screenDimension.Width - g.windowDimension.Width
	maxY := g.screenDimension.Height - g.windowDimension.Height

	if newX < 0 {
		newX = 0
	}
	if newX > maxX {
		newX = maxX
	}
	if newY < 0 {
		newY = 0
	}
	if newY > maxY {
		newY = maxY
	}

	g.windowPos.X = newX
	g.windowPos.Y = newY
}


func (g *Game) moveWindowHorizontally(distance int) bool {
	maxX := g.screenDimension.Width - g.windowDimension.Width
	nextX := g.windowPos.X + distance

	if nextX < 0 {
		g.windowPos.X = 0
		return false
	}
	if nextX > maxX {
		g.windowPos.X = maxX
		return false
	}

	g.windowPos.X = nextX
	return true
}


func (g *Game) startFall() {
	floorY := g.screenDimension.Height - g.windowDimension.Height
	g.physics.start(g.windowPos.Y, floorY)
}

func (g *Game) cancelFallBounce() {
	g.physics.cancel(g.windowPos.Y)
}
