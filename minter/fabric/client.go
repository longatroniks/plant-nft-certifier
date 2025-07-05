package fabric

import (
	"encoding/json"
	"fmt"
	"time"
)

func MintNFTToFabric(cid string, summary map[string]SensorStats) error {
	identity, signer, err := LoadIdentity()
	if err != nil {
		return fmt.Errorf("load identity: %w", err)
	}

	conn, err := ConnectGateway()
	if err != nil {
		return fmt.Errorf("connect gateway: %w", err)
	}
	defer conn.Close()

	gw, err := NewGatewayClient(conn, identity, signer)
	if err != nil {
		return fmt.Errorf("gateway client: %w", err)
	}
	defer gw.Close()

	network := gw.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)

	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("marshal summary: %w", err)
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	fmt.Println("🚀 Submitting MintNFT transaction...")

	_, err = contract.SubmitTransaction("MintNFT", cid, cid, timestamp, string(summaryJSON))
	if err != nil {
		return fmt.Errorf("submit transaction: %w", err)
	}

	fmt.Printf("✅ MintNFT for CID %s committed.\n", cid)

	result, err := contract.EvaluateTransaction("GetNFT", cid)
	if err != nil {
		return fmt.Errorf("GetNFT failed: %w", err)
	}
	fmt.Println("🔎 GetNFT result:", string(result))

	return nil
}
