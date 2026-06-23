package main

import (
	"fmt"
	"image/png"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/jbetancur/the-legend-of-irondeep/engine"
	"github.com/jbetancur/the-legend-of-irondeep/render"
)

type Game struct {
	party    *engine.Party
	viewport *render.Viewport
	ui       *render.UI
	input    engine.InputState

	shotPath  string
	walkDir   string
	frame     int
	walkViews []walkView
	walkIdx   int
}

type walkView struct {
	x, y int
	dir  engine.Direction
	name string
}

func NewGame() *Game {
	engine.LoadMonsterTypes("data/monsters.json")
	world, px, py, facing := engine.LoadLevel("data/levels/level1.json")
	party := engine.NewParty(world, px, py, facing)
	assets := render.NewAssets(world.Wallset)
	viewport := render.NewViewport(assets, ScreenWidth, ScreenHeight)
	ui := render.NewUI()

	return &Game{
		party:    party,
		viewport: viewport,
		ui:       ui,
		walkViews: []walkView{
			{12, 22, engine.North, "01_entrance"},
			{12, 21, engine.North, "02_entry_hall"},
			{5, 16, engine.North, "03_west_corridor"},
			{5, 16, engine.East, "04_west_corridor_E"},
			{10, 13, engine.North, "05_central_chamber"},
			{10, 13, engine.East, "06_central_chamber_E"},
			{10, 13, engine.West, "07_central_chamber_W"},
			{2, 9, engine.North, "08_north_hall"},
			{2, 9, engine.East, "09_north_hall_E"},
			{10, 5, engine.North, "10_upper_passage"},
			{3, 2, engine.East, "11_west_room"},
			{18, 2, engine.West, "12_east_room"},
			{21, 16, engine.West, "13_door_approach"},
			{21, 19, engine.North, "14_east_alcove"},
		},
	}
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		os.Exit(0)
	}

	if g.walkDir == "" {
		g.input.HandleInput(g.party)
	}
	g.party.World.UpdateDoors()
	g.party.World.UpdateMonsters()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.walkDir != "" {
		g.drawWalkthrough(screen)
		return
	}

	g.viewport.Draw(screen, g.party)
	g.ui.Draw(screen, g.party)

	if g.shotPath != "" {
		g.frame++
		if g.frame == 5 {
			g.saveShot(screen, g.shotPath)
			os.Exit(0)
		}
	}
}

func (g *Game) drawWalkthrough(screen *ebiten.Image) {
	g.frame++
	if g.frame < 3 {
		return // let GPU init
	}

	if g.walkIdx >= len(g.walkViews) {
		fmt.Printf("walkthrough done: %d views saved to %s\n", len(g.walkViews), g.walkDir)
		os.Exit(0)
	}

	v := g.walkViews[g.walkIdx]
	g.party.X = v.x
	g.party.Y = v.y
	g.party.Facing = v.dir

	g.viewport.Draw(screen, g.party)
	g.ui.Draw(screen, g.party)
	ebitenutil.DebugPrint(screen, v.name)

	path := fmt.Sprintf("%s/%s.png", g.walkDir, v.name)
	g.saveShot(screen, path)
	g.walkIdx++
}

func (g *Game) saveShot(screen *ebiten.Image, path string) {
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	png.Encode(f, screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
