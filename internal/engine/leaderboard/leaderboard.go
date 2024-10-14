package leaderboard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

type Record struct {
	Name  string
	Score int
	Notes string
}

type Board struct {
	Records []Record
}

func NewBoard(filename string) (*Board, error) {
	board := &Board{}

	err := board.Load(filename)
	if err != nil {
		if os.IsNotExist(err) {
			board = &Board{
				Records: make([]Record, 0),
			}
		}
		return nil, err
	}

	return board, nil
}

func (b *Board) Add(name string, score int, notes string) {
	b.Records = append(b.Records, Record{Name: name, Score: score, Notes: notes})
	sort.Slice(b.Records, func(i, j int) bool {
		return b.Records[i].Score > b.Records[j].Score
	})
	if len(b.Records) > 10 {
		b.Records = b.Records[:10]
	}
}

func ensureDir(filename string) error {
	dir := filepath.Dir(filename)
	return os.MkdirAll(dir, 0755)
}

func (b *Board) Save(filename string) error {
	if len(b.Records) <= 0 {
		return nil
	} else if err := ensureDir(filename); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(b)
}

func (b *Board) Load(filename string) error {
	if err := ensureDir(filename); err != nil {
		return err
	}
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewDecoder(file).Decode(b)
}

func (b *Board) TopScores(n int) []Record {
	sort.Slice(b.Records, func(i, j int) bool {
		return b.Records[i].Score > b.Records[j].Score
	})

	if n > len(b.Records) {
		n = len(b.Records) - 1
	}

	return b.Records[:n]
}
