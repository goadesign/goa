package db

import (
	"fmt"

	"github.com/raphael/goa/examples/cellar/app/autogen"
)

// In-memory "database"
var data map[int][]*autogen.BottleResource

// HREF from id
func BottleHref(account, id int) string {
	return fmt.Sprintf("/api/%d/bottles/%d", account, id)
}

// Initialize "database" with dummy data
func Init() {
	data = map[int][]*autogen.BottleResource{
		1: []*autogen.BottleResource{
			&autogen.BottleResource{
				ID:       100,
				Href:     BottleHref(1, 100),
				Name:     "Number 8",
				Vineyard: "Asti Winery",
				Varietal: "Merlot",
				Vintage:  2012,
				Color:    "red",
				Sweet:    false,
				Country:  "USA",
				Region:   "California",
				Review:   "Great value",
			},
			&autogen.BottleResource{
				ID:       101,
				Href:     BottleHref(1, 101),
				Name:     "Mourvedre",
				Vineyard: "Rideau",
				Varietal: "Mourvedre",
				Vintage:  2012,
				Color:    "red",
				Sweet:    false,
				Country:  "USA",
				Region:   "California",
				Review:   "Good but expensive",
			},
			&autogen.BottleResource{
				ID:       102,
				Href:     BottleHref(1, 102),
				Name:     "Blue's Cuvee",
				Vineyard: "Longoria",
				Varietal: "Cabernet Franc with Merlot, Malbec, Cabernet Sauvignon and Syrah",
				Vintage:  2012,
				Color:    "red",
				Sweet:    false,
				Country:  "USA",
				Region:   "California",
				Review:   "Favorite",
			},
		},
		2: []*autogen.BottleResource{
			&autogen.BottleResource{
				ID:       200,
				Href:     BottleHref(42, 200),
				Name:     "Blackstone Merlot",
				Vineyard: "Blackstone",
				Varietal: "Merlot",
				Vintage:  2012,
				Color:    "red",
				Sweet:    false,
				Country:  "USA",
				Region:   "California",
				Review:   "OK",
			},
			&autogen.BottleResource{
				ID:       201,
				Href:     BottleHref(42, 201),
				Name:     "Wild Horse",
				Vineyard: "Wild Horse",
				Varietal: "Pinot Noir",
				Vintage:  2012,
				Color:    "red",
				Sweet:    false,
				Country:  "USA",
				Region:   "California",
				Review:   "Solid Pinot",
			},
		},
	}
}

// Return bottle with given id from given account or nil if not found.
func GetBottle(account, id int) *autogen.BottleResource {
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

// Return bottles from given account.
func GetBottles(account int) ([]*autogen.BottleResource, error) {
	bottles, ok := data[account]
	if !ok {
		return nil, fmt.Errorf("unknown account %d", account)
	}
	return bottles, nil
}

// Return bottles with vintage in given array from given account.
func GetBottlesByYears(account int, years []int) ([]*autogen.BottleResource, error) {
	bottles, ok := data[account]
	if !ok {
		return nil, fmt.Errorf("unknown account %d", account)
	}
	var res []*autogen.BottleResource
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
func NewBottle(account int) *autogen.BottleResource {
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
	bottle := autogen.BottleResource{ID: newID}
	data[newID] = append(data[newID], &bottle)
	return &bottle
}

// Save persists bottle to database.
func Save(account int, b *autogen.BottleResource) {
	data[account] = append(data[account], b)
}

// Delete deletes bottle from database.
func Delete(account, id int) {
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
