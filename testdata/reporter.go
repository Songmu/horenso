package main

import (
	"io/ioutil"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		return
	}
	file := os.Args[1]
	bytes, _ := ioutil.ReadAll(os.Stdin)
	json := string(bytes)
	ioutil.WriteFile(file, []byte(json), os.ModePerm)
}
