package main

import (
	"os"

	"github.com/Songmu/horenso"
)

func main() {
	os.Exit(horenso.Run(os.Args[1:]))
}
