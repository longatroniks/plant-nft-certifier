package fabric

type SensorStats struct {
	Avg float64 `json:"avg"`
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// Measurement is the batch description recorded on the plant's NFT
type Measurement struct {
	StartTimestamp float64                `json:"start_timestamp"`
	EndTimestamp   float64                `json:"end_timestamp"`
	ReadingCount   int                    `json:"reading_count"`
	DeviceIDs      []string               `json:"device_ids"`
	Summary        map[string]SensorStats `json:"summary"`
}

// PlantNFT mirrors the fields of the chaincode's token that the minter reports
type PlantNFT struct {
	ID         string `json:"id"`
	Version    int    `json:"version"`
	CurrentCID string `json:"current_cid"`
	Owner      string `json:"owner"`
}
