package auth

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const AUTH_MSG = "sign in tradedot.fun"
const HEADER_PUBKEY = "X-Auth-Pubkey"
const HEADER_MESSAGE = "X-Auth-Message"
const HEADER_SIGNATURE = "X-Auth-Signature"

func CheckAuthorization(c *fiber.Ctx) (string, error) {
	pubKeyBase58 := c.Get(HEADER_PUBKEY)
	if pubKeyBase58 == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Missing authorization token")
	}
	message := c.Get(HEADER_MESSAGE)
	if message == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Missing message for verification")
	}
	signatureBase64 := c.Get(HEADER_SIGNATURE)
	if signatureBase64 == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Missing signature for verification")
	}
	// Split message into base message and timestamp
	parts := strings.Split(message, ":")
	if len(parts) != 2 {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Invalid message format")
	}

	if parts[0] != AUTH_MSG {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Invalid message format")
	}

	// Parse timestamp
	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Invalid timestamp")
	}

	// Check if timestamp is expired (current time > timestamp)
	if time.Now().Unix() > timestamp {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Expired timestamp")
	}

	if success := verifySignature(pubKeyBase58, message, signatureBase64); !success {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Invalid signature")
	}

	return pubKeyBase58, nil
}
