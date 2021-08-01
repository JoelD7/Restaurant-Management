package utils

import (
	c "module/constants"
	"time"
)

func DateStringToTimestamp(str string) int64 {
	t, _ := time.Parse(c.DateLayout, str)
	return t.Unix()
}
