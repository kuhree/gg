package gameoflife

import "github.com/kuhree/gg/internal/engine/objects"

// Cell represents a single cell in the Game of Life
type Cell struct {
	objects.GameObject
	Alive bool
}
