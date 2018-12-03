package util

import (
	"math/rand"
	"time"
)

// EightRandomLetters generates a random string
func EightRandomLetters() string {
	rand.Seed(time.Now().UTC().UnixNano())
	letterBytes := "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
