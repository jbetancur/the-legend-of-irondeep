package main

import (
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenWidth  = 1920
	ScreenHeight = 1080
)

func main() {
	ebiten.SetWindowSize(960, 540)
	ebiten.SetWindowTitle("The Legend of Irondeep")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()
	if len(os.Args) > 2 && os.Args[1] == "-shot" {
		game.shotPath = os.Args[2]
	}
	if len(os.Args) > 2 && os.Args[1] == "-walk" {
		game.walkDir = os.Args[2]
	}
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
