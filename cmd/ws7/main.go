package main

import (
	"log"

	"ws7/internal/ui"
)

func main() {
	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}
}

