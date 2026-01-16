package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"sync"
)

var (
	encryptionKey []byte
	keyOnce       sync.Once
)

// getEncryptionKey derives a 32-byte key from the STALE_ENCRYPTION_KEY environment variable
// If not set, uses a default key (not recommended for production)
func getEncryptionKey() []byte {
	keyOnce.Do(func() {
		keyStr := os.Getenv("STALE_ENCRYPTION_KEY")
		if keyStr == "" {
			// Default key for development - NOT SECURE FOR PRODUCTION
			keyStr = "stale-default-encryption-key-change-me"
		}
		// Use SHA-256 to derive a 32-byte key
		hash := sha256.Sum256([]byte(keyStr))
		encryptionKey = hash[:]
	})
	return encryptionKey
}

// Encrypt encrypts plaintext using AES-GCM and returns base64-encoded ciphertext
func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext using AES-GCM
func Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	key := getEncryptionKey()
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		// If decoding fails, assume it's plaintext (for backward compatibility)
		return ciphertext, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		// Data too short, assume it's plaintext (backward compatibility)
		return ciphertext, nil
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		// Decryption failed, assume it's plaintext (backward compatibility)
		return ciphertext, nil
	}

	return string(plaintext), nil
}

// IsEncrypted checks if a string appears to be encrypted (base64 encoded with valid length)
func IsEncrypted(s string) bool {
	if s == "" {
		return false
	}
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return false
	}
	// AES-GCM with 12-byte nonce + at least 16-byte tag
	return len(data) >= 28
}

// MustEncrypt encrypts and panics on error (for initialization)
func MustEncrypt(plaintext string) string {
	encrypted, err := Encrypt(plaintext)
	if err != nil {
		panic(err)
	}
	return encrypted
}

// ErrDecryptionFailed is returned when decryption fails
var ErrDecryptionFailed = errors.New("decryption failed")
