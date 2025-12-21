package idgen

import (
	"crypto/rand"
	"errors"
	"math/big"
)

const (
	Charset  = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	IDLength = 5
)

func GenerateID() (string, error) {
	result := make([]byte, IDLength)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(Charset))))
		if err != nil {
			return "", errors.New("failed to generate random number: " + err.Error())
		}
		result[i] = Charset[n.Int64()]
	}
	return string(result), nil
}
