package utils

import (
	"encoding/base64"
	"fmt"
	qrcode "github.com/skip2/go-qrcode"
)

// GenerateQRCode generates a QR code for a frontend URL
func GenerateQRCode(frontendURL string) (string, error) {
	// Generate QR code as PNG
	qrCode, err := qrcode.Encode(frontendURL, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(qrCode), nil
}

// GenerateFallbackQRCode generates a fallback QR code URL when base64 generation fails
func GenerateFallbackQRCode(frontendURL string) string {
	return fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=%s", frontendURL)
}