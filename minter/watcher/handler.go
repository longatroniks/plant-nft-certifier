package watcher

import (
	"fmt"
	"log"
	"time"

	"github.com/plantnet/minter/fabric"
	"github.com/plantnet/minter/parser"
)

func scheduleHandle(path string) {
	debounceMu.Lock()
	if _, already := debounceMap[path]; already {
		debounceMu.Unlock()
		return
	}

	debounceMap[path] = time.AfterFunc(debounceDur, func() {
		debounceMu.Lock()
		delete(debounceMap, path)
		debounceMu.Unlock()

		handle(path)
	})
	debounceMu.Unlock()
}

func handle(path string) {
	time.Sleep(200 * time.Millisecond)
	fmt.Println("📄 New file detected:", path)

	cidInfo, err := parser.ParseCidFile(path)
	if err != nil {
		log.Println("❌", err)
		return
	}

	batch, err := parser.LoadSummary("/data/" + cidInfo.SummaryPath)
	if err != nil {
		log.Println("❌", err)
		return
	}

	for key, stat := range batch.AggregatedSummary {
		fmt.Printf("   - %s: avg=%.2f, min=%.2f, max=%.2f\n", key, stat.Avg, stat.Min, stat.Max)
	}

	if err := fabric.MintNFTToFabric(cidInfo.CID, batch.AggregatedSummary); err != nil {

		log.Printf("❌ Failed to mint NFT: %v\n", err)
	} else {
		fmt.Printf("🏷️ NFT successfully minted for CID: %s\n", cidInfo.CID)
	}
}
