package main

import (
	"fmt"
	"image/color"
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
	world := engine.BuildTestLevel()
	party := engine.NewParty(world, 4, 12, engine.North)
	assets := render.NewAssets(world.Wallset)
	viewport := render.NewViewport(assets)
	ui := render.NewUI()

	return &Game{
		party:    party,
		viewport: viewport,
		ui:       ui,
		walkViews: []walkView{
			{4, 12, engine.North, "01_corridor_far"},
			{4, 11, engine.North, "02_corridor_d11"},
			{4, 10, engine.North, "03_corridor_d10"},
			{4, 9, engine.North, "04_near_junction"},
			{4, 8, engine.North, "05_at_junction_N"},
			{4, 8, engine.East, "06_at_junction_E"},
			{4, 8, engine.West, "07_at_junction_W"},
			{4, 8, engine.South, "08_at_junction_S"},
			{4, 7, engine.North, "09_past_junction"},
			{4, 5, engine.North, "10_chamber_N"},
			{4, 5, engine.East, "11_chamber_E"},
			{4, 5, engine.West, "12_chamber_W"},
			{4, 12, engine.South, "13_dead_end_S"},
			{4, 12, engine.East, "14_dead_end_E"},
			{4, 12, engine.West, "15_dead_end_W"},
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
	g.party.World.UpdateMonsters()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.walkDir != "" {
		g.drawWalkthrough(screen)
		return
	}

	screen.Fill(color.RGBA{15, 12, 10, 255})
	g.viewport.Draw(screen, g.party)
	g.ui.Draw(screen, g.party)
	ebitenutil.DebugPrint(screen, "The Legend of Irondeep")

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

	screen.Fill(color.RGBA{15, 12, 10, 255})
	g.viewport.Draw(screen, g.party)
	g.ui.Draw(screen, g.party)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("The Legend of Irondeep — %s", v.name))

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
