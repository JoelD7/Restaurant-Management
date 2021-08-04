package main

import (
	"fmt"
	"time"
)

func main() {
	t, _ := time.Parse("2006-01-02", "2020-08-17")
	fmt.Println(t.Unix())
}
