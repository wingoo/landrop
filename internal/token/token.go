package token

import (
	"crypto/rand"
	"math/big"
)

const digits = "0123456789"

func Generate(n int) (string, error) {
	if n <= 0 {
		n = 4
	}
	out := make([]byte, n)
	max := big.NewInt(int64(len(digits)))
	for i := 0; i < n; i++ {
		v, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = digits[v.Int64()]
	}
	return string(out), nil
}
