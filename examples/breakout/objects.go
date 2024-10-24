package breakout

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

// Cell represents a single cell in the Game of Life
type Cell struct {
	GameObject
	Alive bool
}
