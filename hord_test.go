package hord

import (
	"testing"
)

// TestValidations brought to you buy ChatGPT
func TestValidations(t *testing.T) {
	t.Run("ValidKey", func(t *testing.T) {
		// Valid keys
		validKeys := []string{"key1", "KEY_2", "3_key", "KEY-4", "key_5"}
		for _, key := range validKeys {
			err := ValidKey(key)
			if err != nil {
				t.Errorf("ValidKey(%s) returned error: %s, expected nil", key, err)
			}
		}
	})

	t.Run("InvalidKey", func(t *testing.T) {
		// Invalid keys
		invalidKeys := []string{""}
		for _, key := range invalidKeys {
			err := ValidKey(key)
			if err != ErrInvalidKey {
				t.Errorf("ValidKey(%s) returned error: %s, expected ErrInvalidKey", key, err)
			}
		}
	})

	t.Run("ValidData", func(t *testing.T) {
		// Valid data
		validData := [][]byte{{0x01}, {0x01, 0x02, 0x03}, {0xFF}}
		for _, data := range validData {
			err := ValidData(data)
			if err != nil {
				t.Errorf("ValidData(%v) returned error: %s, expected nil", data, err)
			}
		}
	})

	t.Run("InvalidData", func(t *testing.T) {
		// Invalid data
		invalidData := [][]byte{nil, {}}
		for _, data := range invalidData {
			err := ValidData(data)
			if err != ErrInvalidData {
				t.Errorf("ValidData(%v) returned error: %s, expected ErrInvalidData", data, err)
			}
		}
	})
}

