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
	ActionSourceHang         *ActionSource `validate:"nonnil"`
	ActionSourceLookAtCursor *ActionSource `validate:"nonnil"`
	WindowDimension          Dimension     `validate:"nonzero"`
}

func (c GameConfig) Validate() error {
	return validator.Validate(c)
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

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
	ebiten.SetWindowTitle("Desktop Cat")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	maxScreenWidth, maxScreenHeight := ebiten.ScreenSizeInFullscreen()
	windowPos := Point{
		X: (maxScreenWidth - cfg.WindowDimension.Width) / 2,
		Y: maxScreenHeight - cfg.WindowDimension.Height - 40,
	}

	const sampleRate = 44100
	audioCtx := audio.NewContext(sampleRate)
	var meowPlayer *audio.Player

	soundPath := "mew/mew.mp3"
	if _, err := os.Stat(soundPath); os.IsNotExist(err) {
		soundPath = "mew/mew.mp3"
	}

	meowFile, err := os.Open(soundPath)
	if err != nil {
		fmt.Printf("[Audio Error] Could not open sound file: %v\n", err)
	} else {
		decodedMeow, err := mp3.DecodeWithSampleRate(sampleRate, meowFile)
		if err != nil {
			fmt.Printf("[Audio Error] Could not decode MP3 file: %v\n", err)
		} else {
			p, err := audioCtx.NewPlayer(decodedMeow)
			if err != nil {
				fmt.Printf("[Audio Error] Could not create player: %v\n", err)
			} else {
				meowPlayer = p
			}
		}
	}

	return &Game{
		actionIdle:         actionIdle,
		actionSleep:        actionSleep,
		actionWalkingLeft:  actionWalkingLeft,
		actionWalkingRight: actionWalkingRight,
		actionHang:         actionHang,
		actionLookAtCursor: actionLookAtCursor,
		windowPos:          windowPos,
		windowDimension:    cfg.WindowDimension,
		screenDimension: Dimension{
			Width:  maxScreenWidth,
			Height: maxScreenHeight,
		},
		currentAction: actionIdle,
		meowPlayer:    meowPlayer,
	}, nil
}

type Game struct {
	actionIdle         *Action
	actionSleep        *Action
	actionWalkingLeft  *Action
	actionWalkingRight *Action
	actionHang         *Action
	actionLookAtCursor *Action

	displayImgTick  int
	windowPos       Point
	windowDimension Dimension
	screenDimension Dimension
	currentAction   *Action

	lastLeftClickPos Point
	displayImage     *ebiten.Image

	lastRightClickTick int
	leftClickTick      int
	tick               int

	lastCursorPos       Point
	cursorLastMovedTick int

	meowPlayer *audio.Player
}

const (
	closeDoubleClickTicks = 30
	holdClickTicks        = 10
	proximityRadius       = 80.0
	proximityHysteresis   = 12.0
	maxInactivityTicks    = 180
)

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
	cursorPos := Point{
		X: cursorX,
		Y: cursorY,
	}

	g.handleCursorProximity(cursorPos)
	g.updateDisplayImage(cursorPos)
	g.handleWakeUpKittyIfNecessary()
	g.updateWindowPosOnLeftClick(cursorPos)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebiten.SetWindowPosition(
		g.windowPos.X,
		g.windowPos.Y,
	)

	if g.displayImage == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	bounds := g.displayImage.Bounds()

	if bounds.Dx() > 0 && bounds.Dy() > 0 {
		const catScale = 0.5

		scaleX := (float64(g.windowDimension.Width) / float64(bounds.Dx())) * catScale
		scaleY := (float64(g.windowDimension.Height) / float64(bounds.Dy())) * catScale
		op.GeoM.Scale(scaleX, scaleY)

		scaledWidth := float64(bounds.Dx()) * scaleX
		scaledHeight := float64(bounds.Dy()) * scaleY
		offsetX := (float64(g.windowDimension.Width) - scaledWidth) / 2
		offsetY := float64(g.windowDimension.Height) - scaledHeight
		op.GeoM.Translate(offsetX, offsetY)
	}

	screen.DrawImage(g.displayImage, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.windowDimension.Width, g.windowDimension.Height
}

func (g *Game) handleCursorProximity(cursorPos Point) {
	// Always track cursor activity so the state is up to date after another
	// action, such as walking or sleeping, finishes.
	if cursorPos != g.lastCursorPos {
		g.lastCursorPos = cursorPos
		g.cursorLastMovedTick = g.tick
	}

	// Cursor proximity may interrupt idle, but it must not interrupt walking,
	// sleeping, or hanging. Previously, this function immediately replaced a
	// newly selected walking action with LookAtCursor.
	switch g.currentAction.Type {
	case ActionTypeIdle, ActionTypeLookAtCursor:
		// These actions may transition to or from LookAtCursor.
	default:
		return
	}

	dist := g.distanceFromCatBounds(cursorPos)
	radius := proximityRadius

	if g.currentAction.Type == ActionTypeLookAtCursor {
		radius += proximityHysteresis
	}

	isNear := dist <= radius
	isMouseActive := g.cursorLastMovedTick != 0 &&
		g.tick-g.cursorLastMovedTick < maxInactivityTicks

	if isNear && isMouseActive {
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

	const frameTicks = 16

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
		// Play every sleep frame during the first pass. After that, loop only
		// frames 1..N-1 so frame 0 acts as the one-time falling-asleep frame.
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
	}

	g.displayImage = g.currentAction.Images[imgIdx]

	// Move on animation-frame ticks instead of checking whether the image index
	// changed. This also lets a one-frame walking action move correctly.
	if g.displayImgTick%frameTicks == 0 {
		switch g.currentAction.Type {
		case ActionTypeWalkingLeft:
			nextX := g.windowPos.X - walkSpeed
			if nextX >= 0 {
				g.windowPos.X = nextX
			} else {
				g.windowPos.X = 0
				g.updateCurrentAction(ActionTypeIdle)
				return
			}

		case ActionTypeWalkingRight:
			maxX := g.screenDimension.Width - g.windowDimension.Width
			nextX := g.windowPos.X + walkSpeed
			if nextX <= maxX {
				g.windowPos.X = nextX
			} else {
				g.windowPos.X = maxX
				g.updateCurrentAction(ActionTypeIdle)
				return
			}
		}
	}

	switch g.currentAction.Type {
	case ActionTypeIdle:
		if imgIdx == 0 && animLoopCount > 5 {
			nextActions := []ActionType{
				ActionTypeSleep,
				ActionTypeWalkingLeft,
				ActionTypeWalkingRight,
			}

			g.updateCurrentAction(
				nextActions[rng.Intn(len(nextActions))],
			)
			return
		}

	case ActionTypeSleep:
		if frameCount == 1 {
			// There are no remaining frames to loop if sleep has only one frame.
			if animLoopCount > 15 {
				g.updateCurrentAction(ActionTypeIdle)
				return
			}
		} else if sleepLoopCount >= 15 {
			// The complete first pass has played, followed by 15 loops that omit
			// the first frame.
			g.updateCurrentAction(ActionTypeIdle)
			return
		}

	case ActionTypeWalkingLeft, ActionTypeWalkingRight:
		if imgIdx == 0 && animLoopCount > 2 {
			g.updateCurrentAction(ActionTypeIdle)
			return
		}
	}
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
	case ActionTypeHang:
		action = g.actionHang
	case ActionTypeLookAtCursor:
		action = g.actionLookAtCursor
	}

	g.currentAction = action
	g.displayImgTick = 0
}

func (g *Game) handleWakeUpKittyIfNecessary() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.leftClickTick = g.tick
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if g.leftClickTick != 0 &&
			g.tick-g.leftClickTick > holdClickTicks {
			if g.currentAction.Type != ActionTypeHang {
				g.updateCurrentAction(ActionTypeHang)
				g.playMeow()
			}
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
	if g.meowPlayer != nil {
		g.meowPlayer.Rewind()
		g.meowPlayer.Play()
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
