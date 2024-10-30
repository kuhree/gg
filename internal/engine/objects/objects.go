
package objects

// Vector2D represents a 2D position or velocity
type Vector2D struct {
	X, Y float64
}


// GameObject represents a basic game entity
type GameObject struct {
	Position Vector2D
	Height   float64
	Width    float64
}

