package main

import (
	"log"

	"video-transcript/internal/app"
)

func main() {
	if err := app.RunServer(); err != nil {
		log.Fatal(err)
	}
}
