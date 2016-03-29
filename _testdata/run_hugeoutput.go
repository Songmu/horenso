package main

import (
	"fmt"
	"strings"
)

func main() {
	// output 64K + 1 bytes
	for i := 0; i < 1024; i++ {
		fmt.Print(strings.Repeat("x", 63), "\n")
	}
	fmt.Print("\n")
}
