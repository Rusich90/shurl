package idgen

import (
	"crypto/rand"
	"math/big"

	"github.com/Rusich90/shurl.git/internal/config"
)

func GenerateID() string {
	result := make([]byte, config.IDLength)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(config.Charset))))
		if err != nil {
			panic(err)
		}
		result[i] = config.Charset[n.Int64()]
	}
	return string(result)
}
