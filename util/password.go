package util

import (
	"crypto/rand"
	"math/big"
)

const defaultPasswordLength = 20

// GeneratePassword a random password
func GeneratePassword() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	lengthLetterRunes := big.NewInt(int64(len(letterRunes)))
	passwordRunes := make([]rune, defaultPasswordLength)

	for i := range passwordRunes {
		index, err := rand.Int(rand.Reader, lengthLetterRunes)
		if err != nil {
			panic(err)
		}

		passwordRunes[i] = letterRunes[index.Int64()]
	}

	return string(passwordRunes)
}
