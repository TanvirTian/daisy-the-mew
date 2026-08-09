package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)


type GameConfig struct {
	ActionSourceIdle         *ActionSource
	ActionSourceSleep        *ActionSource
	ActionSourceWalkingLeft  *ActionSource
	ActionSourceWalkingRight *ActionSource
	ActionSourceRunning      *ActionSource
	ActionSourceHang         *ActionSource
	ActionSourceLookAtCursor *ActionSource
	WindowDimension          Dimension
	Tuning                   TuningConfig
}

func NewGame(cfg GameConfig) (*Game, error) {
	actionIdle, err := cfg.ActionSourceIdle.ToAction(ActionTypeIdle)
	if err != nil {
		return nil, fmt.Errorf("unable to load idle action: %w", err)
	}

	actionSleep, err := cfg.ActionSourceSleep.ToAction(ActionTypeSleep)
	if err != nil {
		return nil, fmt.Errorf("unable to load sleep action: %w", err)
	}

	actionWalkingLeft, err := cfg.ActionSourceWalkingLeft.ToAction(ActionTypeWalkingLeft)
	if err != nil {
		return nil, fmt.Errorf("unable to load walking-left action: %w", err)
	}

	actionWalkingRight, err := cfg.ActionSourceWalkingRight.ToAction(ActionTypeWalkingRight)
	if err != nil {
		return nil, fmt.Errorf("unable to load walking-right action: %w", err)
	}

	
	actionRunningRight, err := cfg.ActionSourceRunning.ToAction(ActionTypeRunningRight)
	if err != nil {
		return nil, fmt.Errorf("unable to load running action: %w", err)
	}
	actionRunningLeft := &Action{
		Type:   ActionTypeRunningLeft,
		Images: actionRunningRight.Images,
	}

	actionHang, err := cfg.ActionSourceHang.ToAction(ActionTypeHang)
	if err != nil {
		return nil, fmt.Errorf("unable to load hang action: %w", err)
	}

	actionLookAtCursor, err := cfg.ActionSourceLookAtCursor.ToAction(ActionTypeLookAtCursor)
	if err != nil {
		return nil, fmt.Errorf("unable to load look-at-cursor action: %w", err)
	}

	ebiten.SetWindowDecorated(false)
	ebiten.SetScreenTransparent(true)
	ebiten.SetWindowSize(
		cfg.WindowDimension.Width,
		cfg.WindowDimension.Height,
	)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowTitle("Daisy the Mew")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	ebiten.SetRunnableOnUnfocused(true)

	maxScreenWidth, maxScreenHeight := ebiten.ScreenSizeInFullscreen()
	windowPos := Point{
		X: (maxScreenWidth - cfg.WindowDimension.Width) / 2,
		Y: maxScreenHeight - cfg.WindowDimension.Height - 40,
	}

	return &Game{
		actionIdle:           actionIdle,
		actionSleep:          actionSleep,
		actionWalkingLeft:    actionWalkingLeft,
		actionWalkingRight:   actionWalkingRight,
		actionRunningLeft:    actionRunningLeft,
		actionRunningRight:   actionRunningRight,
		actionHang:           actionHang,
		actionLookAtCursor:   actionLookAtCursor,
		windowPos:            windowPos,
		windowDimension:      cfg.WindowDimension,
		screenDimension:      Dimension{Width: maxScreenWidth, Height: maxScreenHeight},
		currentAction:        actionIdle,
		displayImage:         actionIdle.Images[0],
		meowPlayer:           initAudio(),
		idleLoopTarget:       randomLoopTarget(cfg.Tuning.IdleLoopMin, cfg.Tuning.IdleLoopMax),
		sleepLoopTarget:      randomLoopTarget(cfg.Tuning.SleepLoopMin, cfg.Tuning.SleepLoopMax),
		nextSleepAllowedTick: cfg.Tuning.InitialSleepDelayTicks,
		randomRunLoopTarget:  2,
		tuning:               cfg.Tuning,
		physics:              newFallBounce(windowPos.Y, cfg.Tuning),
	}, nil
}

type Game struct {
	actionIdle          *Action
	actionSleep         *Action
	actionWalkingLeft   *Action
	actionWalkingRight  *Action
	actionRunningLeft   *Action
	actionRunningRight  *Action
	actionHang          *Action
	actionLookAtCursor  *Action
	currentAction       *Action
	displayImage        *ebiten.Image
	displayImgTick      int
	windowPos           Point
	windowDimension     Dimension
	screenDimension     Dimension
	lastLeftClickPos    Point
	lastRightClickTick  int
	leftClickTick       int
	tick                int
	meowPlayer          *audio.Player
	cursorTrackingReady bool
	lastGlobalCursorPos Point
	cursorLastMovedTick int

	tuning               TuningConfig
	idleLoopTarget       int
	sleepLoopTarget      int
	nextSleepAllowedTick int
	randomRunLoopTarget  int

	physics fallBounce
}

func (g *Game) Update() error {
	g.tick++

	if err := g.handleWindowClose(); err != nil {
		return err
	}

	cursorX, cursorY := ebiten.CursorPosition()
	cursorPos := Point{X: cursorX, Y: cursorY}
	globalCursorPos := Point{
		X: g.windowPos.X + cursorPos.X,
		Y: g.windowPos.Y + cursorPos.Y,
	}

	// Track desktop coordinates 
	g.trackGlobalCursor(globalCursorPos)

	g.handleWakeUpKittyIfNecessary()
	g.updateWindowPosOnLeftClick(cursorPos)
	g.updateFallBounce()

	
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.handleCursorProximity(cursorPos)
	}

	g.updateDisplayImage(cursorPos)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebiten.SetWindowPosition(g.windowPos.X, g.windowPos.Y)

	if g.displayImage == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest

	bounds := g.displayImage.Bounds()
	if bounds.Dx() > 0 && bounds.Dy() > 0 {
		const catScale = 0.5

		// Squash-and-stretch on landing
		squash := g.physics.squash()
		scaleX := (float64(g.windowDimension.Width) / float64(bounds.Dx())) * catScale * (1 + squash)
		scaleY := (float64(g.windowDimension.Height) / float64(bounds.Dy())) * catScale * (1 - squash)
		scaledWidth := float64(bounds.Dx()) * scaleX
		scaledHeight := float64(bounds.Dy()) * scaleY
		offsetX := (float64(g.windowDimension.Width) - scaledWidth) / 2
		offsetY := float64(g.windowDimension.Height) - scaledHeight

		if g.currentAction.Type == ActionTypeRunningLeft {
			op.GeoM.Scale(-scaleX, scaleY)
			op.GeoM.Translate(offsetX+scaledWidth, offsetY)
		} else {
			op.GeoM.Scale(scaleX, scaleY)
			op.GeoM.Translate(offsetX, offsetY)
		}
	}

	screen.DrawImage(g.displayImage, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.windowDimension.Width, g.windowDimension.Height
}

func (g *Game) updateFallBounce() {
	floorY := g.screenDimension.Height - g.windowDimension.Height
	newY, active := g.physics.update(floorY)
	if active {
		g.windowPos.Y = newY
	}
}
