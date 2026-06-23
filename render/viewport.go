package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/jbetancur/the-legend-of-irondeep/engine"
)

const (
	BaseViewportH = 540
	maxDepth      = 2
)

var (
	ViewportW = 540
	ViewportH = 540
)

type Viewport struct {
	Assets *Assets
	Image  *ebiten.Image
}

func NewViewport(assets *Assets, screenW, screenH int) *Viewport {
	aspect := float64(screenW) / float64(screenH)
	ViewportW = int(float64(BaseViewportH) * aspect)
	ViewportH = BaseViewportH
	return &Viewport{
		Assets: assets,
		Image:  ebiten.NewImage(ViewportW, ViewportH),
	}
}

func (v *Viewport) Draw(screen *ebiten.Image, party *engine.Party) {
	v.Image.Fill(color.RGBA{5, 4, 3, 255})

	midY := float64(ViewportH) / 2
	vpW := float64(ViewportW)
	vpH := float64(ViewportH)
	darkest := cellShade(maxDepth + 1)
	v.drawQuad(v.Assets.Ceiling, 0, 0, vpW, 0, vpW, midY, 0, midY,
		[4]float32{1, 1, darkest, darkest})
	v.drawQuad(v.Assets.Floor, 0, midY, vpW, midY, vpW, vpH, 0, vpH,
		[4]float32{darkest, darkest, 1, 1})

	for depth := maxDepth; depth >= 0; depth-- {
		minCol, maxCol := columnRange(depth)
		for col := minCol; col <= maxCol; col++ {
			v.drawCell(party, depth, col)
		}
	}

	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(sw)/vpW, float64(sh)/vpH)
	screen.DrawImage(v.Image, op)
}

func (v *Viewport) drawCell(party *engine.Party, depth, col int) {
	w := party.World
	facing := party.Facing
	dx, dy := facing.Delta()
	leftDir := facing.TurnLeft()
	ldx, ldy := leftDir.Delta()

	cx := party.X + dx*depth + ldx*(-col)
	cy := party.Y + dy*depth + ldy*(-col)

	shade := cellShade(depth)

	if !w.IsPassable(cx, cy) {
		if depth > 0 {
			behindX := cx - dx
			behindY := cy - dy
			if w.IsPassable(behindX, behindY) {
				v.blitRect(v.Assets.FrontWall, cellRect(depth, col), shade)
			}
		}
		return
	}

	// Outer edge (cell boundary) = current depth shade; inner edge (back wall) = next depth shade.
	innerShade := shade * 0.5

	leftCellX := cx + ldx
	leftCellY := cy + ldy
	if w.IsWall(leftCellX, leftCellY) {
		x0, y0, x1, y1, x2, y2, x3, y3 := leftWallQuad(depth, col)
		s := [4]float32{shade, innerShade, innerShade, shade}
		tex := v.Assets.SideWallL
		if w.IsDoorFrame(leftCellX, leftCellY) {
			tex = v.Assets.DoorFrameL
		}
		v.drawQuad(tex, x0, y0, x1, y1, x2, y2, x3, y3, s)
	}

	rightCellX := cx - ldx
	rightCellY := cy - ldy
	if w.IsWall(rightCellX, rightCellY) {
		x0, y0, x1, y1, x2, y2, x3, y3 := rightWallQuad(depth, col)
		s := [4]float32{innerShade, shade, shade, innerShade}
		tex := v.Assets.SideWallR
		if w.IsDoorFrame(rightCellX, rightCellY) {
			tex = v.Assets.DoorFrameR
		}
		v.drawQuad(tex, x0, y0, x1, y1, x2, y2, x3, y3, s)
	}

	aheadX := cx + dx
	aheadY := cy + dy
	if w.IsWall(aheadX, aheadY) {
		bw := backWallRect(depth, col)
		v.blitRect(v.Assets.FrontWall, bw, cellShade(depth+1))
	}
	if w.IsDoor(aheadX, aheadY) {
		progress := w.DoorProgress(aheadX, aheadY)
		if progress < 1.0 {
			bw := backWallRect(depth, col)
			visibleH := bw.h * (1.0 - progress)
			dstTop := bw.y + bw.h - visibleH
			sh := cellShade(depth + 1)
			v.drawDoorSlice(v.Assets.Door, bw.x, dstTop, bw.w, visibleH, progress, sh)
		}
	}

	monsters := w.MonstersAt(cx, cy)
	count := len(monsters)
	if count > 2 {
		count = 2
	}
	for i := 0; i < count; i++ {
		sprite := v.Assets.Monsters[monsters[i].Type.Sprite]
		if sprite == nil {
			continue
		}
		v.drawMonster(sprite, depth, col, shade, i, count)
	}
}

func (v *Viewport) drawMonster(sprite *ebiten.Image, depth, col int, shade float32, index, total int) {
	c := cellRect(depth, col)

	sw := float64(sprite.Bounds().Dx())
	sh := float64(sprite.Bounds().Dy())
	aspect := sw / sh

	fitH := c.h * 0.85
	fitW := fitH * aspect
	maxW := c.w * 0.7
	if total == 2 {
		maxW = c.w * 0.45
	}
	if fitW > maxW {
		fitW = maxW
		fitH = fitW / aspect
	}

	midX := c.x + c.w/2
	if total == 2 {
		if index == 0 {
			midX = c.x + c.w*0.3
		} else {
			midX = c.x + c.w*0.7
		}
	}
	dstX := midX - fitW/2
	dstY := c.y + c.h - fitH - c.h*0.02

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(fitW/sw, fitH/sh)
	op.GeoM.Translate(dstX, dstY)
	op.ColorScale.Scale(shade, shade, shade, 1)
	v.Image.DrawImage(sprite, op)
}

func (v *Viewport) drawDoorSlice(tex *ebiten.Image, dstX, dstY, dstW, dstH, progress float64, shade float32) {
	tw := float32(tex.Bounds().Dx())
	th := float32(tex.Bounds().Dy())
	srcTop := th * float32(progress)
	vtx := []ebiten.Vertex{
		{DstX: float32(dstX), DstY: float32(dstY), SrcX: 0, SrcY: srcTop, ColorR: shade, ColorG: shade, ColorB: shade, ColorA: 1},
		{DstX: float32(dstX + dstW), DstY: float32(dstY), SrcX: tw, SrcY: srcTop, ColorR: shade, ColorG: shade, ColorB: shade, ColorA: 1},
		{DstX: float32(dstX + dstW), DstY: float32(dstY + dstH), SrcX: tw, SrcY: th, ColorR: shade, ColorG: shade, ColorB: shade, ColorA: 1},
		{DstX: float32(dstX), DstY: float32(dstY + dstH), SrcX: 0, SrcY: th, ColorR: shade, ColorG: shade, ColorB: shade, ColorA: 1},
	}
	idx := []uint16{0, 1, 2, 0, 2, 3}
	v.Image.DrawTriangles(vtx, idx, tex, &ebiten.DrawTrianglesOptions{})
}

func cellShade(depth int) float32 {
	// 50% brightness reduction per depth layer (per SBS article).
	s := float32(1.0)
	for i := 0; i < depth; i++ {
		s *= 0.5
	}
	return s
}

func (v *Viewport) blitRect(tex *ebiten.Image, r rect, shade float32) {
	s := [4]float32{shade, shade, shade, shade}
	v.drawQuad(tex,
		r.left(), r.top(), r.right(), r.top(),
		r.right(), r.bottom(), r.left(), r.bottom(),
		s)
}

func (v *Viewport) drawQuad(tex *ebiten.Image,
	x0, y0, x1, y1, x2, y2, x3, y3 float64, shade [4]float32) {

	tw := float32(tex.Bounds().Dx())
	th := float32(tex.Bounds().Dy())

	isRect := x0 == x3 && x1 == x2 && y0 == y1 && y2 == y3
	if isRect {
		vtx := []ebiten.Vertex{
			{DstX: float32(x0), DstY: float32(y0), SrcX: 0, SrcY: 0, ColorR: shade[0], ColorG: shade[0], ColorB: shade[0], ColorA: 1},
			{DstX: float32(x1), DstY: float32(y1), SrcX: tw, SrcY: 0, ColorR: shade[1], ColorG: shade[1], ColorB: shade[1], ColorA: 1},
			{DstX: float32(x2), DstY: float32(y2), SrcX: tw, SrcY: th, ColorR: shade[2], ColorG: shade[2], ColorB: shade[2], ColorA: 1},
			{DstX: float32(x3), DstY: float32(y3), SrcX: 0, SrcY: th, ColorR: shade[3], ColorG: shade[3], ColorB: shade[3], ColorA: 1},
		}
		idx := []uint16{0, 1, 2, 0, 2, 3}
		v.Image.DrawTriangles(vtx, idx, tex, &ebiten.DrawTrianglesOptions{})
		return
	}

	// Subdivide trapezoids into horizontal strips to avoid affine warping
	// artifacts from DrawTriangles. EoB2 drew side walls as scanline strips.
	const strips = 32
	for i := 0; i < strips; i++ {
		t0 := float64(i) / float64(strips)
		t1 := float64(i+1) / float64(strips)

		lx0 := x0 + (x3-x0)*t0
		ly0 := y0 + (y3-y0)*t0
		rx0 := x1 + (x2-x1)*t0
		ry0 := y1 + (y2-y1)*t0

		lx1 := x0 + (x3-x0)*t1
		ly1 := y0 + (y3-y0)*t1
		rx1 := x1 + (x2-x1)*t1
		ry1 := y1 + (y2-y1)*t1

		sy0 := th * float32(t0)
		sy1 := th * float32(t1)

		s0L := shade[0] + (shade[3]-shade[0])*float32(t0)
		s0R := shade[1] + (shade[2]-shade[1])*float32(t0)
		s1L := shade[0] + (shade[3]-shade[0])*float32(t1)
		s1R := shade[1] + (shade[2]-shade[1])*float32(t1)

		vtx := []ebiten.Vertex{
			{DstX: float32(lx0), DstY: float32(ly0), SrcX: 0, SrcY: sy0, ColorR: s0L, ColorG: s0L, ColorB: s0L, ColorA: 1},
			{DstX: float32(rx0), DstY: float32(ry0), SrcX: tw, SrcY: sy0, ColorR: s0R, ColorG: s0R, ColorB: s0R, ColorA: 1},
			{DstX: float32(rx1), DstY: float32(ry1), SrcX: tw, SrcY: sy1, ColorR: s1R, ColorG: s1R, ColorB: s1R, ColorA: 1},
			{DstX: float32(lx1), DstY: float32(ly1), SrcX: 0, SrcY: sy1, ColorR: s1L, ColorG: s1L, ColorB: s1L, ColorA: 1},
		}
		idx := []uint16{0, 1, 2, 0, 2, 3}
		v.Image.DrawTriangles(vtx, idx, tex, &ebiten.DrawTrianglesOptions{})
	}
}
