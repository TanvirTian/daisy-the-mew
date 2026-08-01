package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"gopkg.in/validator.v2"
)

type GameConfig struct {
	ActionSourceIdle         *ActionSource `validate:"nonnil"`
	ActionSourceSleep        *ActionSource `validate:"nonnil"`
	ActionSourceWalkingLeft  *ActionSource `validate:"nonnil"`
	ActionSourceWalkingRight *ActionSource `validate:"nonnil"`
	ActionSourceRunning      *ActionSource `validate:"nonnil"`
	ActionSourceHang         *ActionSource `validate:"nonnil"`
	ActionSourceLookAtCursor *ActionSource `validate:"nonnil"`
	WindowDimension          Dimension     `validate:"nonzero"`
}

func (c GameConfig) Validate() error {
	return validator.Validate(c)
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

const (
	closeDoubleClickTicks = 30
	holdClickTicks        = 10

	proximityRadius     = 80.0
	proximityHysteresis = 12.0
	maxInactivityTicks  = 180

	initialSleepDelayTicks = 15 * 60
	sleepCooldownMinTicks  = 30 * 60
	sleepCooldownMaxTicks  = 60 * 60

	idleLoopMin  = 10
	idleLoopMax  = 30
	sleepLoopMin = 6
	sleepLoopMax = 10

	frameTicks = 16
)

func NewGame(cfg GameConfig) (*Game, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %+v", cfg)
	}

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

	const sampleRate = 44100
	audioContext := audio.NewContext(sampleRate)
	var meowPlayer *audio.Player

	meowFile, err := os.Open("mew/mew.mp3")
	if err != nil {
		fmt.Printf("[Audio Error] Could not open sound file: %v\n", err)
	} else {
		decodedMeow, err := mp3.DecodeWithSampleRate(sampleRate, meowFile)
		if err != nil {
			fmt.Printf("[Audio Error] Could not decode MP3 file: %v\n", err)
		} else {
			player, err := audioContext.NewPlayer(decodedMeow)
			if err != nil {
				fmt.Printf("[Audio Error] Could not create player: %v\n", err)
			} else {
				meowPlayer = player
			}
		}
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
		meowPlayer:           meowPlayer,
		idleLoopTarget:       randomLoopTarget(idleLoopMin, idleLoopMax),
		sleepLoopTarget:      randomLoopTarget(sleepLoopMin, sleepLoopMax),
		nextSleepAllowedTick: initialSleepDelayTicks,
		randomRunLoopTarget:  2,
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

	idleLoopTarget       int
	sleepLoopTarget      int
	nextSleepAllowedTick int
	randomRunLoopTarget  int
}

func (g *Game) Update() error {
	g.tick++

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		if g.lastRightClickTick != 0 &&
			g.tick-g.lastRightClickTick <= closeDoubleClickTicks {
			return ebiten.Termination
		}
		g.lastRightClickTick = g.tick
	}

	cursorX, cursorY := ebiten.CursorPosition()
	cursorPos := Point{X: cursorX, Y: cursorY}
	globalCursorPos := Point{
		X: g.windowPos.X + cursorPos.X,
		Y: g.windowPos.Y + cursorPos.Y,
	}

	// Track desktop coordinates rather than window-local coordinates. Moving
	// Daisy's window must not be mistaken for real pointer movement.
	g.trackGlobalCursor(globalCursorPos)

	g.handleWakeUpKittyIfNecessary()
	g.updateWindowPosOnLeftClick(cursorPos)

	// Mouse interaction has priority over cursor-looking behavior.
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

		scaleX := (float64(g.windowDimension.Width) / float64(bounds.Dx())) * catScale
		scaleY := (float64(g.windowDimension.Height) / float64(bounds.Dy())) * catScale
		scaledWidth := float64(bounds.Dx()) * scaleX
		scaledHeight := float64(bounds.Dy()) * scaleY
		offsetX := (float64(g.windowDimension.Width) - scaledWidth) / 2
		offsetY := float64(g.windowDimension.Height) - scaledHeight

		if g.currentAction.Type == ActionTypeRunningLeft {
			// run1.png through run6.png face right. Reflect them around the
			// center of the same drawing rectangle when Daisy runs left.
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

func (g *Game) handleCursorProximity(cursorPos Point) {
	
	switch g.currentAction.Type {
	case ActionTypeIdle, ActionTypeLookAtCursor:
	default:
		return
	}

	isMouseActive := g.cursorLastMovedTick != 0 &&
		g.tick-g.cursorLastMovedTick < maxInactivityTicks
	distance := g.distanceFromCatBounds(cursorPos)
	radius := proximityRadius
	if g.currentAction.Type == ActionTypeLookAtCursor {
		radius += proximityHysteresis
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

func (g *Game) updateDisplayImage(cursorPos Point) {
	g.displayImgTick++

	frameCount := len(g.currentAction.Images)
	if frameCount == 0 {
		g.displayImage = nil
		return
	}

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
			if !g.moveWindowHorizontally(-walkSpeed) {
				g.updateCurrentAction(ActionTypeIdle)
				return
			}

		case ActionTypeWalkingRight:
			if !g.moveWindowHorizontally(walkSpeed) {
				g.updateCurrentAction(ActionTypeIdle)
				return
			}

		case ActionTypeRunningLeft:
			if !g.moveWindowHorizontally(-runSpeed) {
				g.updateCurrentAction(ActionTypeIdle)
				return
			}

		case ActionTypeRunningRight:
			if !g.moveWindowHorizontally(runSpeed) {
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

func (g *Game) chooseNextNaturalAction() {
	roll := rng.Intn(100)
	nextAction := ActionTypeIdle


	// a 20% chance. Running receives 30% total: 15% in each direction.
	if g.tick >= g.nextSleepAllowedTick && roll < 20 {
		nextAction = ActionTypeSleep
	} else {
		switch {
		case roll < 40:
			// Remaining idle is still a valid behavior, keeping Daisy from
			// feeling as though she must constantly perform an action.
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
			sleepCooldownMinTicks,
			sleepCooldownMaxTicks,
		)
	}

	switch actionType {
	case ActionTypeIdle:
		g.idleLoopTarget = randomLoopTarget(idleLoopMin, idleLoopMax)
	case ActionTypeSleep:
		g.sleepLoopTarget = randomLoopTarget(sleepLoopMin, sleepLoopMax)
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

func (g *Game) handleWakeUpKittyIfNecessary() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.leftClickTick = g.tick

		
		if g.currentAction.Type == ActionTypeRunningLeft ||
			g.currentAction.Type == ActionTypeRunningRight {
			g.updateCurrentAction(ActionTypeIdle)
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if g.leftClickTick != 0 &&
			g.tick-g.leftClickTick > holdClickTicks &&
			g.currentAction.Type != ActionTypeHang {
			g.updateCurrentAction(ActionTypeHang)
			g.playMeow()
		}
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		g.leftClickTick = 0
		if g.currentAction.Type == ActionTypeHang {
			g.updateCurrentAction(ActionTypeIdle)
		}
	}
}

func (g *Game) playMeow() {
	if g.meowPlayer == nil {
		return
	}

	if err := g.meowPlayer.Rewind(); err != nil {
		fmt.Printf("[Audio Error] Could not rewind meow: %v\n", err)
		return
	}
	g.meowPlayer.Play()
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
