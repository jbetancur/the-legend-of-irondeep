package render

// Cell-grid rendering system based on the Screaming Brain Studios technique.
//
// Floor/ceiling is a full-viewport backdrop drawn first (per EoB2 source).
// Each cell draws walls only: back wall, left wall, right wall.
// Cells at each layer are 50% the size of the previous, centered on the
// vanishing point.
//
// Layer 0 (party's cell): 3 cells (left, center, right)
// Layer 1: 5 cells
// Layer 2: 7 cells
// Layer 3: 9 cells
//
// Cell positions are indexed from center: 0 = center, -1 = left, +1 = right,
// -2 = far-left, etc.

type rect struct {
	x, y, w, h float64
}

func (r rect) left() float64   { return r.x }
func (r rect) right() float64  { return r.x + r.w }
func (r rect) top() float64    { return r.y }
func (r rect) bottom() float64 { return r.y + r.h }

// cellRect returns the bounding rect of a cell at (depth, column) in viewport
// space. Column 0 is center; -1 is one cell left; +1 is one cell right.
func cellRect(depth, col int) rect {
	scale := 1.0
	for i := 0; i < depth; i++ {
		scale *= 0.5
	}

	cellW := float64(ViewportW) * scale
	cellH := float64(ViewportH) * scale

	// Center the grid on the vanishing point.
	cx := float64(ViewportW) / 2
	cy := float64(ViewportH) / 2

	x := cx - cellW/2 + float64(col)*cellW
	y := cy - cellH/2

	return rect{x: x, y: y, w: cellW, h: cellH}
}

// backWallRect returns the back wall of cell (depth, col).
// Per the SBS nesting rule, this equals cellRect(depth+1, col):
// the inner rectangle converges toward the viewport vanishing point,
// not toward the center of the current cell.
func backWallRect(depth, col int) rect {
	return cellRect(depth+1, col)
}

// leftWallQuad returns the 4 corners (clockwise from top-left) of the left wall
// trapezoid within a cell — the left 25%, skewed to meet the back wall.
func leftWallQuad(depth, col int) (x0, y0, x1, y1, x2, y2, x3, y3 float64) {
	c := cellRect(depth, col)
	bw := backWallRect(depth, col)
	// Top-left of cell -> top-left of back wall -> bottom-left of back wall -> bottom-left of cell
	return c.left(), c.top(), bw.left(), bw.top(), bw.left(), bw.bottom(), c.left(), c.bottom()
}

// rightWallQuad returns the 4 corners of the right wall trapezoid.
func rightWallQuad(depth, col int) (x0, y0, x1, y1, x2, y2, x3, y3 float64) {
	c := cellRect(depth, col)
	bw := backWallRect(depth, col)
	return bw.right(), bw.top(), c.right(), c.top(), c.right(), c.bottom(), bw.right(), bw.bottom()
}

// columnRange returns the min and max column indices at a given depth.
// Per the Screaming Brain Studios technique, the closest layer has 3 cells
// (left, center, right), and each deeper layer adds 2 more:
// depth 0: 3 cells [-1, 0, 1], depth 1: 5, depth 2: 7, depth 3: 9.
func columnRange(depth int) (int, int) {
	n := depth + 1
	return -n, n
}
