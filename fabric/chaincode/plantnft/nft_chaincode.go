package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SensorStats holds aggregated sensor values
type SensorStats struct {
	Avg float64 `json:"avg"`
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// PlantNFT represents a plant's lifecycle certificate
type PlantNFT struct {
	ID        string                  `json:"id"`
	CID       string                  `json:"cid"`
	Timestamp int64                   `json:"timestamp"`
	Summary   map[string]SensorStats  `json:"summary"`
}

// SmartContract provides functions for managing NFTs
type SmartContract struct {
	contractapi.Contract
}

// MintNFT creates a new NFT record on the ledger
func (s *SmartContract) MintNFT(ctx contractapi.TransactionContextInterface, id string, cid string, timestamp int64, summaryJSON string) error {
	exists, err := s.NFTExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("NFT with ID %s already exists", id)
	}

	var summary map[string]SensorStats
	err = json.Unmarshal([]byte(summaryJSON), &summary)
	if err != nil {
		return fmt.Errorf("failed to parse summary JSON: %v", err)
	}

	nft := PlantNFT{
		ID:        id,
		CID:       cid,
		Timestamp: timestamp,
		Summary:   summary,
	}

	nftBytes, err := json.Marshal(nft)
	if err != nil {
		return fmt.Errorf("failed to marshal NFT: %v", err)
	}

	return ctx.GetStub().PutState(id, nftBytes)
}

// GetNFT retrieves a PlantNFT from the ledger
func (s *SmartContract) GetNFT(ctx contractapi.TransactionContextInterface, id string) (*PlantNFT, error) {
	nftBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from ledger: %v", err)
	}
	if nftBytes == nil {
		return nil, fmt.Errorf("NFT %s does not exist", id)
	}

	var nft PlantNFT
	err = json.Unmarshal(nftBytes, &nft)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal NFT: %v", err)
	}

	return &nft, nil
}

// NFTExists returns true if an NFT with the given ID exists
func (s *SmartContract) NFTExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	nftBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, err
	}
	return nftBytes != nil, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		panic(fmt.Sprintf("Error creating chaincode: %v", err))
	}

	if err := chaincode.Start(); err != nil {
		panic(fmt.Sprintf("Error starting chaincode: %v", err))
	}
}
