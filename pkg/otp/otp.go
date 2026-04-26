package otp

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
)

func GenerateOtp() (string, string, error) {
	// 1. Generate "securityly random" from 0 to 899999
	n, err := rand.Int(rand.Reader, big.NewInt(900000))

	if err != nil {
		return "", "", err
	}

	// 2. add offset, so the range become 100000 - 999999
	plain := fmt.Sprintf("%06d", n.Int64()+100000)

	// 3. hash the generated number
	hash := sha256.Sum256([]byte(plain))

	// 4. turn the 32bytes generated to hex string
	hashed := fmt.Sprintf("%x", hash)

	return plain, hashed, nil
}
