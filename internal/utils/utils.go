package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"

	"github.com/gofrs/uuid"
)

type Encryptor struct {
	aesblock cipher.Block
	key      []byte
}

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

func (en *Encryptor) EncodeUUIDtoString(value []byte) string {

	encrypted := make([]byte, aes.BlockSize)
	en.aesblock.Encrypt(encrypted, value)

	return hex.EncodeToString(encrypted)
}

func (en *Encryptor) DecodeUUIDfromString(value string) (string, error) {
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
