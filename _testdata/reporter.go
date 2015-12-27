package main

import (
	"io/ioutil"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		return
	}
	file := os.Args[1]
	json := os.Args[2]
	ioutil.WriteFile(file, []byte(json), os.ModePerm)
}
