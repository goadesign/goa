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
		42: []*autogen.BottleResource{
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

// Return bottle with given id from given account.
func GetBottle(account, id int) (*autogen.BottleResource, error) {
	bottles, ok := data[account]
	if !ok {
		return nil, fmt.Errorf("unknown account %d", account)
	}
	for _, b := range bottles {
		if b.ID == id {
			return b, nil
		}
	}
	return nil, fmt.Errorf("no bottle with id %d in account %d", id, account)
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
