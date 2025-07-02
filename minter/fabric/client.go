package fabric

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type SensorStats struct {
	Avg float64 `json:"avg"`
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

const (
	mspID         = "PlantOrgMSP"
	basePath      = "/opt/fabric/crypto-config/peerOrganizations/plantorg.plantnet.com"
	certPath      = basePath + "/users/User1@plantorg.plantnet.com/msp/signcerts/User1@plantorg.plantnet.com-cert.pem"
	keyDir        = basePath + "/users/User1@plantorg.plantnet.com/msp/keystore"
	tlsCertPath   = basePath + "/peers/peer0.plantorg.plantnet.com/tls/ca.crt"
	peerEndpoint  = "peer0.plantorg.plantnet.com:7051"
	gatewayPeer   = "peer0.plantorg.plantnet.com"
	channelName   = "sensor-readings-channel"
	chaincodeName = "plantnft"
)

func MintNFTToFabric(cid string, summary map[string]SensorStats) error {
	// Load certificate
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("read cert: %w", err)
	}
	cert, err := identity.CertificateFromPEM(certPEM)
	if err != nil {
		return fmt.Errorf("parse cert: %w", err)
	}

	// Load private key
	keyPath, err := findPrivateKey(keyDir)
	if err != nil {
		return fmt.Errorf("find private key: %w", err)
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("read key: %w", err)
	}
	privateKey, err := identity.PrivateKeyFromPEM(keyPEM)
	if err != nil {
		return fmt.Errorf("parse key: %w", err)
	}

	id, err := identity.NewX509Identity(mspID, cert)
	if err != nil {
		return fmt.Errorf("create identity: %w", err)
	}
	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return fmt.Errorf("create signer: %w", err)
	}

	// TLS connection
	caCert, err := loadX509Certificate(tlsCertPath)
	if err != nil {
		return fmt.Errorf("load TLS cert: %w", err)
	}
	certPool := x509.NewCertPool()
	certPool.AddCert(caCert)

	transportCreds := credentials.NewClientTLSFromCert(certPool, gatewayPeer)
	conn, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCreds))
	if err != nil {
		return fmt.Errorf("gRPC dial: %w", err)
	}
	defer conn.Close()

	// Connect to Gateway
	gw, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(conn),
	)
	if err != nil {
		return fmt.Errorf("gateway connect: %w", err)
	}
	defer gw.Close()

	// Prepare transaction
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
		return fmt.Errorf("MintNFT transaction failed: %w", err)
	}

	fmt.Printf("✅ MintNFT for CID %s committed to ledger.\n", cid)

	// --- 🔎 Confirm with GetNFT ---
	result, err := contract.EvaluateTransaction("GetNFT", cid)
	if err != nil {
		return fmt.Errorf("🔎 GetNFT query failed: %w", err)
	}

	fmt.Println("🔎 Test GetNFT query result:")
	fmt.Println(string(result))

	return nil
}

// --- Helper functions ---

func loadX509Certificate(path string) (*x509.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block in %s", path)
	}
	return x509.ParseCertificate(block.Bytes)
}

func findPrivateKey(dir string) (string, error) {
	var keyPath string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && keyPath == "" {
			keyPath = path
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if keyPath == "" {
		return "", fmt.Errorf("no key found in %s", dir)
	}
	return keyPath, nil
}
