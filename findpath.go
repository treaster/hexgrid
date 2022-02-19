package hexgrid

import (
	"fmt"
	"math"
	"sort"
)

type costData struct {
	index        int
	distSoFar    float64
	estRemaining float64
	prevIndex    int
}

func (pc costData) estTotal() float64 {
	return pc.distSoFar + pc.estRemaining
}

type byScore []costData

func (a byScore) Len() int           { return len(a) }
func (a byScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byScore) Less(i, j int) bool { return a[i].estTotal() > a[j].estTotal() }

func estDistance(start Coord, end Coord) float64 {
	return math.Abs(float64(start.X-end.X)) + math.Abs(float64(start.Y-end.Y))
}

// Find a path through a hex grid.
// Cost function must be symmetric
// Costs must be positive. Nonpositive costs will be considered infinite.
func (g HexGrid[T]) FindPath(
	start Coord,
	end Coord,
	costFn func(hexA *T, rawHexB *T) float64) (
	float64, []Coord, error) {

	// reverse start and end so the final path is trivially in the right order.
	start, end = end, start

	realCosts := make([]costData, len(g.hexes))
	for i, _ := range realCosts {
		realCosts[i] = costData{i, -1, -1, -1}
	}

	startIndex := g.coordToIndex(start)
	endIndex := g.coordToIndex(end)

	// stack is a stack of cost estimates.
	stack := []costData{
		costData{startIndex, 0, estDistance(start, end), -1},
	}

	for len(stack) > 0 {
		stackSize := len(stack)
		parent := stack[stackSize-1]
		stack = stack[:stackSize-1]

		// If we've already processed a coord, the earlier one will definitely be
		// better than the later one.
		if realCosts[parent.index].distSoFar >= 0 {
			continue
		}

		// Mark this parent as visited.
		realCosts[parent.index] = parent

		if parent.index == endIndex {
			finalPath := []Coord{}
			node := realCosts[endIndex]
			finalCost := node.distSoFar
			for {
				finalPath = append(finalPath, g.indexToCoord(node.index))
				if node.prevIndex == -1 {
					break
				}
				node = realCosts[node.prevIndex]
			}

			return finalCost, finalPath, nil
		}

		parentCoord := g.indexToCoord(parent.index)
		neighbors := g.GetNeighbors(parentCoord)
		for _, neighbor := range neighbors {
			neighborIndex := g.coordToIndex(neighbor)

			// Don't add neighbors that are already locked in.
			if realCosts[neighborIndex].distSoFar >= 0 {
				continue
			}

			edgeCost := costFn(g.GetAt(parentCoord), g.GetAt(neighbor))

			// If edge cost is negative, we could get an infinite loop through no-cost nodes.
			if edgeCost < 0 {
				continue
			}

			newComponent := costData{
				neighborIndex,
				parent.distSoFar + edgeCost,
				estDistance(neighbor, end),
				parent.index,
			}
			stack = append(stack, newComponent)
		}

		sort.Sort(byScore(stack))
	}

	return 0.0, []Coord{}, fmt.Errorf("no path exists")
}
