package qr

import (
    "fmt"
    "github.com/skip2/go-qrcode"
)

func PrintIPFSQRCode(cid string) {
    url := fmt.Sprintf("https://ipfs.io/ipfs/%s", cid)
    qr, err := qrcode.New(url, qrcode.Medium)
    if err != nil {
        fmt.Printf("❌ Failed to generate QR code: %v\n", err)
        return
    }

    fmt.Println("🔗 IPFS Link:", url)
    fmt.Println("📎 QR Code:")
    fmt.Println(qr.ToSmallString(false))
}