package token

import (
	"crypto/rand"
	"math/big"
)

const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"

func Generate(n int) (string, error) {
	if n <= 0 {
		l, err := rand.Int(rand.Reader, big.NewInt(5))
		if err != nil {
			return "", err
		}
		n = 6 + int(l.Int64())
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, n)
	for i, b := range buf {
		out[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(out), nil
}
