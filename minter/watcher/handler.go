package watcher

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/plantnet/minter/fabric"
	"github.com/plantnet/minter/parser"
	"github.com/plantnet/minter/qr"
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

	tokenID := batch.TreeID
	if tokenID == "" {
		// batches from before tree_id was added
		tokenID = defaultTokenID()
	}

	for key, stat := range batch.AggregatedSummary {
		fmt.Printf("   - %s: avg=%.2f, min=%.2f, max=%.2f\n", key, stat.Avg, stat.Min, stat.Max)
	}

	nft, err := fabric.RecordMeasurementToFabric(tokenID, cidInfo.CID, fabric.Measurement{
		StartTimestamp: batch.StartTimestamp,
		EndTimestamp:   batch.EndTimestamp,
		ReadingCount:   batch.ReadingCount,
		DeviceIDs:      batch.DeviceIDs,
		Summary:        batch.AggregatedSummary,
	})
	if err != nil {
		log.Printf("❌ Failed to record measurement: %v\n", err)
		return
	}
	if nft == nil {
		return
	}
	if nft.Version == 1 {
		fmt.Printf("🏷️ NFT %s minted with CID %s\n", nft.ID, cidInfo.CID)
	} else {
		fmt.Printf("🏷️ NFT %s updated to version %d with CID %s\n", nft.ID, nft.Version, cidInfo.CID)
	}
	qr.PrintIPFSQRCode(cidInfo.CID)
}

func defaultTokenID() string {
	if id := os.Getenv("DEFAULT_TOKEN_ID"); id != "" {
		return id
	}
	return "olive-tree-01"
}
