package watcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	debounceMap = make(map[string]*time.Timer)
	debounceMu  sync.Mutex
	debounceDur = 1 * time.Second
)

func Watch(dir string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	defer watcher.Close()

	// The aggregator only creates the directory after its first batch,
	// so create it here instead of failing on a fresh deployment.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	if err := watcher.Add(dir); err != nil {
		return fmt.Errorf("watch directory: %w", err)
	}

	fmt.Println("👀 Watching directory:", dir)

	// Mint anything written while the minter was not running.
	// Already-minted CIDs are skipped by MintNFTToFabric.
	existing, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return fmt.Errorf("scan directory: %w", err)
	}
	for _, path := range existing {
		scheduleHandle(path)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if (event.Op&(fsnotify.Create|fsnotify.Write)) != 0 && filepath.Ext(event.Name) == ".json" {
				scheduleHandle(event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Println("Watcher error:", err)
		}
	}
}
