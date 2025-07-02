package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/plantnet/minter/fabric"
)

type CidFile struct {
	CID         string `json:"cid"`
	SummaryPath string `json:"summary_path"`
}

type BatchFile struct {
	Summary map[string]fabric.SensorStats `json:"summary"`
}

var (
	debounceMap    = make(map[string]*time.Timer)
	alreadyHandled = make(map[string]bool)
	debounceMu     sync.Mutex
	debounceDur    = 1 * time.Second
)

func main() {
	dir := "/data/cids"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Fatalf("❌ CID directory does not exist: %s", dir)
	}
	watchCids(dir)
}

func watchCids(dir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("❌ Failed to create watcher:", err)
	}
	defer watcher.Close()

	err = watcher.Add(dir)
	if err != nil {
		log.Fatal("❌ Failed to watch directory:", err)
	}

	fmt.Println("👀 Watching directory:", dir)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if (event.Op&(fsnotify.Create|fsnotify.Write)) != 0 && filepath.Ext(event.Name) == ".json" {
				scheduleHandle(event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}

func scheduleHandle(path string) {
	debounceMu.Lock()
	if _, already := debounceMap[path]; already {
		// Already scheduling, don't queue again
		debounceMu.Unlock()
		return
	}

	fmt.Printf("🕒 Scheduled handler for %s\n", path)

	debounceMap[path] = time.AfterFunc(debounceDur, func() {
		debounceMu.Lock()
		alreadyHandled[path] = true
		delete(debounceMap, path)
		debounceMu.Unlock()

		handleNewFile(path)
	})
	debounceMu.Unlock()
}

func handleNewFile(path string) {
	time.Sleep(200 * time.Millisecond) // final buffer for file flush
	fmt.Println("📄 New file detected:", path)

	data, err := os.ReadFile(path)
	if err != nil {
		log.Println("❌ Failed to read CID file:", err)
		return
	}

	var cidInfo CidFile
	if err := json.Unmarshal(data, &cidInfo); err != nil {
		log.Println("❌ Failed to parse CID JSON:", err)
		return
	}

	fmt.Printf("✅ Parsed CID file: %+v\n", cidInfo)

	// Load batch summary
	summaryPath := "/data/" + cidInfo.SummaryPath
	summaryData, err := os.ReadFile(summaryPath)
	if err != nil {
		log.Printf("❌ Failed to read summary JSON at %s: %v\n", summaryPath, err)
		return
	}

	var batch BatchFile
	if err := json.Unmarshal(summaryData, &batch); err != nil {
		log.Printf("❌ Failed to parse batch summary JSON: %v\n", err)
		return
	}

	fmt.Println("📊 Summary:")
	for key, stat := range batch.Summary {
		fmt.Printf("   - %s: avg=%.2f, min=%.2f, max=%.2f\n", key, stat.Avg, stat.Min, stat.Max)
	}

	err = fabric.MintNFTToFabric(cidInfo.CID, batch.Summary)
	if err != nil {
		log.Printf("❌ Failed to mint NFT: %v\n", err)
	} else {
		fmt.Printf("🏷️ NFT successfully minted for CID: %s\n", cidInfo.CID)
	}
}
