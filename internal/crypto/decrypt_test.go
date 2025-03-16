package crypto

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestParseKeyFromBase64_Valid(t *testing.T) {
	keyStr := base64.RawURLEncoding.EncodeToString(make([]byte, 32))
	keyData, err := ParseKeyFromBase64(keyStr)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(keyData.Key) != 16 || len(keyData.IV) != 8 || len(keyData.MetaMAC) != 8 {
		t.Fatal("parsed key structure has incorrect lengths")
	}
}

func TestParseKeyFromBase64_InvalidLength(t *testing.T) {
	keyStr := base64.RawURLEncoding.EncodeToString(make([]byte, 16))
	_, err := ParseKeyFromBase64(keyStr)
	if err == nil {
		t.Fatal("expected error for invalid key length")
	}
}

func TestParseKeyFromBase64_InvalidBase64(t *testing.T) {
	_, err := ParseKeyFromBase64("!!!notbase64$$$")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestNewAESCTRReader_BasicDecryption(t *testing.T) {
	key := bytes.Repeat([]byte{0x01}, 16)
	iv := bytes.Repeat([]byte{0x00}, 8)
	metaMAC := bytes.Repeat([]byte{0x00}, 8)
	keyData := &KeyData{Key: key, IV: iv, MetaMAC: metaMAC}

	plaintext := []byte("test-encrypted-content")
	encrypted := make([]byte, len(plaintext))

	block, _ := NewAESCTRReader(plaintext, keyData, 0)
	copy(encrypted, block)

	decrypted, err := NewAESCTRReader(encrypted, keyData, 0)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decryption mismatch, got %s, want %s", decrypted, plaintext)
	}
}

func TestNewAESCTRReader_EmptyInput(t *testing.T) {
	key := bytes.Repeat([]byte{0x01}, 16)
	keyData := &KeyData{Key: key, IV: bytes.Repeat([]byte{0x00}, 8), MetaMAC: []byte{}}

	result, err := NewAESCTRReader([]byte{}, keyData, 0)
	if err != nil {
		t.Fatalf("decryption should not fail for empty input: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestDecryptChunk_Basic(t *testing.T) {
	key := bytes.Repeat([]byte{0x01}, 16)
	iv := bytes.Repeat([]byte{0x00}, 8)
	metaMAC := bytes.Repeat([]byte{0x00}, 8)
	keyData := &KeyData{Key: key, IV: iv, MetaMAC: metaMAC}

	content := []byte("mega-encrypted-content")
	encrypted, _ := NewAESCTRReader(content, keyData, 0)

	decrypted, err := DecryptChunk(encrypted, keyData, 0)
	if err != nil {
		t.Fatalf("decrypt chunk failed: %v", err)
	}
	if !bytes.Equal(decrypted, content) {
		t.Errorf("decrypt chunk mismatch, got %s, want %s", decrypted, content)
	}
}

func TestDecryptChunk_Empty(t *testing.T) {
	keyData := &KeyData{Key: []byte("1234567890123456"), IV: []byte("12345678")}
	_, err := DecryptChunk([]byte{}, keyData, 0)
	if err == nil {
		t.Fatal("expected error on empty chunk")
	}
}

func TestIncrementIV_Correctness(t *testing.T) {
	iv := []byte{0, 0, 0, 0, 0, 0, 0, 1}
	result := incrementIV(iv, 1)
	expected := []byte{0, 0, 0, 0, 0, 0, 0, 2}
	if !bytes.Equal(result, expected) {
		t.Errorf("IV increment mismatch, got %v, want %v", result, expected)
	}

	result = incrementIV(iv, 256)
	expected = []byte{0, 0, 0, 0, 0, 0, 1, 1}
	if !bytes.Equal(result, expected) {
		t.Errorf("IV increment mismatch, got %v, want %v", result, expected)
	}
}
