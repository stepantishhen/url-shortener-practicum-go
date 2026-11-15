package main

import (
	"fmt"
	operatingsystem "os"
)

func main() {
	operatingsystem.Exit(1) // want "direct call to os.Exit is forbidden"
	fmt.Errorf("not os.Exit")
}

func helper() {
	operatingsystem.Exit(1)
}
