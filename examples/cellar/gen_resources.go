package main

// A bottle of wine with associated rating
type BottleResource struct {
	Id              int      // ID of bottle
	Href            string   // API href of bottle
	Name            string   // Name of wine
	Vineyard        string   // Name of vineyard / winery
	Varietal        string   // Wine varietal
	Vintage         int      // Wine vintage
	Color           string   // Type of wine, one of "red", "white", "rose" or "yellow".
	Sweet           bool     // Whether wine is sweet or dry
	Country         string   // Country of origin
	Region          string   // Region
	Review          string   // Review
	Characteristics []string // Wine characteristics
}
