package utils

const (
	Breakfast = "breakfast"
	Lunch     = "lunch"
	Dinner    = "dinner"
)

func IsValidCatalog(catalog string) bool {
	switch catalog {
	case Breakfast, Lunch, Dinner:
		return true
	}
	return false
}
