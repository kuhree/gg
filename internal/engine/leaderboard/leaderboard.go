package leaderboard

import (
	"encoding/json"
	"os"
	"sort"

	"github.com/kuhree/gg/internal/utils"
)

type Record struct {
	Name    string
	Score   int
	Details string
}

type Board struct {
	Records []Record
}

func NewBoard(filename string) (*Board, error) {
	board := &Board{}

	err := board.Load(filename)
	if err != nil {
		if os.IsNotExist(err) {
			board = &Board{}
		}

		return board, nil
	}

	return board, nil
}

func (b *Board) Add(name string, score int, notes string) {
	if b.Records == nil {
		b.Records = make([]Record, 0)
	}

	b.Records = append(b.Records, Record{Name: name, Score: score, Details: notes})
	sort.Slice(b.Records, func(i, j int) bool {
		return b.Records[i].Score > b.Records[j].Score
	})
	if len(b.Records) > 10 {
		b.Records = b.Records[:10]
	}
}

func (b *Board) Save(filename string) error {
	if len(b.Records) <= 0 {
		return nil
	} else if err := utils.EnsureDir(filename); err != nil {
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
	if err := utils.EnsureDir(filename); err != nil {
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
		n = len(b.Records)
	}

	return b.Records[:n]
}
