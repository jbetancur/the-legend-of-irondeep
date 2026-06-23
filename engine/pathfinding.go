package engine

import "container/heap"

type pathNode struct {
	x, y   int
	g, h   int
	parent *pathNode
}

func (n *pathNode) f() int { return n.g + n.h }

type pathHeap []*pathNode

func (h pathHeap) Len() int            { return len(h) }
func (h pathHeap) Less(i, j int) bool  { return h[i].f() < h[j].f() }
func (h pathHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *pathHeap) Push(x interface{}) { *h = append(*h, x.(*pathNode)) }
func (h *pathHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func manhattan(x0, y0, x1, y1 int) int {
	return abs(x0-x1) + abs(y0-y1)
}

type point struct{ x, y int }

// FindPath returns the next step (x,y) toward the target, or (-1,-1) if no path.
// The caller's own position is always passable. The target position is always
// considered reachable (so monsters can pathfind to the party's tile).
func FindPath(w *World, fromX, fromY, toX, toY int, self *Monster) (int, int, bool) {
	if fromX == toX && fromY == toY {
		return fromX, fromY, true
	}

	open := &pathHeap{}
	heap.Init(open)
	heap.Push(open, &pathNode{x: fromX, y: fromY, g: 0, h: manhattan(fromX, fromY, toX, toY)})

	closed := make(map[point]bool)
	dirs := [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

	for open.Len() > 0 {
		current := heap.Pop(open).(*pathNode)
		if current.x == toX && current.y == toY {
			// Walk back to find the first step
			n := current
			for n.parent != nil && n.parent.parent != nil {
				n = n.parent
			}
			return n.x, n.y, true
		}

		p := point{current.x, current.y}
		if closed[p] {
			continue
		}
		closed[p] = true

		for _, d := range dirs {
			nx, ny := current.x+d[0], current.y+d[1]
			np := point{nx, ny}
			if closed[np] {
				continue
			}
			if !w.IsPassable(nx, ny) {
				continue
			}
			// Allow moving to the target tile (the party) but block other occupied tiles
			if nx != toX || ny != toY {
				if w.IsOccupiedExcept(nx, ny, self) {
					continue
				}
			}

			heap.Push(open, &pathNode{
				x: nx, y: ny,
				g: current.g + 1,
				h: manhattan(nx, ny, toX, toY),
				parent: current,
			})
		}
	}

	return -1, -1, false
}
