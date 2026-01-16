package util

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
	}{
		{"empty string", ""},
		{"simple text", "hello world"},
		{"special chars", "token_123!@#$%^&*()"},
		{"unicode", "日本語テスト"},
		{"long token", "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			// Empty string should return empty
			if tt.plaintext == "" && encrypted != "" {
				t.Errorf("Encrypt() of empty string should return empty, got %v", encrypted)
			}

			// Non-empty should be different from plaintext
			if tt.plaintext != "" && encrypted == tt.plaintext {
				t.Errorf("Encrypt() should not return plaintext")
			}

			decrypted, err := Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestDecryptPlaintext(t *testing.T) {
	// Decrypt should handle plaintext gracefully (backward compatibility)
	plaintext := "this-is-not-encrypted"
	result, err := Decrypt(plaintext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if result != plaintext {
		t.Errorf("Decrypt() = %v, want %v", result, plaintext)
	}
}

func TestIsEncrypted(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty", "", false},
		{"plaintext", "not-encrypted", false},
		{"short base64", "YWJj", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEncrypted(tt.input); got != tt.expected {
				t.Errorf("IsEncrypted() = %v, want %v", got, tt.expected)
			}
		})
	}

	// Test with actual encrypted value
	encrypted, _ := Encrypt("test-token")
	if !IsEncrypted(encrypted) {
		t.Error("IsEncrypted() should return true for encrypted value")
	}
}

func TestMustEncrypt(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustEncrypt() should not panic for valid input")
		}
	}()

	result := MustEncrypt("test-value")
	if result == "" || result == "test-value" {
		t.Errorf("MustEncrypt() should return encrypted value")
	}
}
