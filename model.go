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
	ActionTypeRunningLeft  ActionType = "running_left"
	ActionTypeRunningRight ActionType = "running_right"
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

func (src ActionSource) ToAction(actionType ActionType) (*Action, error) {
	if len(src.ImagePaths) == 0 {
		return nil, fmt.Errorf("action %q has no image paths", actionType)
	}

	images := make([]*ebiten.Image, 0, len(src.ImagePaths))
	for _, imagePath := range src.ImagePaths {
		image, _, err := ebitenutil.NewImageFromFile(imagePath)
		if err != nil {
			return nil, fmt.Errorf("unable to load image %q: %w", imagePath, err)
		}
		images = append(images, image)
	}

	return &Action{
		Type:   actionType,
		Images: images,
	}, nil
}

type Point struct {
	X int
	Y int
}

type Dimension struct {
	Width  int
	Height int
}

const (
	walkSpeed = 16
	runSpeed  = 90
)
