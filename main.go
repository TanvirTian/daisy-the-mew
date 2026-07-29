package main

import (
	"log"

	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 243
	screenHeight = 233
)

func main() {
	game, err := NewGame(GameConfig{
		ActionSourceIdle: &ActionSource{
			ImagePaths: []string{
				"assets/idle/idle.png",
			},
		},
		ActionSourceSleep: &ActionSource{
			ImagePaths: []string{
				"assets/sleep/sleep1.png",
				"assets/sleep/sleep2.png",
				"assets/sleep/sleep3.png",
				"assets/sleep/sleep4.png",
				"assets/sleep/sleep5.png",
				"assets/sleep/sleep6.png",
			},
		},
		ActionSourceWalkingLeft: &ActionSource{
			ImagePaths: []string{
				"assets/walk/wl1.png",
				"assets/walk/wl2.png",
				"assets/walk/wl3.png",
				"assets/walk/wl4.png",
				"assets/walk/wl5.png",
				"assets/walk/wl6.png",
			},
		},
		ActionSourceWalkingRight: &ActionSource{
			ImagePaths: []string{
				"assets/walk/wr1.png",
				"assets/walk/wr2.png",
				"assets/walk/wr3.png",
				"assets/walk/wr4.png",
				"assets/walk/wr5.png",
				"assets/walk/wr6.png",
			},
		},
		ActionSourceHang: &ActionSource{
			ImagePaths: []string{			
				"assets/pick/pick01.png",
				"assets/pick/pick02.png",
			},
		},
		ActionSourceLookAtCursor: &ActionSource{
			ImagePaths: []string{
				"assets/look/look1.png",
				"assets/look/look2.png",
				"assets/look/look3.png",
				"assets/look/look4.png",
				"assets/look/look5.png",
			},
		},
		WindowDimension: Dimension{
			Width:  screenWidth,
			Height: screenHeight,
		},
	})
	if err != nil {
		log.Fatalf("unable to initialize desktop pet: %v", err)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
