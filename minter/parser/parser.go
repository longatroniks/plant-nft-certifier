package parser

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/plantnet/minter/fabric"
)

type CidFile struct {
	CID         string `json:"cid"`
	SummaryPath string `json:"summary_path"`
}

type BatchFile struct {
	AggregatedSummary map[string]fabric.SensorStats `json:"aggregated_summary"`
}

func ParseCidFile(path string) (*CidFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read CID file: %w", err)
	}

	var cidInfo CidFile
	if err := json.Unmarshal(data, &cidInfo); err != nil {
		return nil, fmt.Errorf("parse CID JSON: %w", err)
	}
	return &cidInfo, nil
}

func LoadSummary(path string) (*BatchFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read summary JSON: %w", err)
	}

	var batch BatchFile
	if err := json.Unmarshal(data, &batch); err != nil {
		return nil, fmt.Errorf("parse summary JSON: %w", err)
	}
	return &batch, nil
}
