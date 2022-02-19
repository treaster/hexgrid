package hexgrid

import (
	"fmt"
)

type Coord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (c Coord) String() string {
	return fmt.Sprintf("C(%d, %d)", c.X, c.Y)
}

type ByXY []Coord

func (a ByXY) Len() int      { return len(a) }
func (a ByXY) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByXY) Less(i, j int) bool {
	return a[i].Y < a[j].Y || (a[i].Y == a[j].Y && a[i].X < a[j].X)
}

type HexGrid[T any] struct {
	xdim  int
	ydim  int
	hexes []T
}

func (g HexGrid[T]) Dims() (int, int) {
	return g.xdim, g.ydim
}

func Generate[T any](xdim int, ydim int, initFn func(coord Coord) T) HexGrid[T] {
	hexes := make([]T, 0, xdim*ydim)
	for y := 0; y < ydim; y++ {
		for x := 0; x < xdim; x++ {
			val := initFn(Coord{x, y})
			hexes = append(hexes, val)
		}
	}
	return HexGrid[T]{
		xdim,
		ydim,
		hexes,
	}
}

func (g HexGrid[T]) MapHexes(hexFn func(coord Coord, hexData *T)) {
	for y := 0; y < g.ydim; y++ {
		for x := 0; x < g.xdim; x++ {
			hexFn(Coord{x, y}, &g.hexes[g.xyToIndex(x, y)])
		}
	}
}

func (g HexGrid[T]) xyToIndex(x int, y int) int {
	return y*g.xdim + x
}

func (g HexGrid[T]) coordToIndex(coord Coord) int {
	return g.xyToIndex(coord.X, coord.Y)
}

func (g HexGrid[T]) indexToCoord(i int) Coord {
	return Coord{i % g.xdim, i / g.xdim}
}

func (g HexGrid[T]) isInBounds(coord Coord) bool {
	if coord.X < 0 || coord.X >= g.xdim {
		return false
	}

	if coord.Y < 0 || coord.Y >= g.ydim {
		return false
	}

	return true
}

func (g HexGrid[T]) GetAt(coord Coord) *T {
	return g.GetAtXY(coord.X, coord.Y)
}

func (g HexGrid[T]) GetAtXY(x int, y int) *T {
	if x < 0 || x >= g.xdim || y < 0 || y >= g.ydim {
		return nil
	}

	index := g.xyToIndex(x, y)
	return &g.hexes[index]
}

func (g HexGrid[T]) GetNeighbors(coord Coord) []Coord {
	return g.GetNeighborsXY(coord.X, coord.Y)
}

type offset struct {
	x int
	y int
}

func (g HexGrid[T]) GetNeighborsXY(x int, y int) []Coord {

	var offsets []offset
	if y%2 == 1 {
		offsets = []offset{
			{+0, -1},
			{+1, -1},
			{+1, +0},
			{+1, +1},
			{+0, +1},
			{-1, +0},
		}
	} else {
		offsets = []offset{
			{-1, -1},
			{+0, -1},
			{+1, +0},
			{+0, +1},
			{-1, +1},
			{-1, +0},
		}
	}

	result := make([]Coord, 0, len(offsets))
	for _, offset := range offsets {
		coord := Coord{x + offset.x, y + offset.y}
		if g.isInBounds(coord) {
			result = append(result, coord)
		}
	}
	return result
}
