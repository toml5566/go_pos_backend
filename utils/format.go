package utils

import (
	"fmt"
	"time"
)

func FormattedDateNow() string {
	return time.Now().Format("2006-01-02")
}

func FormottedDecimalToString(d float64) string {
	return fmt.Sprintf("%.2f", d)
}
