package main

import (
	"fmt"

	mach "github.com/machbase/neo-engine"
)

func main() {
	fmt.Println("hello world")
	fmt.Println("-------------------------------")
	fmt.Println(mach.LinkInfo())

	mach.Initialize(".")
}
