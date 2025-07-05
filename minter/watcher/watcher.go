package watcher

import (
	"fmt"
	"log"
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

	if err := watcher.Add(dir); err != nil {
		return fmt.Errorf("watch directory: %w", err)
	}

	fmt.Println("👀 Watching directory:", dir)

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
