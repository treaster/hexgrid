package hexgrid

type RangeResult struct {
	Coord Coord
	Dist  float64
}

type ByDist []RangeResult

func (a ByDist) Len() int      { return len(a) }
func (a ByDist) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDist) Less(i, j int) bool {
	if a[i].Dist != a[j].Dist {
		return a[i].Dist < a[j].Dist
	}

	if a[i].Coord.X != a[j].Coord.X {
		return a[i].Coord.X < a[j].Coord.X
	}

	if a[i].Coord.Y != a[j].Coord.Y {
		return a[i].Coord.Y < a[j].Coord.Y
	}

	return false
}

func (g HexGrid[T]) FindInRange(
	start Coord,
	maxRange float64,
	includeSrc bool,
	costFn func(hexA *T, rawHexB *T) float64,
) []RangeResult {

	visitSet := map[Coord]float64{
		start: 0.0,
	}

	queue := []Coord{start}
	result := []RangeResult{}
	for i := 0; i < len(queue); i++ {
		candidate := queue[i]
		candidateCost := visitSet[candidate]
		if candidateCost > maxRange {
			continue
		}

		if i != 0 || includeSrc {
			result = append(result, RangeResult{candidate, candidateCost})
		}

		neighbors := g.GetNeighbors(candidate)
		for _, neighbor := range neighbors {
			_, isVisited := visitSet[neighbor]
			if isVisited {
				continue
			}

			stepCost := costFn(g.GetAt(candidate), g.GetAt(neighbor))
			if stepCost < 0 {
				continue
			}

			queue = append(queue, neighbor)

			totalCost := candidateCost + stepCost
			visitSet[neighbor] = totalCost
		}
	}

	return result
}
