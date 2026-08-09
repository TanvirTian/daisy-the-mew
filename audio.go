package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
)


func initAudio() *audio.Player {
	const sampleRate = 44100
	audioContext := audio.NewContext(sampleRate)

	meowFile, err := os.Open("mew/mew.mp3")
	if err != nil {
		fmt.Printf("[Audio Error] Could not open sound file: %v\n", err)
		return nil
	}

	decodedMeow, err := mp3.DecodeWithSampleRate(sampleRate, meowFile)
	if err != nil {
		fmt.Printf("[Audio Error] Could not decode MP3 file: %v\n", err)
		return nil
	}

	player, err := audioContext.NewPlayer(decodedMeow)
	if err != nil {
		fmt.Printf("[Audio Error] Could not create player: %v\n", err)
		return nil
	}

	return player
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
