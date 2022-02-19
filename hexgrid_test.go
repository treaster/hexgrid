package hexgrid_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/treaster/hexgrid"
	"github.com/treaster/testutil"
)

func countNeighbors(hexes []hexgrid.Coord) int {
	count := 0
	for _, _ = range hexes {
		count++
	}
	return count
}

func noopInitFn(coord hexgrid.Coord) int {
	return 0
}

func TestGenerate(t *testing.T) {
	expectedHexes := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

	count := 0
	grid := hexgrid.Generate[int](3, 3, func(hexgrid.Coord) int {
		count++
		return count
	})

	foundHexes := []int{}
	grid.MapHexes(func(coord hexgrid.Coord, hexData *int) {
		foundHexes = append(foundHexes, *hexData)

		fetchedInt := grid.GetAt(coord)
		require.Equal(t, *hexData, *fetchedInt)

		*hexData++
		fetchedInt = grid.GetAt(coord)
		require.Equal(t, *hexData, *fetchedInt)
	})
	require.Equal(t, expectedHexes, foundHexes)

}

func TestGetNeighbors(t *testing.T) {
	testCases := []struct {
		gridX          int
		gridY          int
		expectedCounts [][]int
	}{
		{
			3, 3,
			[][]int{
				{2, 5, 2},
				{4, 6, 4},
				{3, 3, 3},
			},
		},
		{
			3, 2,
			[][]int{
				{2, 3},
				{4, 4},
				{3, 2},
			},
		},
		{
			2, 4,
			[][]int{
				// This will display as two columns of four hexes each.
				{2, 5, 3, 3},
				{3, 3, 5, 2},
			},
		},
	}

	for testI, testCase := range testCases {
		grid := hexgrid.Generate[int](testCase.gridX, testCase.gridY, noopInitFn)
		grid.MapHexes(func(coord hexgrid.Coord, _ *int) {
			n := grid.GetNeighbors(coord)
			nonNilCount := countNeighbors(n)
			require.Equal(t, testCase.expectedCounts[coord.X][coord.Y], nonNilCount, fmt.Sprintf("Test case %d: %v", testI, coord))
		})
	}
}

func TestGetNeighborsSpecific(t *testing.T) {
	grid := hexgrid.Generate(3, 3, noopInitFn)
	neighbors := grid.GetNeighborsXY(2, 1)
	require.Equal(t, 3, len(neighbors))
}

func TestGetNeighborsWithId(t *testing.T) {
	count := 0
	grid := hexgrid.Generate(3, 4, func(coord hexgrid.Coord) int {
		count++
		return count
	})

	testCases := []struct {
		label       string
		coord       hexgrid.Coord
		expectedIds []int
	}{
		{
			"odd row",
			hexgrid.Coord{1, 1},
			[]int{2, 3, 6, 9, 8, 4},
		},
		{
			"even row",
			hexgrid.Coord{1, 2},
			[]int{4, 5, 9, 11, 10, 7},
		},
		{
			"upper left corner",
			hexgrid.Coord{0, 0},
			[]int{2, 4},
		},
		{
			"lower right corner",
			hexgrid.Coord{2, 3},
			[]int{9, 11},
		},
	}

	for _, testCase := range testCases {
		neighbors := grid.GetNeighbors(testCase.coord)
		neighborIds := []int{}
		for _, neighborCoord := range neighbors {
			hexData := grid.GetAt(neighborCoord)
			neighborIds = append(neighborIds, *hexData)
		}

		require.Equal(t, testCase.expectedIds, neighborIds)
	}
}

func TestFindPath(t *testing.T) {
	testCases := []struct {
		label        string
		start        hexgrid.Coord
		end          hexgrid.Coord
		nodeCosts    [][]float64
		expectedCost float64
		expectedPath []hexgrid.Coord
	}{
		{
			"Simple case of uniform costs",
			hexgrid.Coord{1, 0},
			hexgrid.Coord{1, 3},
			[][]float64{
				{1, 1, 1, 1},
				{1, 1, 1, 1},
				{1, 1, 1, 1},
				{1, 1, 1, 1},
			},
			3,
			[]hexgrid.Coord{
				{1, 0},
				{1, 1},
				{1, 2},
				{1, 3},
			},
		},
		{
			"Must navigate a circuitous path",
			hexgrid.Coord{0, 0},
			hexgrid.Coord{1, 2},
			[][]float64{
				{1, 1, 1, 1},
				{50, 50, 50, 1},
				{50, 1, 50, 1},
				{50, 1, 1, 1},
			},
			8,
			[]hexgrid.Coord{
				{0, 0},
				{1, 0},
				{2, 0},
				{3, 0},
				{3, 1},
				{3, 2},
				{2, 3},
				{1, 3},
				{1, 2},
			},
		},

		{
			"Verify that 0's are legal parts of the path",
			hexgrid.Coord{0, 0},
			hexgrid.Coord{1, 2},
			[][]float64{
				{1, 1, 1, 1},
				{0, 1, 1, 1},
				{0, 1, 0, 1},
				{0, 1, 1, 1},
			},
			1,
			[]hexgrid.Coord{
				{0, 0},
				{0, 1},
				{1, 2},
			},
		},

		{
			"Verify that negative numbers are impassable",
			hexgrid.Coord{0, 0},
			hexgrid.Coord{1, 2},
			[][]float64{
				{1, 1, 1, 1},
				{-1, -1, -1, 1},
				{-1, 1, -1, 1},
				{-1, 1, 1, 1},
			},
			8,
			[]hexgrid.Coord{
				{0, 0},
				{1, 0},
				{2, 0},
				{3, 0},
				{3, 1},
				{3, 2},
				{2, 3},
				{1, 3},
				{1, 2},
			},
		},

		{
			"Verify that we'll cut through medium-expense nodes when necessary to produce a lower cost.",
			hexgrid.Coord{0, 0},
			hexgrid.Coord{1, 2},
			[][]float64{
				{1, 1, 1, 1},
				{3, 3, 3, 1},
				{3, 1, 3, 1},
				{3, 1, 1, 1},
			},
			4,
			[]hexgrid.Coord{
				{0, 0},
				{0, 1},
				{1, 2},
			},
		},

		{
			"No available path",
			hexgrid.Coord{0, 0},
			hexgrid.Coord{1, 2},
			[][]float64{
				{1, 1, 1, 1},
				{-1, -1, -1, -1},
				{3, 1, 3, 1},
				{3, 1, 1, 1},
			},
			0,
			[]hexgrid.Coord{},
		},
	}

	for testI, testCase := range testCases {
		grid := hexgrid.Generate(4, 4, func(coord hexgrid.Coord) float64 {
			return testCase.nodeCosts[coord.Y][coord.X]
		})

		costFn := func(hexA *float64, hexB *float64) float64 {
			return *hexB
		}
		outputCost, outputPath, err := grid.FindPath(testCase.start, testCase.end, costFn)
		if len(testCase.expectedPath) == 0 {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, testCase.expectedCost, outputCost, fmt.Sprintf("Test case %d: %s", testI, testCase.label))
			require.Equal(t, testCase.expectedPath, outputPath, fmt.Sprintf("Test case %d: %s", testI, testCase.label))
		}
	}
}

func TestFindInRange(t *testing.T) {
	testCases := []struct {
		label           string
		nodeCosts       [][]float64
		start           hexgrid.Coord
		maxRange        float64
		includeSrc      bool
		expectedResults []hexgrid.RangeResult
	}{
		{
			"Simple case of uniform costs",
			[][]float64{
				{1, 1, 1, 1},
				{1, 1, 1, 1},
				{1, 1, 1, 1},
				{1, 1, 1, 1},
			},
			hexgrid.Coord{1, 0},
			2,
			false,
			[]hexgrid.RangeResult{
				{hexgrid.Coord{0, 0}, 1},
				{hexgrid.Coord{2, 0}, 1},
				{hexgrid.Coord{0, 1}, 1},
				{hexgrid.Coord{1, 1}, 1},
				{hexgrid.Coord{0, 2}, 2},
				{hexgrid.Coord{1, 2}, 2},
				{hexgrid.Coord{2, 2}, 2},
				{hexgrid.Coord{2, 1}, 2},
				{hexgrid.Coord{3, 0}, 2},
			},
		},
		{
			"Simple case of uniform costs, with include source",
			[][]float64{
				{1, 1, 1, 1},
				{1, 1, 1, 1},
				{1, 1, 1, 1},
				{1, 1, 1, 1},
			},
			hexgrid.Coord{1, 0},
			2,
			true,
			[]hexgrid.RangeResult{
				{hexgrid.Coord{1, 0}, 0},
				{hexgrid.Coord{0, 0}, 1},
				{hexgrid.Coord{2, 0}, 1},
				{hexgrid.Coord{0, 1}, 1},
				{hexgrid.Coord{1, 1}, 1},
				{hexgrid.Coord{0, 2}, 2},
				{hexgrid.Coord{1, 2}, 2},
				{hexgrid.Coord{2, 2}, 2},
				{hexgrid.Coord{2, 1}, 2},
				{hexgrid.Coord{3, 0}, 2},
			},
		},
		{
			"Nonuniform costs",
			[][]float64{
				{1, 1, 5, 1},
				{9, 1, 5, 1},
				{1, 1, 3, 1},
				{9, 9, 1, 1},
			},
			hexgrid.Coord{0, 2},
			5,
			false,
			[]hexgrid.RangeResult{
				{hexgrid.Coord{1, 2}, 1},
				{hexgrid.Coord{2, 2}, 4},
				{hexgrid.Coord{3, 2}, 5},
				{hexgrid.Coord{2, 3}, 5},
				{hexgrid.Coord{1, 1}, 2},
				{hexgrid.Coord{1, 0}, 3},
				{hexgrid.Coord{0, 0}, 4},
			},
		},
		{
			"Infinite costs",
			[][]float64{
				{-1, -1, -1, -1},
				{1, 1, 1, 1},
				{-1, -1, -1, -1},
				{-1, -1, -1, -1},
			},
			hexgrid.Coord{0, 1},
			500,
			false,
			[]hexgrid.RangeResult{
				{hexgrid.Coord{1, 1}, 1},
				{hexgrid.Coord{2, 1}, 2},
				{hexgrid.Coord{3, 1}, 3},
			},
		},
	}

	for testI, testCase := range testCases {
		grid := hexgrid.Generate(4, 4, func(coord hexgrid.Coord) float64 {
			return testCase.nodeCosts[coord.Y][coord.X]
		})

		costFn := func(hexA *float64, hexB *float64) float64 {
			return *hexB
		}
		results := grid.FindInRange(testCase.start, testCase.maxRange, testCase.includeSrc, costFn)
		sort.Sort(hexgrid.ByDist(testCase.expectedResults))
		sort.Sort(hexgrid.ByDist(results))
		require.Equal(t, testCase.expectedResults, results, fmt.Sprintf("Test case %d: %s", testI, testCase.label))
	}
}

func TestCoordJson(t *testing.T) {
	coord := hexgrid.Coord{
		X: 10,
		Y: 5,
	}

	jsonBytes := testutil.MustMarshalJson(t, coord)
	require.Equal(t, `{"x":10,"y":5}`, string(jsonBytes))
}
