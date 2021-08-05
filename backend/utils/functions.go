package utils

import (
	c "module/constants"
	"time"
)

func DateStringToTimestamp(str string) int64 {
	t, _ := time.Parse(c.DateLayout, str)
	return t.Unix()
}

func ArrayContains(array []string, element string) bool {
	var result bool
	for _, v := range array {
		if v == element {
			result = true
			break
		}
	}

	return result
}
