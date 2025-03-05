package auth

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mr-tron/base58"
	"github.com/valyala/fasthttp"
)

// Keypair 구조체
type Keypair struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

// GenerateKeypair - Solana용 키페어 생성
func GenerateKeypair() Keypair {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatalf("키 생성 실패: %v", err)
	}
	return Keypair{PrivateKey: privateKey, PublicKey: publicKey}
}

// Sign - 메시지 서명 (Base64 인코딩)
func Sign(privateKey ed25519.PrivateKey, message string) string {
	signature := ed25519.Sign(privateKey, []byte(message))
	return base64.StdEncoding.EncodeToString(signature)
}

func (kp Keypair) String() string {
	return base58.Encode(kp.PublicKey)
}

func TestIsValidAuthorization(t *testing.T) {
	// Generate keypair for valid test case
	keypair := GenerateKeypair()
	pubKeyBase58 := keypair.String()

	// Create valid message with future timestamp
	futureTime := time.Now().Add(1 * time.Hour).Unix()
	validMessage := AUTH_MSG + ":" + strconv.FormatInt(futureTime, 10)

	// Sign the valid message
	signatureBase64 := Sign(keypair.PrivateKey, validMessage)

	fmt.Println("pubKeyBase58:", pubKeyBase58)
	fmt.Println("validMessage:", validMessage)
	fmt.Println("signatureBase64:", signatureBase64)

	// Create expired message and sign it
	expiredTime := time.Now().Add(-1 * time.Hour).Unix()
	expiredMessage := AUTH_MSG + ":" + strconv.FormatInt(expiredTime, 10)
	expiredSignature := Sign(keypair.PrivateKey, expiredMessage)

	// Create invalid format message and sign it
	invalidMessage := "invalid message"
	invalidSignature := Sign(keypair.PrivateKey, invalidMessage)

	// Create wrong auth message and sign it
	wrongAuthMessage := "wrong message:1234567890"
	wrongAuthSignature := Sign(keypair.PrivateKey, wrongAuthMessage)

	// Create invalid timestamp message and sign it
	invalidTimeMessage := AUTH_MSG + ":notanumber"
	invalidTimeSignature := Sign(keypair.PrivateKey, invalidTimeMessage)

	wrongSignature := "wrong signature"

	tests := []struct {
		name            string
		pubKeyBase58    string
		message         string
		signatureBase64 string
		want            bool
	}{
		{
			name:            "Valid signature and message",
			pubKeyBase58:    pubKeyBase58,
			message:         validMessage,
			signatureBase64: signatureBase64,
			want:            true,
		},
		{
			name:            "Invalid message format",
			pubKeyBase58:    pubKeyBase58,
			message:         invalidMessage,
			signatureBase64: invalidSignature,
			want:            false,
		},
		{
			name:            "Wrong auth message",
			pubKeyBase58:    pubKeyBase58,
			message:         wrongAuthMessage,
			signatureBase64: wrongAuthSignature,
			want:            false,
		},
		{
			name:            "Invalid timestamp format",
			pubKeyBase58:    pubKeyBase58,
			message:         invalidTimeMessage,
			signatureBase64: invalidTimeSignature,
			want:            false,
		},
		{
			name:            "Expired timestamp",
			pubKeyBase58:    pubKeyBase58,
			message:         expiredMessage,
			signatureBase64: expiredSignature,
			want:            false,
		},
		{
			name:            "Wrong signature",
			pubKeyBase58:    pubKeyBase58,
			message:         validMessage,
			signatureBase64: wrongSignature,
			want:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			req := app.AcquireCtx(&fasthttp.RequestCtx{})
			req.Request().Header.Set(HEADER_PUBKEY, tt.pubKeyBase58)
			req.Request().Header.Set(HEADER_MESSAGE, tt.message)
			req.Request().Header.Set(HEADER_SIGNATURE, tt.signatureBase64)

			_, err := CheckAuthorization(req)
			if tt.want {
				if err != nil {
					t.Errorf("CheckAuthorization() error = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Error("CheckAuthorization() error = nil, want error")
				}
			}
		})
	}
}
