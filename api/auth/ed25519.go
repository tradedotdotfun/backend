package auth

import (
	"crypto/ed25519"
	"encoding/base64"

	"github.com/mr-tron/base58"
)

func verifySignature(pubKeyBase58, message, signatureBase64 string) bool {
	pubKey, err := base58.Decode(pubKeyBase58)
	if err != nil {
		return false
	}

	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false
	}

	// Ed25519 서명 검증
	if isValid := ed25519.Verify(pubKey, []byte(message), signature); !isValid {
		return false
	}

	return true
}
