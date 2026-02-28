// Package idgen предоставляет функции для генерации уникальных идентификаторов.
//
// Идентификаторы используются для создания коротких URL в системе.
package idgen

import (
	"crypto/rand"
	"errors"
)

const (
	Charset     = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	IDLength    = 5
	charsetSize = len(Charset)
)

func GenerateID() (string, error) {
	result := make([]byte, IDLength)
	buffer := make([]byte, IDLength)

	_, err := rand.Read(buffer)
	if err != nil {
		return "", errors.New("failed to read random bytes: " + err.Error())
	}

	for i := 0; i < IDLength; i++ {
		result[i] = Charset[buffer[i]%byte(charsetSize)]
	}

	return string(result), nil
}
