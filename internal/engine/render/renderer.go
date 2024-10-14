package render

import (
	"errors"
	"fmt"
	"strings"
)

type Color int

const (
	ColorBlack Color = iota
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
	ColorBrightBlack
	ColorBrightRed
	ColorBrightGreen
	ColorBrightYellow
	ColorBrightBlue
	ColorBrightMagenta
	ColorBrightCyan
	ColorBrightWhite
)

type ColorInfo struct {
	Name string
	ANSI string
}

type Palette struct {
	Colors []ColorInfo
}

var DefaultPalette = Palette{
	Colors: []ColorInfo{
		ColorBlack:         {"black", "\033[30m"},
		ColorRed:           {"red", "\033[31m"},
		ColorGreen:         {"green", "\033[32m"},
		ColorYellow:        {"yellow", "\033[33m"},
		ColorBlue:          {"blue", "\033[34m"},
		ColorMagenta:       {"magenta", "\033[35m"},
		ColorCyan:          {"cyan", "\033[36m"},
		ColorWhite:         {"white", "\033[37m"},
		ColorBrightBlack:   {"bright_black", "\033[90m"},
		ColorBrightRed:     {"bright_red", "\033[91m"},
		ColorBrightGreen:   {"bright_green", "\033[92m"},
		ColorBrightYellow:  {"bright_yellow", "\033[93m"},
		ColorBrightBlue:    {"bright_blue", "\033[94m"},
		ColorBrightMagenta: {"bright_magenta", "\033[95m"},
		ColorBrightCyan:    {"bright_cyan", "\033[96m"},
		ColorBrightWhite:   {"bright_white", "\033[97m"},
	},
}

const (
	// Block Elements
	FullBlock       = '█'
	LightShade      = '░'
	MediumShade     = '▒'
	DarkShade       = '▓'
	UpperHalfBlock  = '▀'
	LowerHalfBlock  = '▄'
	LeftHalfBlock   = '▌'
	RightHalfBlock  = '▐'
	QuadrantLowerLeft  = '▖'
	QuadrantLowerRight = '▗'
	QuadrantUpperLeft  = '▘'
	QuadrantUpperRight = '▝'

	// Box Drawing Characters
	LightHorizontal = '─'
	LightVertical   = '│'
	LightDownAndRight = '┌'
	LightDownAndLeft  = '┐'
	LightUpAndRight   = '└'
	LightUpAndLeft    = '┘'
	LightVerticalAndRight = '├'
	LightVerticalAndLeft  = '┤'
	LightHorizontalAndDown = '┬'
	LightHorizontalAndUp   = '┴'
	LightCross = '┼'

	// Geometric Shapes
	BlackCircle = '●'
	BlackDot = '•'
	WhiteCircle = '○'
	BlackSquare = '■'
	WhiteSquare = '□'
	BlackTriangle = '▲'
	WhiteTriangle = '△'

	// Arrows
	LeftArrow  = '←'
	UpArrow    = '↑'
	RightArrow = '→'
	DownArrow  = '↓'
)

// Renderer handles the ASCII rendering for the game
type Renderer struct {
	width   int
	height  int
	buffer  [][]rune
	colors  [][]Color
	palette Palette
}

// NewRenderer creates a new Renderer with the specified dimensions
func NewRenderer(width, height int) *Renderer {
	buffer := make([][]rune, height)
	colors := make([][]Color, height)
	for i := range buffer {
		buffer[i] = make([]rune, width)
		colors[i] = make([]Color, width)
		for j := range buffer[i] {
			buffer[i][j] = ' '
			colors[i][j] = ColorBlack // Default color
		}
	}

	return &Renderer{
		width:   width,
		height:  height,
		buffer:  buffer,
		colors:  colors,
		palette: DefaultPalette,
	}
}

// Size returns the width and height of the canvas
func (r *Renderer) Size() (int, int) {
	return r.width, r.height
}

// Clear clears the render buffer
func (r *Renderer) Clear() {
	for y := 0; y < r.height; y++ {
		for x := 0; x < r.width; x++ {
			r.buffer[y][x] = ' '
			r.colors[y][x] = ColorBlack
		}
	}
}

// DrawPixel draws a pixel at the specified position
func (r *Renderer) DrawPixel(x, y int, color Color) error {
	return r.DrawChar(FullBlock, x, y, color)
}

// DrawChar draws a character at the specified position
func (r *Renderer) DrawChar(char rune, x, y int, color Color) error {
	if x < 0 || x >= r.width || y < 0 || y >= r.height {
		return errors.New("drawing outside buffer bounds")
	}
	r.buffer[y][x] = char
	r.colors[y][x] = color
	return nil
}

// DrawText draws a string of text at the specified position
func (r *Renderer) DrawText(text string, x, y int, color Color) error {
	for i, char := range text {
		if err := r.DrawChar(char, x+i, y, color); err != nil {
			return err
		}
	}
	return nil
}

// DrawRect draws a rectangle with the specified dimensions
func (r *Renderer) DrawRect(x, y, width, height int, char rune, color Color) error {
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			if err := r.DrawChar(char, x+dx, y+dy, color); err != nil {
				return err
			}
		}
	}
	return nil
}

// Render outputs the current buffer to the console
func (r *Renderer) Render() {
	var sb strings.Builder
	sb.Grow(r.width * r.height * 20) // Estimate capacity

	currentColor := ColorBlack
	for y := 0; y < r.height; y++ {
		sb.WriteString("\033[0m") // Reset color 

		for x := 0; x < r.width; x++ {
			if r.colors[y][x] != currentColor {
				currentColor = r.colors[y][x]
				sb.WriteString(r.palette.Colors[currentColor].ANSI)
			}
			sb.WriteRune(r.buffer[y][x])
		}
		sb.WriteRune('\n')
	}
	fmt.Print("\033[H\033[2J") // Clear the console
	fmt.Print("\033[H") // Move cursor to top-left
	fmt.Print(sb.String())
}
