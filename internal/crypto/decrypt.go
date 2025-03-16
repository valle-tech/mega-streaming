package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
)

type KeyData struct {
	Key     []byte
	IV      []byte
	MetaMAC []byte
}

func ParseKeyFromBase64(keyStr string) (*KeyData, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, err
	}
	if len(decoded) != 32 {
		return nil, errors.New("invalid key length")
	}
	key := make([]byte, 16)
	iv := make([]byte, 8)
	metaMAC := make([]byte, 8)
	for i := range 16 {
		key[i] = decoded[i] ^ decoded[i+16]
	}
	copy(iv, decoded[16:24])
	copy(metaMAC, decoded[24:32])
	return &KeyData{Key: key, IV: iv, MetaMAC: metaMAC}, nil
}

func NewAESCTRReader(data []byte, keyData *KeyData, counterOffset int64) ([]byte, error) {
	block, err := aes.NewCipher(keyData.Key)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, aes.BlockSize)
	copy(iv, keyData.IV)
	iv = incrementIV(iv, counterOffset/16)
	stream := cipher.NewCTR(block, iv)
	decrypted := make([]byte, len(data))
	stream.XORKeyStream(decrypted, data)
	return decrypted, nil
}

func incrementIV(iv []byte, n int64) []byte {
	ivCopy := make([]byte, len(iv))
	copy(ivCopy, iv)
	for i := len(ivCopy) - 1; i >= 0 && n > 0; i-- {
		sum := int64(ivCopy[i]) + (n & 0xff)
		ivCopy[i] = byte(sum & 0xff)
		n >>= 8
		if sum > 255 {
			n++
		}
	}
	return ivCopy
}

func DecryptChunk(chunk []byte, keyData *KeyData, offset int64) ([]byte, error) {
	if len(chunk) == 0 {
		return nil, fmt.Errorf("empty chunk")
	}
	return NewAESCTRReader(chunk, keyData, offset)
}
