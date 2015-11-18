package controllers

import (
	"fmt"
	"sync"

	"github.com/raphael/goa/examples/cellar/app"
)

// DB emulates a database driver using in-memory data structures.
type DB struct {
	sync.Mutex
	maxAccountID int
	accounts     map[int]*app.Account
	bottles      map[int][]*app.Bottle
}

// NewDB initializes a new "DB" with dummy data.
func NewDB() *DB {
	account := &app.Account{ID: 1, Name: "account 1", Href: app.AccountHref(1)}
	account2 := &app.Account{ID: 2, Name: "account 2", Href: app.AccountHref(2)}
	bottles := map[int][]*app.Bottle{
		1: []*app.Bottle{
			&app.Bottle{
				ID:        100,
				Account:   account,
				Href:      app.BottleHref(1, 100),
				Name:      "Number 8",
				Vineyard:  "Asti Winery",
				Varietal:  "Merlot",
				Vintage:   2012,
				Color:     "red",
				Sweetness: 1,
				Country:   "USA",
				Region:    "California",
				Review:    "Great value",
				Rating:    4,
			},
			&app.Bottle{
				ID:        101,
				Account:   account,
				Href:      app.BottleHref(1, 101),
				Name:      "Mourvedre",
				Vineyard:  "Rideau",
				Varietal:  "Mourvedre",
				Vintage:   2012,
				Color:     "red",
				Sweetness: 1,
				Country:   "USA",
				Region:    "California",
				Review:    "Good but expensive",
				Rating:    3,
			},
			&app.Bottle{
				ID:        102,
				Account:   account,
				Href:      app.BottleHref(1, 102),
				Name:      "Blue's Cuvee",
				Vineyard:  "Longoria",
				Varietal:  "Cabernet Franc with Merlot, Malbec, Cabernet Sauvignon and Syrah",
				Vintage:   2012,
				Color:     "red",
				Sweetness: 1,
				Country:   "USA",
				Region:    "California",
				Review:    "Favorite",
				Rating:    5,
			},
		},
		2: []*app.Bottle{
			&app.Bottle{
				ID:        200,
				Account:   account2,
				Href:      app.BottleHref(42, 200),
				Name:      "Blackstone Merlot",
				Vineyard:  "Blackstone",
				Varietal:  "Merlot",
				Vintage:   2012,
				Color:     "red",
				Sweetness: 1,
				Country:   "USA",
				Region:    "California",
				Review:    "OK",
				Rating:    3,
			},
			&app.Bottle{
				ID:        201,
				Account:   account2,
				Href:      app.BottleHref(42, 201),
				Name:      "Wild Horse",
				Vineyard:  "Wild Horse",
				Varietal:  "Pinot Noir",
				Vintage:   2012,
				Color:     "red",
				Sweetness: 1,
				Country:   "USA",
				Region:    "California",
				Review:    "Solid Pinot",
				Rating:    4,
			},
		},
	}
	return &DB{accounts: map[int]*app.Account{1: account, 2: account2}, bottles: bottles, maxAccountID: 2}
}

// GetAccount returns the account with given id if any, nil otherwise.
func (db *DB) GetAccount(id int) *app.Account {
	db.Lock()
	defer db.Unlock()
	return db.accounts[id]
}

// NewAccount creates a new blank account resource.
func (db *DB) NewAccount() *app.Account {
	db.Lock()
	defer db.Unlock()
	db.maxAccountID++
	account := &app.Account{ID: db.maxAccountID}
	db.accounts[db.maxAccountID] = account
	return account
}

// SaveAccount "persists" the account.
func (db *DB) SaveAccount(a *app.Account) {
	db.Lock()
	defer db.Unlock()
	db.accounts[a.ID] = a
}

// DeleteAccount deletes the account.
func (db *DB) DeleteAccount(account *app.Account) {
	db.Lock()
	defer db.Unlock()
	if account == nil {
		return
	}
	delete(db.bottles, account.ID)
	delete(db.accounts, account.ID)
}

// GetBottle returns the bottle with the given id from the given account or nil if not found.
func (db *DB) GetBottle(account, id int) *app.Bottle {
	db.Lock()
	defer db.Unlock()
	bottles, ok := db.bottles[account]
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
func (db *DB) GetBottles(account int) ([]*app.Bottle, error) {
	db.Lock()
	defer db.Unlock()
	bottles, ok := db.bottles[account]
	if !ok {
		return nil, fmt.Errorf("unknown account %d", account)
	}
	return bottles, nil
}

// GetBottlesByYears returns the bottles with the vintage in the given array from the given account.
func (db *DB) GetBottlesByYears(account int, years []int) ([]*app.Bottle, error) {
	db.Lock()
	defer db.Unlock()
	bottles, ok := db.bottles[account]
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
func (db *DB) NewBottle(account int) *app.Bottle {
	db.Lock()
	defer db.Unlock()
	bottles, _ := db.bottles[account]
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
	db.bottles[newID] = append(db.bottles[newID], &bottle)
	return &bottle
}

// SaveBottle persists bottle to bottlesbase.
func (db *DB) SaveBottle(b *app.Bottle) {
	db.Lock()
	defer db.Unlock()
	db.bottles[b.Account.ID] = append(db.bottles[b.Account.ID], b)
}

// DeleteBottle deletes bottle from bottlesbase.
func (db *DB) DeleteBottle(bottle *app.Bottle) {
	db.Lock()
	defer db.Unlock()
	if bottle == nil {
		return
	}
	account := bottle.Account
	id := bottle.ID
	if bs, ok := db.bottles[account.ID]; ok {
		idx := -1
		for i, b := range bs {
			if b.ID == id {
				idx = i
				break
			}
		}
		if idx > -1 {
			bs = append(bs[:idx], bs[idx+1:]...)
			db.bottles[account.ID] = bs
		}
	}
}
