package main

import (
	"time"
)

func main() {
	t, _ := time.Parse("2006-01-02", "2020-08-17")
	t.Format("2006-01-02T15:04:05.000Z")
}
