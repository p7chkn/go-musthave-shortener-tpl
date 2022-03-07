// Package utils - пакет вспомогательных инструментов.
package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"

	"github.com/gofrs/uuid"
)

// Encryptor - стрктура для шифрования/расшифрование.
type Encryptor struct {
	aesblock cipher.Block
	key      []byte
}

// New - создание новой структуры Encryptor.
func New(key []byte) (*Encryptor, error) {
	enc := Encryptor{
		key: key,
	}
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	enc.aesblock = aesblock
	return &enc, nil
}

// EncodeUUIDtoString - зашифровка UUID в строку.
func (en *Encryptor) EncodeUUIDtoString(value []byte) string {

	encrypted := make([]byte, aes.BlockSize)
	en.aesblock.Encrypt(encrypted, value)

	return hex.EncodeToString(encrypted)
}

// DecodeUUIDFromString - расшифровка строки в UUID.
func (en *Encryptor) DecodeUUIDFromString(value string) (string, error) {
	encrypted, err := hex.DecodeString(value)
	if err != nil {
		return "", err
	}
	decrypted := make([]byte, aes.BlockSize)
	en.aesblock.Decrypt(decrypted, encrypted)
	result, err := uuid.FromBytes(decrypted)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}
