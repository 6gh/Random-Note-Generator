package main

import (
	"log"
)

func main() {
	createGUI()
}

func handleErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func logf(format string, a ...any) {
	log.Printf(format, a...)
}
