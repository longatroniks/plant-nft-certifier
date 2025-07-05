package fabric

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-gateway/pkg/identity"
)


func LoadIdentity() (*identity.X509Identity, identity.Sign, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read cert: %w", err)
	}
	cert, err := identity.CertificateFromPEM(certPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("parse cert: %w", err)
	}

	keyPath, err := findPrivateKey(keyDir)
	if err != nil {
		return nil, nil, fmt.Errorf("find key: %w", err)
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read key: %w", err)
	}
	privateKey, err := identity.PrivateKeyFromPEM(keyPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("parse key: %w", err)
	}

	id, err := identity.NewX509Identity(mspID, cert)
	if err != nil {
		return nil, nil, fmt.Errorf("new identity: %w", err)
	}
	signer, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("new signer: %w", err)
	}

	return id, signer, nil
}

func findPrivateKey(dir string) (string, error) {
	var keyPath string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
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
