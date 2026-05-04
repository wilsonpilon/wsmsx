package main

import (
	"log"

	"ws7/internal/wsdev"
)

func main() {
	if err := wsdev.Run(); err != nil {
		log.Fatal(err)
	}
}
