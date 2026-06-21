package render

import (
	"math"
	"testing"
)

const epsilon = 0.001

func approxEq(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

// --- Scaling: each layer is 50% the size of the previous layer ---

func TestCellScaling(t *testing.T) {
	for depth := 0; depth <= 3; depth++ {
		c := cellRect(depth, 0)
		next := cellRect(depth+1, 0)
		if !approxEq(next.w, c.w*0.5) {
			t.Errorf("depth %d->%d: width should halve (got %f, want %f)", depth, depth+1, next.w, c.w*0.5)
		}
		if !approxEq(next.h, c.h*0.5) {
			t.Errorf("depth %d->%d: height should halve (got %f, want %f)", depth, depth+1, next.h, c.h*0.5)
		}
	}
}

func TestViewportIsSquare(t *testing.T) {
	if ViewportW != ViewportH {
		t.Errorf("viewport should be square: %dx%d", ViewportW, ViewportH)
	}
	if ViewportW != 540 {
		t.Errorf("viewport size = %d, want 540", ViewportW)
	}
}

func TestDepthZeroIsFullViewport(t *testing.T) {
	c := cellRect(0, 0)
	if !approxEq(c.w, ViewportW) {
		t.Errorf("depth 0 width = %f, want %f", c.w, float64(ViewportW))
	}
	if !approxEq(c.h, ViewportH) {
		t.Errorf("depth 0 height = %f, want %f", c.h, float64(ViewportH))
	}
}

// --- Back wall occupies the center 50% of the cell ---

func TestBackWallIsCenterHalf(t *testing.T) {
	for depth := 0; depth <= 3; depth++ {
		c := cellRect(depth, 0)
		bw := backWallRect(depth, 0)

		if !approxEq(bw.w, c.w*0.5) {
			t.Errorf("depth %d: back wall width = %f, want %f (50%% of cell)", depth, bw.w, c.w*0.5)
		}
		if !approxEq(bw.h, c.h*0.5) {
			t.Errorf("depth %d: back wall height = %f, want %f (50%% of cell)", depth, bw.h, c.h*0.5)
		}

		bwCenterX := bw.x + bw.w/2
		cCenterX := c.x + c.w/2
		if !approxEq(bwCenterX, cCenterX) {
			t.Errorf("depth %d: back wall not centered horizontally in cell", depth)
		}
		bwCenterY := bw.y + bw.h/2
		cCenterY := c.y + c.h/2
		if !approxEq(bwCenterY, cCenterY) {
			t.Errorf("depth %d: back wall not centered vertically in cell", depth)
		}
	}
}

// --- Critical nesting property: BackWall(depth N, col) == CellRect(depth N+1, col) ---
// The back wall at ANY column must equal the next depth's cell at the same column.
// This ensures side walls converge toward the global vanishing point, not each
// cell's own center. Previously only worked for col 0.

func TestNestingProperty(t *testing.T) {
	for depth := 0; depth <= 3; depth++ {
		minCol, maxCol := columnRange(depth)
		for col := minCol; col <= maxCol; col++ {
			bw := backWallRect(depth, col)
			next := cellRect(depth+1, col)

			if !approxEq(bw.x, next.x) {
				t.Errorf("depth %d col %d: backWall.x=%f != cellRect(%d,%d).x=%f",
					depth, col, bw.x, depth+1, col, next.x)
			}
			if !approxEq(bw.y, next.y) {
				t.Errorf("depth %d col %d: backWall.y=%f != cellRect(%d,%d).y=%f",
					depth, col, bw.y, depth+1, col, next.y)
			}
			if !approxEq(bw.w, next.w) {
				t.Errorf("depth %d col %d: backWall.w=%f != cellRect(%d,%d).w=%f",
					depth, col, bw.w, depth+1, col, next.w)
			}
			if !approxEq(bw.h, next.h) {
				t.Errorf("depth %d col %d: backWall.h=%f != cellRect(%d,%d).h=%f",
					depth, col, bw.h, depth+1, col, next.h)
			}
		}
	}
}

// --- Vanishing point: all center cells are centered on the viewport center ---

func TestVanishingPointCentering(t *testing.T) {
	vpCenterX := float64(ViewportW) / 2
	vpCenterY := float64(ViewportH) / 2

	for depth := 0; depth <= 4; depth++ {
		c := cellRect(depth, 0)
		cx := c.x + c.w/2
		cy := c.y + c.h/2
		if !approxEq(cx, vpCenterX) {
			t.Errorf("depth %d: center cell midX=%f, want viewport center %f", depth, cx, vpCenterX)
		}
		if !approxEq(cy, vpCenterY) {
			t.Errorf("depth %d: center cell midY=%f, want viewport center %f", depth, cy, vpCenterY)
		}
	}
}

// --- Column count: 2*depth + 1 cells per layer ---

func TestColumnCount(t *testing.T) {
	cases := []struct {
		depth int
		want  int
	}{
		{0, 3},
		{1, 5},
		{2, 7},
		{3, 9},
	}
	for _, tc := range cases {
		min, max := columnRange(tc.depth)
		got := max - min + 1
		if got != tc.want {
			t.Errorf("depth %d: column count = %d, want %d", tc.depth, got, tc.want)
		}
	}
}

// --- Side walls are exactly 25% of cell width ---

func TestSideWallWidth(t *testing.T) {
	for depth := 0; depth <= 3; depth++ {
		c := cellRect(depth, 0)
		bw := backWallRect(depth, 0)

		leftGap := bw.left() - c.left()
		rightGap := c.right() - bw.right()
		quarter := c.w * 0.25

		if !approxEq(leftGap, quarter) {
			t.Errorf("depth %d: left wall width = %f, want %f (25%%)", depth, leftGap, quarter)
		}
		if !approxEq(rightGap, quarter) {
			t.Errorf("depth %d: right wall width = %f, want %f (25%%)", depth, rightGap, quarter)
		}
	}
}

// --- Wall quads connect cell edges to back wall edges ---

func TestWallQuadGeometry(t *testing.T) {
	for depth := 0; depth <= 3; depth++ {
		c := cellRect(depth, 0)
		bw := backWallRect(depth, 0)

		// Left wall: outer = cell left edge, inner = back wall left edge
		lx0, ly0, lx1, ly1, _, ly2, lx3, ly3 := leftWallQuad(depth, 0)
		if !approxEq(lx0, c.left()) || !approxEq(lx3, c.left()) {
			t.Errorf("depth %d: left wall outer not at cell left", depth)
		}
		if !approxEq(lx1, bw.left()) {
			t.Errorf("depth %d: left wall inner not at back wall left", depth)
		}
		if !approxEq(ly0, c.top()) || !approxEq(ly3, c.bottom()) {
			t.Errorf("depth %d: left wall outer should span cell height", depth)
		}
		if !approxEq(ly1, bw.top()) || !approxEq(ly2, bw.bottom()) {
			t.Errorf("depth %d: left wall inner should span back wall height", depth)
		}

		// Right wall: outer = cell right edge, inner = back wall right edge
		_, ry0, rx1, _, rx2, ry2, _, ry3 := rightWallQuad(depth, 0)
		if !approxEq(rx1, c.right()) || !approxEq(rx2, c.right()) {
			t.Errorf("depth %d: right wall outer not at cell right", depth)
		}
		if !approxEq(ry0, bw.top()) || !approxEq(ry3, bw.bottom()) {
			t.Errorf("depth %d: right wall inner should span back wall height", depth)
		}
		if !approxEq(ry2, c.bottom()) {
			t.Errorf("depth %d: right wall outer should reach cell bottom", depth)
		}
	}
}

// --- Side wall quads at off-center columns converge to vanishing point ---

func TestSideWallsConvergeToVanishingPoint(t *testing.T) {
	vpCX := float64(ViewportW) / 2

	for depth := 0; depth <= 2; depth++ {
		minCol, maxCol := columnRange(depth)
		for col := minCol; col <= maxCol; col++ {
			// Left wall: inner X must be closer to VP center than outer X
			lxOuter, _, lxInner, _, _, _, _, _ := leftWallQuad(depth, col)
			if math.Abs(lxInner-vpCX) > math.Abs(lxOuter-vpCX)+epsilon {
				t.Errorf("depth %d col %d: left wall inner X (%.1f) farther from VP center (%.1f) than outer X (%.1f)",
					depth, col, lxInner, vpCX, lxOuter)
			}

			// Right wall: inner X must be closer to VP center than outer X
			rxInner, _, rxOuter, _, _, _, _, _ := rightWallQuad(depth, col)
			if math.Abs(rxInner-vpCX) > math.Abs(rxOuter-vpCX)+epsilon {
				t.Errorf("depth %d col %d: right wall inner X (%.1f) farther from VP center (%.1f) than outer X (%.1f)",
					depth, col, rxInner, vpCX, rxOuter)
			}
		}
	}
}

// --- Adjacent cells tile horizontally without gaps or overlaps ---

func TestAdjacentCellsTileHorizontally(t *testing.T) {
	for depth := 1; depth <= 3; depth++ {
		min, max := columnRange(depth)
		for col := min; col < max; col++ {
			left := cellRect(depth, col)
			right := cellRect(depth, col+1)
			if !approxEq(left.right(), right.left()) {
				t.Errorf("depth %d: gap between col %d right (%f) and col %d left (%f)",
					depth, col, left.right(), col+1, right.left())
			}
		}
	}
}

// --- All cells at the same depth share the same vertical position ---

func TestCellsShareVerticalPosition(t *testing.T) {
	for depth := 1; depth <= 3; depth++ {
		center := cellRect(depth, 0)
		min, max := columnRange(depth)
		for col := min; col <= max; col++ {
			c := cellRect(depth, col)
			if !approxEq(c.top(), center.top()) || !approxEq(c.bottom(), center.bottom()) {
				t.Errorf("depth %d col %d: vertical mismatch (top=%f, want %f)",
					depth, col, c.top(), center.top())
			}
		}
	}
}

// --- Back walls at any column have the same dimensions as the next layer's cells ---
// The back wall SIZE should always equal the next depth's cell size (both halve),
// even though off-center back walls have different positions.

// --- Shading: 50% brightness per depth, smooth per-vertex interpolation ---

func TestCellShade50PercentPerLayer(t *testing.T) {
	cases := []struct {
		depth int
		want  float32
	}{
		{0, 1.0},
		{1, 0.5},
		{2, 0.25},
		{3, 0.125},
	}
	for _, tc := range cases {
		got := cellShade(tc.depth)
		if math.Abs(float64(got-tc.want)) > 0.001 {
			t.Errorf("cellShade(%d) = %f, want %f", tc.depth, got, tc.want)
		}
	}
}

func TestBackWallSizeMatchesNextLayer(t *testing.T) {
	for depth := 0; depth <= 2; depth++ {
		min, max := columnRange(depth)
		for col := min; col <= max; col++ {
			bw := backWallRect(depth, col)
			next := cellRect(depth+1, 0)
			if !approxEq(bw.w, next.w) {
				t.Errorf("depth %d col %d: backWall.w=%f != nextLayer cellW=%f",
					depth, col, bw.w, next.w)
			}
			if !approxEq(bw.h, next.h) {
				t.Errorf("depth %d col %d: backWall.h=%f != nextLayer cellH=%f",
					depth, col, bw.h, next.h)
			}
		}
	}
}
