package render

import (
	"fmt"
	"strings"
)

// Renderer handles the ASCII rendering for the game
type Renderer struct {
	width  int
	height int
	buffer [][]rune
}

// NewRenderer creates a new Renderer with the specified dimensions
func NewRenderer(width, height int) *Renderer {
	buffer := make([][]rune, height)
	for i := range buffer {
		buffer[i] = make([]rune, width)
	}

	return &Renderer{
		width:  width,
		height: height,
		buffer: buffer,
	}
}

// Size returns the height and width of the canvas
func (r *Renderer) Size() (int, int ) {
	return r.width, r.height
}

// Clear clears the render buffer
func (r *Renderer) Clear() {
	for y := 0; y < r.height; y++ {
		for x := 0; x < r.width; x++ {
			r.buffer[y][x] = ' '
		}
	}
}

// DrawChar draws a character at the specified position
func (r *Renderer) DrawChar(char rune, x, y int) {
	if x >= 0 && x < r.width && y >= 0 && y < r.height {
		r.buffer[y][x] = char
	}
}

// DrawText draws a string of text at the specified position
func (r *Renderer) DrawText(text string, x, y int) {
	for i, char := range text {
		r.DrawChar(char, x+i, y)
	}
}

// DrawRect draws a rectangle with the specified dimensions
func (r *Renderer) DrawRect(x, y, width, height int, char rune) {
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			r.DrawChar(char, x+dx, y+dy)
		}
	}
}

// Render outputs the current buffer to the console
func (r *Renderer) Render() {
	var sb strings.Builder
	for y := 0; y < r.height; y++ {
		for x := 0; x < r.width; x++ {
			sb.WriteRune(r.buffer[y][x])
		}
		sb.WriteRune('\n')
	}
	fmt.Print("\033[H\033[2J") // Clear the console
	fmt.Print(sb.String())
}

// GetDimensions returns the width and height of the renderer
func (r *Renderer) GetDimensions() (int, int) {
	return r.width, r.height
}
