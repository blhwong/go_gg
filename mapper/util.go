package mapper

import (
	"strconv"
)

func getOrdinal(n int) string {
	s := strconv.Itoa(n)
	if n >= 11 && n <= 13 {
		return s + "th"
	}

	switch n % 10 {
	case 1:
		return s + "st"
	case 2:
		return s + "nd"
	case 3:
		return s + "rd"
	default:
		return s + "th"
	}
}
