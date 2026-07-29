package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type ActionType string

const (
	ActionTypeIdle         ActionType = "idle"
	ActionTypeSleep        ActionType = "sleep"
	ActionTypeWalkingLeft  ActionType = "walking_left"
	ActionTypeWalkingRight ActionType = "walking_right"
	ActionTypeHang         ActionType = "hang"
	ActionTypeLookAtCursor ActionType = "look_at_cursor"
)

type Action struct {
	Type   ActionType
	Images []*ebiten.Image
}

type ActionSource struct {
	ImagePaths []string
}

func (src ActionSource) ToAction(actType ActionType) (*Action, error) {
	images := make([]*ebiten.Image, 0, len(src.ImagePaths))
	for _, imgPath := range src.ImagePaths {
		img, _, err := ebitenutil.NewImageFromFile(imgPath)
		if err != nil {
			return nil, fmt.Errorf("unable to load image due: %v", err)
		}
		images = append(images, img)
	}
	return &Action{Type: actType, Images: images}, nil
}

type Point struct {
	X int
	Y int
}

type Dimension struct {
	Width  int
	Height int
}

const walkSpeed = 4