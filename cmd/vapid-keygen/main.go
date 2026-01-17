package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func main() {
	// Generate ECDSA P-256 key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Printf("Error generating key: %v\n", err)
		return
	}

	// Encode private key (D value)
	privateKeyBytes := privateKey.D.Bytes()
	// Pad to 32 bytes if necessary
	if len(privateKeyBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(privateKeyBytes):], privateKeyBytes)
		privateKeyBytes = padded
	}
	privateKeyBase64 := base64.RawURLEncoding.EncodeToString(privateKeyBytes)

	// Encode public key (uncompressed point: 0x04 || X || Y)
	publicKeyBytes := make([]byte, 65)
	publicKeyBytes[0] = 0x04
	xBytes := privateKey.PublicKey.X.Bytes()
	yBytes := privateKey.PublicKey.Y.Bytes()
	// Pad X and Y to 32 bytes if necessary
	copy(publicKeyBytes[1+32-len(xBytes):33], xBytes)
	copy(publicKeyBytes[33+32-len(yBytes):65], yBytes)
	publicKeyBase64 := base64.RawURLEncoding.EncodeToString(publicKeyBytes)

	fmt.Println("VAPID Keys Generated:")
	fmt.Println()
	fmt.Println("Add these to your .env file:")
	fmt.Println()
	fmt.Printf("VAPID_PUBLIC_KEY=%s\n", publicKeyBase64)
	fmt.Printf("VAPID_PRIVATE_KEY=%s\n", privateKeyBase64)
	fmt.Println("VAPID_SUBJECT=mailto:admin@example.com")
}
