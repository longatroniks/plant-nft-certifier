package main

import (
	"log"

	"github.com/plantnet/minter/watcher"
)

func main() {
	const dir = "/data/cids"
	if err := watcher.Watch(dir); err != nil {
		log.Fatalf("❌ %v", err)
	}
}
