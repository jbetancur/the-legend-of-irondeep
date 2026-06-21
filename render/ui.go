package render

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/jbetancur/the-legend-of-irondeep/engine"
)

type UI struct {
	barBg    *ebiten.Image
	sideFill *ebiten.Image
}

func NewUI() *UI {
	bg := ebiten.NewImage(ScreenW, partyBarH)
	bg.Fill(color.RGBA{38, 32, 28, 255})
	sep := color.RGBA{70, 60, 52, 255}
	for x := 0; x < ScreenW; x++ {
		bg.Set(x, 0, sep)
		bg.Set(x, 1, sep)
	}

	side := ebiten.NewImage(viewportScreenX, viewportDisplayH)
	side.Fill(color.RGBA{22, 19, 17, 255})

	return &UI{barBg: bg, sideFill: side}
}

func (u *UI) Draw(screen *ebiten.Image, party *engine.Party) {
	// Fill the dark margins beside the centered viewport.
	left := &ebiten.DrawImageOptions{}
	screen.DrawImage(u.sideFill, left)
	right := &ebiten.DrawImageOptions{}
	right.GeoM.Translate(viewportScreenX+viewportDisplayW, 0)
	screen.DrawImage(u.sideFill, right)

	u.drawMinimap(screen, party)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, partyBarY)
	screen.DrawImage(u.barBg, op)

	// Status line above the character slots.
	info := fmt.Sprintf("Pos: %d, %d   Facing: %s        WASD/Arrows: Move/Turn    Q/E: Strafe    ESC: Quit",
		party.X, party.Y, party.Facing)
	ebitenutil.DebugPrintAt(screen, info, 24, partyBarY+16)

	// Four character slots in a horizontal row.
	pad := 24
	top := partyBarY + 48
	gap := 16
	slotH := partyBarH - 48 - pad
	totalGaps := gap * 3
	slotW := (ScreenW - pad*2 - totalGaps) / 4

	for i := 0; i < 4; i++ {
		x := pad + i*(slotW+gap)
		drawCharSlot(screen, x, top, slotW, slotH, i+1)
	}
}

func (u *UI) drawMinimap(screen *ebiten.Image, party *engine.Party) {
	w := party.World
	cellSize := 14
	mapW := w.Width * cellSize
	mapH := w.Height * cellSize

	// Position in the left margin, vertically centered
	ox := (viewportScreenX - mapW) / 2
	if ox < 8 {
		ox = 8
	}
	oy := (viewportDisplayH - mapH) / 2
	if oy < 8 {
		oy = 8
	}

	wall := color.RGBA{70, 65, 58, 255}
	floor := color.RGBA{30, 26, 22, 255}
	border := color.RGBA{50, 45, 40, 255}

	for y := 0; y < w.Height; y++ {
		for x := 0; x < w.Width; x++ {
			px := ox + x*cellSize
			py := oy + y*cellSize
			c := wall
			if w.IsPassable(x, y) {
				c = floor
			}
			cell := ebiten.NewImage(cellSize, cellSize)
			cell.Fill(c)
			// border between cells
			for i := 0; i < cellSize; i++ {
				cell.Set(i, 0, border)
				cell.Set(0, i, border)
			}
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(px), float64(py))
			screen.DrawImage(cell, op)
		}
	}

	// Party position
	px := ox + party.X*cellSize + cellSize/2
	py := oy + party.Y*cellSize + cellSize/2
	partyColor := color.RGBA{220, 180, 40, 255}
	dot := ebiten.NewImage(6, 6)
	dot.Fill(partyColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(px-3), float64(py-3))
	screen.DrawImage(dot, op)

	// Facing indicator
	fdx, fdy := party.Facing.Delta()
	arrowX := px + fdx*5
	arrowY := py + fdy*5
	arrow := ebiten.NewImage(4, 4)
	arrow.Fill(color.RGBA{255, 220, 60, 255})
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(arrowX-2), float64(arrowY-2))
	screen.DrawImage(arrow, op2)
}

func drawCharSlot(screen *ebiten.Image, x, y, w, h, slot int) {
	img := ebiten.NewImage(w, h)
	img.Fill(color.RGBA{52, 46, 40, 255})

	border := color.RGBA{90, 78, 66, 255}
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
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Character %d", slot), x+12, y+12)
	ebitenutil.DebugPrintAt(screen, rowName, x+12, y+32)
}
