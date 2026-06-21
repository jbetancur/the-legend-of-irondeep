package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
)

const (
	texW = 512
	texH = 512
)

func main() {
	wallset := "stone"
	if len(os.Args) > 1 {
		wallset = os.Args[1]
	}
	dir := "assets/textures/" + wallset
	os.MkdirAll(dir, 0755)

	save(dir+"/wall_front.png", generateWall())
	save(dir+"/wall_side.png", generateSideWall())
	save(dir+"/floor.png", generateFloor())
	save(dir+"/ceiling.png", generateCeiling())
	save(dir+"/door.png", generateDoor())
}

func save(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	png.Encode(f, img)
}

// Perlin-like value noise for organic variation
type noiseGen struct {
	perm [512]int
}

func newNoise(seed int64) *noiseGen {
	n := &noiseGen{}
	r := rand.New(rand.NewSource(seed))
	p := make([]int, 256)
	for i := range p {
		p[i] = i
	}
	r.Shuffle(256, func(i, j int) { p[i], p[j] = p[j], p[i] })
	for i := 0; i < 512; i++ {
		n.perm[i] = p[i%256]
	}
	return n
}

func (n *noiseGen) noise2D(x, y float64) float64 {
	X := int(math.Floor(x)) & 255
	Y := int(math.Floor(y)) & 255
	x -= math.Floor(x)
	y -= math.Floor(y)
	u := fade(x)
	v := fade(y)

	A := n.perm[X] + Y
	B := n.perm[X+1] + Y

	return lerp(v,
		lerp(u, grad(n.perm[A], x, y), grad(n.perm[B], x-1, y)),
		lerp(u, grad(n.perm[A+1], x, y-1), grad(n.perm[B+1], x-1, y-1)))
}

func (n *noiseGen) fbm(x, y float64, octaves int, lacunarity, persistence float64) float64 {
	val := 0.0
	amp := 1.0
	freq := 1.0
	max := 0.0
	for i := 0; i < octaves; i++ {
		val += n.noise2D(x*freq, y*freq) * amp
		max += amp
		amp *= persistence
		freq *= lacunarity
	}
	return val / max
}

func fade(t float64) float64 { return t * t * t * (t*(t*6-15) + 10) }
func lerp(t, a, b float64) float64 { return a + t*(b-a) }
func grad(hash int, x, y float64) float64 {
	h := hash & 3
	u, v := x, y
	if h&1 != 0 {
		u = -u
	}
	if h&2 != 0 {
		v = -v
	}
	return u + v
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampU8(v float64) uint8 {
	return uint8(clampF(v, 0, 255))
}

func generateWall() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, texW, texH))
	n1 := newNoise(42)
	n2 := newNoise(137)
	n3 := newNoise(999)
	rng := rand.New(rand.NewSource(7))

	// Stone block layout
	type block struct {
		x, y, w, h int
		rOff, gOff, bOff float64
	}

	var blocks []block
	rowH := []int{58, 62, 54, 66, 56, 60, 64, 52}
	mortarW := 3
	y := 0
	row := 0
	for y < texH {
		h := rowH[row%len(rowH)]
		if y+h > texH {
			h = texH - y
		}
		colW := []int{72, 88, 64, 96, 80, 70, 92, 76}
		offset := 0
		if row%2 == 1 {
			offset = colW[0] / 2
		}
		x := -offset
		col := 0
		for x < texW {
			w := colW[col%len(colW)]
			blocks = append(blocks, block{
				x: x, y: y, w: w, h: h,
				rOff: rng.Float64()*16 - 8,
				gOff: rng.Float64()*16 - 8,
				bOff: rng.Float64()*12 - 6,
			})
			x += w + mortarW
			col++
		}
		y += h + mortarW
		row++
	}

	// Base colors
	baseR, baseG, baseB := 115.0, 105.0, 90.0
	mortarR, mortarG, mortarB := 55.0, 50.0, 42.0

	for py := 0; py < texH; py++ {
		for px := 0; px < texW; px++ {
			fx := float64(px) / float64(texW)
			fy := float64(py) / float64(texH)

			// Determine if pixel is mortar or stone
			inMortar := false
			var blk *block
			for i := range blocks {
				b := &blocks[i]
				bx := px
				if bx < b.x {
					bx += texW
				}
				if bx >= b.x && bx < b.x+b.w && py >= b.y && py < b.y+b.h {
					// Check if near edge (mortar zone)
					dx := min(bx-b.x, b.x+b.w-1-bx)
					dy := min(py-b.y, b.y+b.h-1-py)
					if dx < mortarW || dy < mortarW {
						inMortar = true
					} else {
						blk = b
					}
					break
				}
			}

			var r, g, b float64
			if inMortar || blk == nil {
				// Mortar with slight noise
				mn := n1.fbm(fx*20, fy*20, 3, 2.0, 0.5) * 10
				r = mortarR + mn
				g = mortarG + mn*0.9
				b = mortarB + mn*0.8
				// Mortar depth shadow
				r *= 0.85
				g *= 0.85
				b *= 0.85
			} else {
				// Stone surface
				// Large-scale color variation per block
				r = baseR + blk.rOff
				g = baseG + blk.gOff
				b = baseB + blk.bOff

				// Multi-octave noise for surface texture
				surf := n1.fbm(fx*8+blk.rOff, fy*8+blk.gOff, 5, 2.2, 0.55)
				r += surf * 20
				g += surf * 18
				b += surf * 15

				// Fine grain noise
				grain := n2.fbm(fx*32, fy*32, 3, 2.0, 0.5)
				r += grain * 8
				g += grain * 7
				b += grain * 6

				// Subtle cracks/veins
				crack := n3.fbm(fx*12, fy*12, 4, 2.5, 0.6)
				if crack > 0.35 {
					darkening := (crack - 0.35) * 30
					r -= darkening
					g -= darkening
					b -= darkening
				}

				// Edge darkening within each stone (ambient occlusion)
				bx := px
				if bx < blk.x {
					bx += texW
				}
				dx := float64(min(bx-blk.x, blk.x+blk.w-1-bx))
				dy := float64(min(py-blk.y, blk.y+blk.h-1-py))
				edgeDist := math.Min(dx, dy)
				if edgeDist < 8 {
					ao := 1.0 - (8-edgeDist)/8*0.25
					r *= ao
					g *= ao
					b *= ao
				}

				// Occasional darker staining patches
				stain := n2.fbm(fx*4+0.5, fy*4+0.5, 3, 2.0, 0.5)
				if stain > 0.3 {
					factor := 1.0 - (stain-0.3)*0.3
					r *= factor
					g *= factor
					b *= factor * 1.05 // Slightly greenish stain
				}
			}

			// Global subtle vignette (darker at top/bottom)
			vigY := math.Abs(fy-0.5) * 2
			vig := 1.0 - vigY*vigY*0.1
			r *= vig
			g *= vig
			b *= vig

			img.SetRGBA(px, py, color.RGBA{clampU8(r), clampU8(g), clampU8(b), 255})
		}
	}

	return img
}

func generateSideWall() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, texW, texH))
	n1 := newNoise(73)
	n2 := newNoise(211)
	rng := rand.New(rand.NewSource(31))

	// Slightly darker and more worn for side walls
	baseR, baseG, baseB := 100.0, 92.0, 78.0
	mortarR, mortarG, mortarB := 48.0, 43.0, 37.0

	type block struct {
		x, y, w, h int
		rOff, gOff, bOff float64
	}

	var blocks []block
	rowH := []int{60, 55, 65, 58, 62, 53, 67, 56}
	mortarW := 3
	y := 0
	row := 0
	for y < texH {
		h := rowH[row%len(rowH)]
		if y+h > texH {
			h = texH - y
		}
		colW := []int{80, 70, 90, 75, 85, 68, 92}
		offset := 0
		if row%2 == 1 {
			offset = colW[0] / 2
		}
		x := -offset
		col := 0
		for x < texW {
			w := colW[col%len(colW)]
			blocks = append(blocks, block{
				x: x, y: y, w: w, h: h,
				rOff: rng.Float64()*20 - 10,
				gOff: rng.Float64()*20 - 10,
				bOff: rng.Float64()*14 - 7,
			})
			x += w + mortarW
			col++
		}
		y += h + mortarW
		row++
	}

	for py := 0; py < texH; py++ {
		for px := 0; px < texW; px++ {
			fx := float64(px) / float64(texW)
			fy := float64(py) / float64(texH)

			inMortar := false
			var blk *block
			for i := range blocks {
				b := &blocks[i]
				bx := px
				if bx < b.x {
					bx += texW
				}
				if bx >= b.x && bx < b.x+b.w && py >= b.y && py < b.y+b.h {
					dx := min(bx-b.x, b.x+b.w-1-bx)
					dy := min(py-b.y, b.y+b.h-1-py)
					if dx < mortarW || dy < mortarW {
						inMortar = true
					} else {
						blk = b
					}
					break
				}
			}

			var r, g, bv float64
			if inMortar || blk == nil {
				mn := n1.fbm(fx*18, fy*18, 3, 2.0, 0.5) * 8
				r = mortarR + mn
				g = mortarG + mn*0.9
				bv = mortarB + mn*0.8
				r *= 0.8
				g *= 0.8
				bv *= 0.8
			} else {
				r = baseR + blk.rOff
				g = baseG + blk.gOff
				bv = baseB + blk.bOff

				surf := n1.fbm(fx*8+blk.rOff*0.1, fy*8+blk.gOff*0.1, 5, 2.2, 0.55)
				r += surf * 18
				g += surf * 16
				bv += surf * 13

				grain := n2.fbm(fx*30, fy*30, 3, 2.0, 0.5)
				r += grain * 7
				g += grain * 6
				bv += grain * 5

				// More pronounced moss/moisture staining on side walls
				moss := n2.fbm(fx*3+2.0, fy*5+1.0, 4, 2.0, 0.6)
				if moss > 0.2 {
					intensity := (moss - 0.2) * 0.4
					r *= (1.0 - intensity)
					g *= (1.0 - intensity*0.3)
					bv *= (1.0 - intensity*0.8)
				}

				// Edge AO
				bx := px
				if bx < blk.x {
					bx += texW
				}
				dx := float64(min(bx-blk.x, blk.x+blk.w-1-bx))
				dy := float64(min(py-blk.y, blk.y+blk.h-1-py))
				edgeDist := math.Min(dx, dy)
				if edgeDist < 8 {
					ao := 1.0 - (8-edgeDist)/8*0.3
					r *= ao
					g *= ao
					bv *= ao
				}
			}

			// Vertical moisture gradient (darker toward bottom)
			moistGrad := 1.0 - fy*0.15
			r *= moistGrad
			g *= moistGrad
			bv *= moistGrad

			img.SetRGBA(px, py, color.RGBA{clampU8(r), clampU8(g), clampU8(bv), 255})
		}
	}

	return img
}

func generateFloor() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, texW, texH))
	n1 := newNoise(55)
	n2 := newNoise(189)
	n3 := newNoise(333)

	// Flagstone floor
	baseR, baseG, baseB := 75.0, 68.0, 58.0

	// Generate irregular flagstones using voronoi-like pattern
	type point struct{ x, y float64 }
	rng := rand.New(rand.NewSource(88))
	numPoints := 32
	pts := make([]point, numPoints)
	for i := range pts {
		pts[i] = point{rng.Float64(), rng.Float64()}
	}

	for py := 0; py < texH; py++ {
		for px := 0; px < texW; px++ {
			fx := float64(px) / float64(texW)
			fy := float64(py) / float64(texH)

			// Find two closest voronoi points (for edge detection)
			d1, d2 := 999.0, 999.0
			for _, p := range pts {
				// Tileable distance
				dx := math.Abs(fx - p.x)
				if dx > 0.5 {
					dx = 1.0 - dx
				}
				dy := math.Abs(fy - p.y)
				if dy > 0.5 {
					dy = 1.0 - dy
				}
				d := math.Sqrt(dx*dx + dy*dy)
				if d < d1 {
					d2 = d1
					d1 = d
				} else if d < d2 {
					d2 = d
				}
			}

			edgeDist := d2 - d1

			var r, g, b float64
			if edgeDist < 0.015 {
				// Mortar/gap between flagstones
				r = 35
				g = 30
				b = 25
				// Slight noise in mortar
				mn := n1.fbm(fx*25, fy*25, 2, 2.0, 0.5) * 5
				r += mn
				g += mn
				b += mn
			} else {
				// Stone surface with per-cell variation
				cellNoise := n1.fbm(fx*3+d1*10, fy*3+d1*10, 2, 2.0, 0.5)
				r = baseR + cellNoise*15
				g = baseG + cellNoise*13
				b = baseB + cellNoise*10

				// Surface texture
				surf := n2.fbm(fx*16, fy*16, 5, 2.0, 0.5)
				r += surf * 12
				g += surf * 10
				b += surf * 8

				// Fine roughness
				fine := n3.fbm(fx*40, fy*40, 2, 2.0, 0.5)
				r += fine * 5
				g += fine * 4
				b += fine * 3

				// Edge shadow near flagstone borders
				if edgeDist < 0.04 {
					shadow := (0.04 - edgeDist) / 0.04 * 0.3
					r *= (1.0 - shadow)
					g *= (1.0 - shadow)
					b *= (1.0 - shadow)
				}

				// Wear patterns - lighter in center of well-trodden areas
				wear := n1.fbm(fx*2, fy*2, 2, 2.0, 0.5)
				if wear > 0.3 {
					lighten := (wear - 0.3) * 8
					r += lighten
					g += lighten
					b += lighten
				}

				// Scattered dark spots (old stains, dirt)
				dirt := n2.fbm(fx*6+3.0, fy*6+3.0, 3, 2.0, 0.6)
				if dirt > 0.4 {
					factor := 1.0 - (dirt-0.4)*0.35
					r *= factor
					g *= factor
					b *= factor
				}
			}

			img.SetRGBA(px, py, color.RGBA{clampU8(r), clampU8(g), clampU8(b), 255})
		}
	}

	return img
}

func generateCeiling() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, texW, texH))
	n1 := newNoise(101)
	n2 := newNoise(277)
	n3 := newNoise(451)

	// Rough hewn stone ceiling - darker, more uniform
	baseR, baseG, baseB := 60.0, 55.0, 48.0

	for py := 0; py < texH; py++ {
		for px := 0; px < texW; px++ {
			fx := float64(px) / float64(texW)
			fy := float64(py) / float64(texH)

			r, g, b := baseR, baseG, baseB

			// Large-scale undulation (rough hewn surface)
			undulate := n1.fbm(fx*4, fy*4, 4, 2.0, 0.6)
			r += undulate * 15
			g += undulate * 13
			b += undulate * 10

			// Medium detail
			med := n2.fbm(fx*12, fy*12, 4, 2.2, 0.5)
			r += med * 10
			g += med * 9
			b += med * 7

			// Fine rough texture
			fine := n3.fbm(fx*35, fy*35, 3, 2.0, 0.5)
			r += fine * 6
			g += fine * 5
			b += fine * 4

			// Soot/smoke darkening in patches
			soot := n1.fbm(fx*5+10.0, fy*5+10.0, 3, 2.0, 0.5)
			if soot > 0.2 {
				factor := 1.0 - (soot-0.2)*0.3
				r *= factor
				g *= factor
				b *= factor
			}

			// Occasional lighter mineral deposits
			mineral := n2.fbm(fx*8+5.0, fy*3+5.0, 2, 2.0, 0.5)
			if mineral > 0.45 {
				lighten := (mineral - 0.45) * 15
				r += lighten
				g += lighten * 1.1
				b += lighten * 0.9
			}

			// Subtle cracks
			crack := n3.fbm(fx*6, fy*10, 3, 2.5, 0.6)
			if crack > 0.42 && crack < 0.46 {
				r *= 0.7
				g *= 0.7
				b *= 0.7
			}

			img.SetRGBA(px, py, color.RGBA{clampU8(r), clampU8(g), clampU8(b), 255})
		}
	}

	return img
}

func generateDoor() image.Image {
	const dW, dH = 256, 384
	img := image.NewRGBA(image.Rect(0, 0, dW, dH))
	n1 := newNoise(66)
	n2 := newNoise(202)

	// Heavy oak door with iron bands
	woodR, woodG, woodB := 55.0, 38.0, 22.0
	ironR, ironG, ironB := 50.0, 48.0, 52.0

	plankWidth := dW / 4
	bandPositions := []int{dH / 6, dH / 2, 5 * dH / 6}
	bandH := 12

	for py := 0; py < dH; py++ {
		for px := 0; px < dW; px++ {
			fx := float64(px) / float64(dW)
			fy := float64(py) / float64(dH)

			// Check if in iron band
			inBand := false
			for _, by := range bandPositions {
				if py >= by-bandH/2 && py < by+bandH/2 {
					inBand = true
					break
				}
			}

			// Check for nail heads (at band/plank intersections)
			isNail := false
			if inBand {
				for p := 0; p < 4; p++ {
					nx := plankWidth/2 + p*plankWidth
					for _, by := range bandPositions {
						dist := math.Sqrt(float64((px-nx)*(px-nx) + (py-by)*(py-by)))
						if dist < 6 {
							isNail = true
						}
					}
				}
			}

			var r, g, b float64

			if isNail {
				// Iron nail - slightly lighter than bands
				r = ironR + 15
				g = ironG + 15
				b = ironB + 15
				// Dome highlight
				r += 10
				g += 10
				b += 10
			} else if inBand {
				r, g, b = ironR, ironG, ironB
				// Iron texture - hammered look
				hamm := n1.fbm(fx*20, fy*3, 3, 2.0, 0.5)
				r += hamm * 8
				g += hamm * 7
				b += hamm * 9
				// Slight rust
				rust := n2.fbm(fx*15+1.0, fy*15+1.0, 2, 2.0, 0.5)
				if rust > 0.3 {
					r += (rust - 0.3) * 20
					g += (rust - 0.3) * 5
				}
			} else {
				r, g, b = woodR, woodG, woodB

				// Per-plank variation
				plankIdx := px / plankWidth
				plankOff := float64(plankIdx) * 7.3
				r += math.Sin(plankOff)*6 + 3
				g += math.Sin(plankOff*1.3)*4 + 2
				b += math.Sin(plankOff*0.7)*3 + 1

				// Wood grain (vertical streaks)
				grain := n1.fbm(fx*3+plankOff, fy*30, 4, 2.0, 0.55)
				r += grain * 12
				g += grain * 8
				b += grain * 5

				// Knots
				knot := n2.fbm(fx*8+plankOff, fy*8, 3, 2.5, 0.6)
				if knot > 0.4 {
					darkening := (knot - 0.4) * 25
					r -= darkening
					g -= darkening * 0.8
					b -= darkening * 0.6
				}

				// Plank edge (gap shadow)
				distToEdge := math.Min(float64(px%plankWidth), float64(plankWidth-px%plankWidth))
				if distToEdge < 3 {
					shadow := (3 - distToEdge) / 3 * 0.5
					r *= (1.0 - shadow)
					g *= (1.0 - shadow)
					b *= (1.0 - shadow)
				}

				// Aging/wear
				age := n2.fbm(fx*5+5.0, fy*5+5.0, 3, 2.0, 0.5)
				if age > 0.35 {
					factor := 1.0 - (age-0.35)*0.2
					r *= factor
					g *= factor
					b *= factor
				}
			}

			// Frame darkening at edges
			edgeX := math.Min(float64(px), float64(dW-px)) / float64(dW)
			edgeY := math.Min(float64(py), float64(dH-py)) / float64(dH)
			edge := math.Min(edgeX, edgeY)
			if edge < 0.03 {
				shadow := (0.03 - edge) / 0.03 * 0.4
				r *= (1.0 - shadow)
				g *= (1.0 - shadow)
				b *= (1.0 - shadow)
			}

			img.SetRGBA(px, py, color.RGBA{clampU8(r), clampU8(g), clampU8(b), 255})
		}
	}

	return img
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
