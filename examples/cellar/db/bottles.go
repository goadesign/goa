package db

import (
	"fmt"

	"github.com/raphael/goa/examples/cellar/autogen/app"
)

// In-memory "database"
var data map[int][]*app.Bottle

// BottleHref computes the bottle href from its id.
func BottleHref(account, id int) string {
	return fmt.Sprintf("/api/%d/bottles/%d", account, id)
}

// Initialize "database" with dummy data
func init() {
	data = map[int][]*app.Bottle{
		1: []*app.Bottle{
			&app.Bottle{
				ID:        100,
				AccountID: 1,
				Href:      BottleHref(1, 100),
				Name:      "Number 8",
				Vineyard:  "Asti Winery",
				Varietal:  "Merlot",
				Vintage:   2012,
				Color:     "red",
				Sweet:     false,
				Country:   "USA",
				Region:    "California",
				Review:    "Great value",
			},
			&app.Bottle{
				ID:        101,
				AccountID: 1,
				Href:      BottleHref(1, 101),
				Name:      "Mourvedre",
				Vineyard:  "Rideau",
				Varietal:  "Mourvedre",
				Vintage:   2012,
				Color:     "red",
				Sweet:     false,
				Country:   "USA",
				Region:    "California",
				Review:    "Good but expensive",
			},
			&app.Bottle{
				ID:        102,
				AccountID: 1,
				Href:      BottleHref(1, 102),
				Name:      "Blue's Cuvee",
				Vineyard:  "Longoria",
				Varietal:  "Cabernet Franc with Merlot, Malbec, Cabernet Sauvignon and Syrah",
				Vintage:   2012,
				Color:     "red",
				Sweet:     false,
				Country:   "USA",
				Region:    "California",
				Review:    "Favorite",
			},
		},
		2: []*app.Bottle{
			&app.Bottle{
				ID:        200,
				AccountID: 2,
				Href:      BottleHref(42, 200),
				Name:      "Blackstone Merlot",
				Vineyard:  "Blackstone",
				Varietal:  "Merlot",
				Vintage:   2012,
				Color:     "red",
				Sweet:     false,
				Country:   "USA",
				Region:    "California",
				Review:    "OK",
			},
			&app.Bottle{
				ID:        201,
				AccountID: 2,
				Href:      BottleHref(42, 201),
				Name:      "Wild Horse",
				Vineyard:  "Wild Horse",
				Varietal:  "Pinot Noir",
				Vintage:   2012,
				Color:     "red",
				Sweet:     false,
				Country:   "USA",
				Region:    "California",
				Review:    "Solid Pinot",
			},
		},
	}
}

// GetBottle returns the bottle with the given id from the given account or nil if not found.
func GetBottle(account, id int) *app.Bottle {
	bottles, ok := data[account]
	if !ok {
		return nil
	}
	for _, b := range bottles {
		if b.ID == id {
			return b
		}
	}
	return nil
}

// GetBottles return the bottles from the given account.
func GetBottles(account int) ([]*app.Bottle, error) {
	bottles, ok := data[account]
	if !ok {
		return nil, fmt.Errorf("unknown account %d", account)
	}
	return bottles, nil
}

// GetBottlesByYears returns the bottles with the vintage in the given array from the given account.
func GetBottlesByYears(account int, years []int) ([]*app.Bottle, error) {
	bottles, ok := data[account]
	if !ok {
		return nil, fmt.Errorf("unknown account %d", account)
	}
	var res []*app.Bottle
	for _, b := range bottles {
		selected := false
		for _, y := range years {
			if y == b.Vintage {
				selected = true
				break
			}
		}
		if selected {
			res = append(res, b)
		}
	}
	return res, nil
}

// NewBottle creates a new bottle resource.
func NewBottle(account int) *app.Bottle {
	bottles, _ := data[account]
	newID := 1
	taken := true
	for ; taken; newID++ {
		taken = false
		for _, b := range bottles {
			if b.ID == newID {
				taken = true
				break
			}
		}
	}
	bottle := app.Bottle{ID: newID}
	data[newID] = append(data[newID], &bottle)
	return &bottle
}

// Save persists bottle to database.
func Save(b *app.Bottle) {
	data[b.AccountID] = append(data[b.AccountID], b)
}

// Delete deletes bottle from database.
func Delete(bottle *app.Bottle) {
	if bottle == nil {
		return
	}
	account := bottle.AccountID
	id := bottle.ID
	if bs, ok := data[account]; ok {
		idx := -1
		for i, b := range bs {
			if b.ID == id {
				idx = i
				break
			}
		}
		if idx > -1 {
			bs = append(bs[:idx], bs[idx+1:]...)
			data[account] = bs
		}
	}
}
