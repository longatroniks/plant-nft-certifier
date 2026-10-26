package fabric

import (
	"encoding/json"
	"fmt"
)

// RecordMeasurementToFabric certifies one batch on the plant's dynamic NFT,
// minting the token if this is its first measurement. It returns the token
// state after the update, or nil if the CID had already been recorded.
func RecordMeasurementToFabric(tokenID string, cid string, m Measurement) (*PlantNFT, error) {
	identity, signer, err := LoadIdentity()
	if err != nil {
		return nil, fmt.Errorf("load identity: %w", err)
	}

	conn, err := ConnectGateway()
	if err != nil {
		return nil, fmt.Errorf("connect gateway: %w", err)
	}
	defer conn.Close()

	gw, err := NewGatewayClient(conn, identity, signer)
	if err != nil {
		return nil, fmt.Errorf("gateway client: %w", err)
	}
	defer gw.Close()

	network := gw.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)

	recordedOn, err := contract.EvaluateTransaction("GetTokenByCID", cid)
	if err != nil {
		return nil, fmt.Errorf("GetTokenByCID failed: %w", err)
	}
	if len(recordedOn) > 0 {
		fmt.Printf("⏭️ CID %s already recorded on %s, skipping.\n", cid, string(recordedOn))
		return nil, nil
	}

	measurementJSON, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("marshal measurement: %w", err)
	}

	fmt.Printf("🚀 Submitting RecordMeasurement for %s...\n", tokenID)
	_, err = contract.SubmitTransaction("RecordMeasurement", tokenID, cid, string(measurementJSON))
	if err != nil {
		return nil, fmt.Errorf("submit transaction: %w", err)
	}

	result, err := contract.EvaluateTransaction("GetNFT", tokenID)
	if err != nil {
		return nil, fmt.Errorf("GetNFT failed: %w", err)
	}
	var nft PlantNFT
	if err := json.Unmarshal(result, &nft); err != nil {
		return nil, fmt.Errorf("parse GetNFT result: %w", err)
	}
	return &nft, nil
}
