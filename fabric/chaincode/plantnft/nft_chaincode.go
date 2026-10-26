package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const cidIndex = "cid~token"

// SensorStats holds aggregated sensor values
type SensorStats struct {
	Avg float64 `json:"avg"`
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// Measurement is one certified batch of sensor readings, stored on IPFS under CID
type Measurement struct {
	CID            string                 `json:"cid"`
	RecordedAt     int64                  `json:"recorded_at"`
	TxID           string                 `json:"tx_id"`
	StartTimestamp float64                `json:"start_timestamp"`
	EndTimestamp   float64                `json:"end_timestamp"`
	ReadingCount   int                    `json:"reading_count"`
	DeviceIDs      []string               `json:"device_ids"`
	Summary        map[string]SensorStats `json:"summary"`
}

// PlantNFT is a dynamic certificate for one plant: it is minted with the first
// measurement and grows as new measurements of the same plant are certified.
type PlantNFT struct {
	ID             string                 `json:"id"`
	Owner          string                 `json:"owner"`
	MintedAt       int64                  `json:"minted_at"`
	UpdatedAt      int64                  `json:"updated_at"`
	Version        int                    `json:"version"`
	CurrentCID     string                 `json:"current_cid"`
	CurrentSummary map[string]SensorStats `json:"current_summary"`
	Measurements   []Measurement          `json:"measurements"`
}

// HistoryEntry is one past state of a token, as kept by the ledger itself
type HistoryEntry struct {
	TxID      string    `json:"tx_id"`
	Timestamp int64     `json:"timestamp"`
	IsDelete  bool      `json:"is_delete"`
	NFT       *PlantNFT `json:"nft,omitempty"`
}

// SmartContract provides functions for managing dynamic plant NFTs
type SmartContract struct {
	contractapi.Contract
}

// RecordMeasurement certifies a new batch for a plant. The first call for a
// token ID mints it, owned by the caller; later calls append to it and may
// only be made by the current owner. A CID can be recorded only once.
func (s *SmartContract) RecordMeasurement(ctx contractapi.TransactionContextInterface, id string, cid string, measurementJSON string) error {
	if id == "" || cid == "" {
		return fmt.Errorf("token id and cid are required")
	}

	existing, err := s.GetTokenByCID(ctx, cid)
	if err != nil {
		return err
	}
	if existing != "" {
		return fmt.Errorf("CID %s is already recorded on token %s", cid, existing)
	}

	var m Measurement
	if err := json.Unmarshal([]byte(measurementJSON), &m); err != nil {
		return fmt.Errorf("failed to parse measurement JSON: %v", err)
	}
	if len(m.Summary) == 0 {
		return fmt.Errorf("measurement summary is empty")
	}

	// Use the transaction timestamp, not the caller's clock, so every
	// endorsing peer computes the same value.
	ts, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("failed to read tx timestamp: %v", err)
	}
	m.CID = cid
	m.RecordedAt = ts.GetSeconds()
	m.TxID = ctx.GetStub().GetTxID()

	caller, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to read client identity: %v", err)
	}

	nft, err := s.readNFT(ctx, id)
	if err != nil {
		return err
	}
	event := "PlantNFTUpdated"
	if nft == nil {
		nft = &PlantNFT{ID: id, Owner: caller, MintedAt: m.RecordedAt}
		event = "PlantNFTMinted"
	} else if nft.Owner != caller {
		return fmt.Errorf("only the owner of %s can record measurements", id)
	}

	nft.Measurements = append(nft.Measurements, m)
	nft.Version = len(nft.Measurements)
	nft.UpdatedAt = m.RecordedAt
	nft.CurrentCID = cid
	nft.CurrentSummary = m.Summary

	if err := s.putNFT(ctx, nft); err != nil {
		return err
	}

	key, err := ctx.GetStub().CreateCompositeKey(cidIndex, []string{cid, id})
	if err != nil {
		return fmt.Errorf("failed to create CID index key: %v", err)
	}
	if err := ctx.GetStub().PutState(key, []byte{0x00}); err != nil {
		return fmt.Errorf("failed to write CID index: %v", err)
	}

	payload, _ := json.Marshal(map[string]interface{}{"id": id, "cid": cid, "version": nft.Version})
	return ctx.GetStub().SetEvent(event, payload)
}

// TransferNFT hands a token to a new owner; only the current owner may do so
func (s *SmartContract) TransferNFT(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
	if newOwner == "" {
		return fmt.Errorf("new owner is required")
	}
	nft, err := s.readNFT(ctx, id)
	if err != nil {
		return err
	}
	if nft == nil {
		return fmt.Errorf("NFT %s does not exist", id)
	}
	caller, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to read client identity: %v", err)
	}
	if nft.Owner != caller {
		return fmt.Errorf("only the owner of %s can transfer it", id)
	}

	nft.Owner = newOwner
	if err := s.putNFT(ctx, nft); err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]string{"id": id, "owner": newOwner})
	return ctx.GetStub().SetEvent("PlantNFTTransferred", payload)
}

// GetNFT retrieves a PlantNFT from the ledger
func (s *SmartContract) GetNFT(ctx contractapi.TransactionContextInterface, id string) (*PlantNFT, error) {
	nft, err := s.readNFT(ctx, id)
	if err != nil {
		return nil, err
	}
	if nft == nil {
		return nil, fmt.Errorf("NFT %s does not exist", id)
	}
	return nft, nil
}

// NFTExists returns true if an NFT with the given ID exists
func (s *SmartContract) NFTExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	nftBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, err
	}
	return nftBytes != nil, nil
}

// GetTokenByCID returns the token a CID was recorded on, or "" if none
func (s *SmartContract) GetTokenByCID(ctx contractapi.TransactionContextInterface, cid string) (string, error) {
	it, err := ctx.GetStub().GetStateByPartialCompositeKey(cidIndex, []string{cid})
	if err != nil {
		return "", fmt.Errorf("failed to query CID index: %v", err)
	}
	defer it.Close()

	if !it.HasNext() {
		return "", nil
	}
	kv, err := it.Next()
	if err != nil {
		return "", err
	}
	_, parts, err := ctx.GetStub().SplitCompositeKey(kv.Key)
	if err != nil || len(parts) != 2 {
		return "", fmt.Errorf("malformed CID index key")
	}
	return parts[1], nil
}

// GetNFTHistory returns every committed state of a token, oldest first
func (s *SmartContract) GetNFTHistory(ctx contractapi.TransactionContextInterface, id string) ([]HistoryEntry, error) {
	it, err := ctx.GetStub().GetHistoryForKey(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read history: %v", err)
	}
	defer it.Close()

	history := []HistoryEntry{}
	for it.HasNext() {
		mod, err := it.Next()
		if err != nil {
			return nil, err
		}
		entry := HistoryEntry{TxID: mod.TxId, IsDelete: mod.IsDelete}
		if mod.Timestamp != nil {
			entry.Timestamp = mod.Timestamp.GetSeconds()
		}
		if !mod.IsDelete {
			var nft PlantNFT
			if err := json.Unmarshal(mod.Value, &nft); err == nil {
				entry.NFT = &nft
			}
		}
		history = append(history, entry)
	}
	return history, nil
}

func (s *SmartContract) readNFT(ctx contractapi.TransactionContextInterface, id string) (*PlantNFT, error) {
	nftBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from ledger: %v", err)
	}
	if nftBytes == nil {
		return nil, nil
	}
	var nft PlantNFT
	if err := json.Unmarshal(nftBytes, &nft); err != nil {
		return nil, fmt.Errorf("failed to unmarshal NFT: %v", err)
	}
	return &nft, nil
}

func (s *SmartContract) putNFT(ctx contractapi.TransactionContextInterface, nft *PlantNFT) error {
	nftBytes, err := json.Marshal(nft)
	if err != nil {
		return fmt.Errorf("failed to marshal NFT: %v", err)
	}
	return ctx.GetStub().PutState(nft.ID, nftBytes)
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
