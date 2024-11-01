package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/MichaelTJones/pcg"
)

// Make and return a zero-initialized 1D bool slice to represent a 2D grid
func makeGrid(width, height uint) []bool {
	return make([]bool, width*height)
}

// Make and return a random-initialized 1D bool slice to represent a 2D grid
func makeRandomBoolGrid(width, height uint) []bool {
	grid := makeGrid(width, height)

	pcg32 := pcg.NewPCG32()
	pcg32.Seed(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano())*2-47)

	for i := uint(0); i < width*height; i++ {
		grid[i] = pcg32.Random()&0x01 == 0
	}

	return grid
}

// Wrap an int around 0 and max
func wrapIndex(n int, max uint) uint {
	if n < 0 {
		return max - 1
	} else if n > int(max-1) {
		return 0
	} else {
		return uint(n)
	}
}

// Make and return a 1D uint8 slice to represent a 2D game of lif grid's living neighbours
func makeLiveNeighboursGrid(sourceGrid *[]bool, width, height uint) []uint8 {
	liveNeighboursGrid := make([]uint8, width*height)

	offsets := [8][2]int{
		{-1, -1}, {0, -1}, {1, -1},
		{-1, 0}, {1, 0},
		{-1, 1}, {0, 1}, {1, 1},
	}

	for y := uint(0); y < height; y++ {
		for x := uint(0); x < width; x++ {
			var liveNeighbours uint8 = 0

			for _, direction := range offsets {
				if (*sourceGrid)[wrapIndex(int(y)+direction[1], height)*width+wrapIndex(int(x)+direction[0], width)] {
					liveNeighbours++
				}
			}

			liveNeighboursGrid[y*width+x] = liveNeighbours
		}
	}

	return liveNeighboursGrid
}

// Print a 2D bool grid to the termina in ASCII
func displayGrid(grid *[]bool, width, height uint) {
	fmt.Println("\033[48;2;127;0;255m" + strings.Repeat(" ", int(width*2+4)) + "\033[0m")

	for y := uint(0); y < height; y++ {
		row := "\033[48;2;127;0;255m  "
		for x := uint(0); x < width; x++ {
			if (*grid)[y*width+x] == true {
				row += "\033[48;2;255;255;255m  "
			} else {
				row += "\033[48;2;0;0;0m  "
			}
		}
		row += "\033[48;2;127;0;255m  \033[0m"
		fmt.Println(row)
	}

	fmt.Println("\033[48;2;127;0;255m" + strings.Repeat(" ", int(width*2+4)) + "\033[0m")
}

// Display the live neighbours grid, for debugging
func displayLiveNeighboursGrid(liveNeighboursGrid *[]uint8, width, height uint) {
	for y := uint(0); y < height; y++ {
		row := ""
		for x := uint(0); x < width; x++ {
			row += fmt.Sprint((*liveNeighboursGrid)[y*width+x])
		}
		fmt.Println(row)
	}
	fmt.Println()
	fmt.Println()
}

// Advanced debugging display
func displayDebugging(grid *[]bool, liveNeighboursGrid *DiffGrid, width, height uint) {
	fmt.Println("\033[48;2;127;0;255m" + strings.Repeat(" ", int(width*2+4)) + "\033[0m")

	for y := uint(0); y < height; y++ {
		row := "\033[48;2;127;0;255m  "
		for x := uint(0); x < width; x++ {
			liveNeighbours := liveNeighboursGrid.get(y*width + x)
			var countCharacter string
			if liveNeighbours == 0 {
				countCharacter = " "
			} else if liveNeighbours <= 8 {
				countCharacter = string(rune(liveNeighbours + 48))
			} else {
				countCharacter = "E"
			}

			cellState := (*grid)[y*width+x]

			var countColour string

			if cellState == true && (liveNeighbours != 2 && liveNeighbours != 3) {
				countColour = "\033[38;2;255;0;0m"
			} else if cellState == false && liveNeighbours == 3 {
				countColour = "\033[38;2;0;255;0m"
			} else {
				countColour = "\033[38;2;127;127;127m"
			}

			if cellState == true {
				row += "\033[48;2;255;255;255m" + countColour + " " + countCharacter
			} else {
				row += "\033[48;2;0;0;0m" + countColour + " " + countCharacter
			}
		}
		row += "\033[48;2;127;0;255m  \033[0m"
		fmt.Println(row)
	}

	fmt.Println("\033[48;2;127;0;255m" + strings.Repeat(" ", int(width*2+4)) + "\033[0m")
}

// Fill newGrid with the next game of life iteration using oldGrid as the previous iteration
func iterateGrids(oldGrid *[]bool, newGrid *[]bool, liveNeighboursGrid *DiffGrid, width, height uint) {
	offsets := [8][2]int{
		{-1, -1}, {0, -1}, {1, -1},
		{-1, 0}, {1, 0},
		{-1, 1}, {0, 1}, {1, 1},
	}

	for y := uint(0); y < height; y++ {
		for x := uint(0); x < width; x++ {
			liveNeighbours := liveNeighboursGrid.getOld(y*width + x)

			oldValue := (*oldGrid)[y*width+x]
			newValue := liveNeighbours == 3 || (oldValue == true && liveNeighbours == 2)

			(*newGrid)[y*width+x] = newValue

			if newValue != oldValue {
				for _, direction := range offsets {
					neighbourY := wrapIndex(int(y)+direction[1], height)
					neighbourX := wrapIndex(int(x)+direction[0], width)

					liveNeighboursGrid.change(neighbourY*width+neighbourX, newValue)
				}
			}
		}
	}

	liveNeighboursGrid.merge()
}

// Type to track changes made to an array within one diff period, then merge then for the next diff period.
type DiffGrid struct {
	grid []uint8
	diff []int8
}

// Get val
func (grid *DiffGrid) getOld(index uint) uint8 {
	return grid.grid[index]
}
func (grid *DiffGrid) get(index uint) uint8 {
	return grid.grid[index] + uint8(grid.diff[index])
}

// Change value at index by 1 if upOrDown is true, else -1
func (grid *DiffGrid) change(index uint, upOrDown bool) {
	if upOrDown {
		grid.diff[index] += 1
	} else {
		grid.diff[index] -= 1
	}
}

// Merge diff into grid and zero the diff.
func (grid *DiffGrid) merge() {
	for i := 0; i < len(grid.grid); i++ {
		grid.grid[i] += uint8(grid.diff[i])
		grid.diff[i] = 0
	}
}

func main() {
	fmt.Println("Game of life program starting.")

	var width, height uint

	fmt.Println("Enter width and height to use:")
	fmt.Print("width: ")
	fmt.Scan(&width)
	fmt.Print("height: ")
	fmt.Scan(&height)

	grid1 := makeRandomBoolGrid(width, height)
	grid2 := makeGrid(width, height)

	mainGrid := &grid1
	oldGrid := &grid2

	liveNeighboursGrid := DiffGrid{
		makeLiveNeighboursGrid(mainGrid, width, height),
		make([]int8, width*height),
	}

	for i := uint(0); i < 10000; i++ {
		if i > 0 {
			fmt.Printf("\033[%dA\033[0J", height+2)
		}
		displayGrid(mainGrid, width, height)
		//displayLiveNeighboursGrid(&liveNeighboursGrid, width, height)
		//displayDebugging(mainGrid, &liveNeighboursGrid, width, height)

		mainGrid, oldGrid = oldGrid, mainGrid

		iterateGrids(oldGrid, mainGrid, &liveNeighboursGrid, width, height)

		time.Sleep(67 * time.Millisecond)
	}
}
