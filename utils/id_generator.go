package utils

import (
	"math/rand"
	"time"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
const charset1 = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func userIdGenerate() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset1[seededRand.Intn(len(charset1))]
	}

	return string(b)
}

func IdGenerate(length int, withFri bool) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	var prefix string
	if withFri {
		prefix = "F"
	} else {
		prefix = "O"
	}
	return prefix + string(b)
}
