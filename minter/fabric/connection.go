package fabric

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
)

func ConnectGateway() (*grpc.ClientConn, error) {
	caCert, err := loadX509Certificate(tlsCertPath)
	if err != nil {
		return nil, fmt.Errorf("load TLS cert: %w", err)
	}
	certPool := x509.NewCertPool()
	certPool.AddCert(caCert)

	creds := credentials.NewClientTLSFromCert(certPool, gatewayPeer)
	return grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(creds))
}

func NewGatewayClient(conn *grpc.ClientConn, id *identity.X509Identity, sign identity.Sign) (*client.Gateway, error) {
	return client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(conn),
	)
}

func loadX509Certificate(path string) (*x509.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in %s", path)
	}
	return x509.ParseCertificate(block.Bytes)
}
