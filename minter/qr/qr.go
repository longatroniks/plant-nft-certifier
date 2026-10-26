package qr

import (
	"fmt"
	"os"

	"github.com/skip2/go-qrcode"
)

// Public gateway used in the QR link; any gateway works once the CID is pinned
func gatewayURL() string {
	if gw := os.Getenv("IPFS_GATEWAY"); gw != "" {
		return gw
	}
	return "https://ipfs.io/ipfs/"
}

func PrintIPFSQRCode(cid string) {
	url := gatewayURL() + cid
	qr, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		fmt.Printf("❌ Failed to generate QR code: %v\n", err)
		return
	}

	fmt.Println("🔗 IPFS Link:", url)
	fmt.Println("📎 QR Code:")
	fmt.Println(qr.ToSmallString(false))
}
