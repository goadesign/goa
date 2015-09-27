package app

import (
	"fmt"
	"time"
)

// An account represents the unit of tenancy.
type Account struct {
	ID   int    // ID of account
	Name string // Name of account
}

// ComputeHref returns the account href.
func (a *Account) ComputeHref() string {
	return fmt.Sprintf("/accounts/%d", a.ID)
}

// A bottle of wine with associated rating
type Bottle struct {
	ID        int       // ID of bottle
	AccountID int       // ID of account which owns bottle
	Href      string    // API href of bottle
	Name      string    // Name of wine
	Vineyard  string    // Name of vineyard / winery
	Varietal  string    // Wine varietal
	Vintage   int       // Wine vintage
	Color     string    // Type of wine, one of "red", "white", "rose" or "yellow".
	Sweet     bool      // Whether wine is sweet or dry
	Country   string    // Country of origin
	Region    string    // Region
	Review    string    // Review
	Ratings   int       // Bottle rattings
	RatedAt   time.Time // last rating timestamp
}

// ComputeHref returns the bottle href.
func (b *Bottle) ComputeHref() string {
	account := &Account{ID: b.AccountID}
	return fmt.Sprintf("%s/bottles/%d", account.ComputeHref(), b.ID)
}

// Validate implements the validation rules defined in the corresponding media type.
// It returns nil if the validation succeeds, an error otherwise.
func (b *Bottle) Validate() error {
	if b.Name == "" {
		return fmt.Errorf(`field "name" is required and cannot be empty`)
	}
	if b.Vineyard == "" {
		return fmt.Errorf(`field "vineyard" is required and cannot be empty`)
	}
	if b.Color != "" && b.Color != "red" && b.Color != "white" && b.Color != "rose" && b.Color != "yellow" {
		return fmt.Errorf(`field "color" must be one of "red", "white", "rose" or "yellow"`)
	}
	return nil
}
