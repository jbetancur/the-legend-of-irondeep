package render

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/jbetancur/the-legend-of-irondeep/engine"
)

type UI struct{}

func NewUI() *UI {
	return &UI{}
}

func (u *UI) Draw(screen *ebiten.Image, party *engine.Party) {
	u.drawMinimap(screen, party)
	u.drawPartyOverlay(screen)

	info := fmt.Sprintf("Pos: %d, %d   Facing: %s", party.X, party.Y, party.Facing)
	ebitenutil.DebugPrintAt(screen, info, 10, ScreenH-20)
}

func (u *UI) drawPartyOverlay(screen *ebiten.Image) {
	panelW := 440
	panelH := 280
	px := ScreenW - panelW - 16
	py := ScreenH - panelH - 16

	bg := ebiten.NewImage(panelW, panelH)
	bg.Fill(color.RGBA{20, 18, 15, 180})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(px), float64(py))
	screen.DrawImage(bg, op)

	border := color.RGBA{90, 78, 66, 200}
	for x := 0; x < panelW; x++ {
		bg.Set(x, 0, border)
		bg.Set(x, panelH-1, border)
	}
	for y := 0; y < panelH; y++ {
		bg.Set(0, y, border)
		bg.Set(panelW-1, y, border)
	}
	screen.DrawImage(bg, op)

	gap := 8
	slotW := (panelW - gap*5) / 4
	slotH := panelH - gap*2
	for i := 0; i < 4; i++ {
		sx := px + gap + i*(slotW+gap)
		sy := py + gap
		drawCharSlot(screen, sx, sy, slotW, slotH, i+1)
	}
}

func (u *UI) drawMinimap(screen *ebiten.Image, party *engine.Party) {
	w := party.World
	cellSize := 10
	mapW := w.Width * cellSize
	mapH := w.Height * cellSize
	pad := 12

	bg := ebiten.NewImage(mapW+pad*2, mapH+pad*2)
	bg.Fill(color.RGBA{10, 9, 8, 160})
	bgOp := &ebiten.DrawImageOptions{}
	bgOp.GeoM.Translate(float64(pad/2), float64(pad/2))
	screen.DrawImage(bg, bgOp)

	ox := pad
	oy := pad

	wall := color.RGBA{70, 65, 58, 255}
	floor := color.RGBA{30, 26, 22, 255}
	doorColor := color.RGBA{100, 70, 40, 255}
	cellBorder := color.RGBA{50, 45, 40, 255}

	for y := 0; y < w.Height; y++ {
		for x := 0; x < w.Width; x++ {
			px := ox + x*cellSize
			py := oy + y*cellSize
			c := wall
			if w.IsPassable(x, y) {
				c = floor
			} else if w.IsDoor(x, y) {
				c = doorColor
			}
			cell := ebiten.NewImage(cellSize, cellSize)
			cell.Fill(c)
			for i := 0; i < cellSize; i++ {
				cell.Set(i, 0, cellBorder)
				cell.Set(0, i, cellBorder)
			}
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(px), float64(py))
			screen.DrawImage(cell, op)
		}
	}

	ppx := ox + party.X*cellSize + cellSize/2
	ppy := oy + party.Y*cellSize + cellSize/2
	partyColor := color.RGBA{220, 180, 40, 255}
	dot := ebiten.NewImage(6, 6)
	dot.Fill(partyColor)
	dotOp := &ebiten.DrawImageOptions{}
	dotOp.GeoM.Translate(float64(ppx-3), float64(ppy-3))
	screen.DrawImage(dot, dotOp)

	fdx, fdy := party.Facing.Delta()
	arrowX := ppx + fdx*5
	arrowY := ppy + fdy*5
	arrow := ebiten.NewImage(4, 4)
	arrow.Fill(color.RGBA{255, 220, 60, 255})
	arrowOp := &ebiten.DrawImageOptions{}
	arrowOp.GeoM.Translate(float64(arrowX-2), float64(arrowY-2))
	screen.DrawImage(arrow, arrowOp)
}

func drawCharSlot(screen *ebiten.Image, x, y, w, h, slot int) {
	img := ebiten.NewImage(w, h)
	img.Fill(color.RGBA{40, 36, 30, 200})

	border := color.RGBA{80, 70, 58, 220}
	for px := 0; px < w; px++ {
		img.Set(px, 0, border)
		img.Set(px, h-1, border)
	}
	for py := 0; py < h; py++ {
		img.Set(0, py, border)
		img.Set(w-1, py, border)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)

	rowName := "Front Row"
	if slot > 2 {
		rowName = "Back Row"
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Char %d", slot), x+8, y+8)
	ebitenutil.DebugPrintAt(screen, rowName, x+8, y+24)
}
