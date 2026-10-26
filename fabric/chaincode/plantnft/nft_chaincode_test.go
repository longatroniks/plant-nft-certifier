package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/msp"
)

func creator(t *testing.T, cn string) []byte {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: cn},
		Issuer:       pkix.Name{CommonName: "ca"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := proto.Marshal(&msp.SerializedIdentity{
		Mspid:   "PlantOrgMSP",
		IdBytes: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
	})
	return id
}

func newStub(t *testing.T) *shimtest.MockStub {
	cc, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		t.Fatal(err)
	}
	return shimtest.NewMockStub("plantnft", cc)
}

var txn = 0

func invoke(stub *shimtest.MockStub, who []byte, fn string, args ...string) (int32, string) {
	txn++
	stub.Creator = who
	argv := [][]byte{[]byte(fn)}
	for _, a := range args {
		argv = append(argv, []byte(a))
	}
	res := stub.MockInvoke(string(rune('a'+txn)), argv)
	if res.Status != 200 {
		return res.Status, res.Message
	}
	return res.Status, string(res.Payload)
}

const m1 = `{"start_timestamp":1,"end_timestamp":2,"reading_count":10,"device_ids":["node-01"],"summary":{"temperature":{"avg":32.45,"min":30.3,"max":36.9}}}`
const m2 = `{"start_timestamp":3,"end_timestamp":4,"reading_count":12,"device_ids":["node-01"],"summary":{"temperature":{"avg":31.0,"min":29.0,"max":33.0}}}`

func TestDynamicNFTLifecycle(t *testing.T) {
	stub := newStub(t)
	farmer, buyer := creator(t, "farmer"), creator(t, "buyer")

	if st, msg := invoke(stub, farmer, "RecordMeasurement", "olive-01", "QmA", m1); st != 200 {
		t.Fatalf("mint failed: %s", msg)
	}
	if st, msg := invoke(stub, farmer, "RecordMeasurement", "olive-01", "QmB", m2); st != 200 {
		t.Fatalf("update failed: %s", msg)
	}

	_, out := invoke(stub, farmer, "GetNFT", "olive-01")
	var nft PlantNFT
	if err := json.Unmarshal([]byte(out), &nft); err != nil {
		t.Fatalf("bad GetNFT payload %q: %v", out, err)
	}
	if nft.Version != 2 || nft.CurrentCID != "QmB" || len(nft.Measurements) != 2 {
		t.Fatalf("unexpected token state: %+v", nft)
	}
	if nft.Measurements[0].CID != "QmA" || nft.CurrentSummary["temperature"].Avg != 31.0 {
		t.Fatalf("measurement history wrong: %+v", nft.Measurements)
	}

	if st, _ := invoke(stub, farmer, "RecordMeasurement", "olive-02", "QmA", m1); st == 200 {
		t.Fatal("same CID recorded twice")
	}
	if _, tok := invoke(stub, farmer, "GetTokenByCID", "QmB"); tok != "olive-01" {
		t.Fatalf("GetTokenByCID = %q", tok)
	}
	if _, tok := invoke(stub, farmer, "GetTokenByCID", "QmZ"); tok != "" {
		t.Fatalf("GetTokenByCID for unknown CID = %q", tok)
	}

	if st, _ := invoke(stub, buyer, "RecordMeasurement", "olive-01", "QmC", m2); st == 200 {
		t.Fatal("non-owner could update the token")
	}
	if st, _ := invoke(stub, buyer, "TransferNFT", "olive-01", "someone"); st == 200 {
		t.Fatal("non-owner could transfer the token")
	}

	if st, msg := invoke(stub, farmer, "TransferNFT", "olive-01", "new-owner"); st != 200 {
		t.Fatalf("transfer failed: %s", msg)
	}
	_, out = invoke(stub, farmer, "GetNFT", "olive-01")
	json.Unmarshal([]byte(out), &nft)
	if nft.Owner != "new-owner" {
		t.Fatalf("owner not updated: %s", nft.Owner)
	}
	if st, _ := invoke(stub, farmer, "RecordMeasurement", "olive-01", "QmD", m2); st == 200 {
		t.Fatal("previous owner could still update after transfer")
	}
}
