package capture

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestTruncateBytes(t *testing.T) {
	t.Run("no_truncation_when_within_limit", func(t *testing.T) {
		input := []byte("hello world")
		out, truncated, origLen, hash := truncateBytes(input, len(input))

		if truncated {
			t.Fatalf("expected truncated=false, got true")
		}
		if origLen != len(input) {
			t.Fatalf("expected original size %d, got %d", len(input), origLen)
		}
		if hash != "" {
			t.Fatalf("expected empty hash, got %q", hash)
		}
		if string(out) != string(input) {
			t.Fatalf("expected output %q, got %q", string(input), string(out))
		}
	})

	t.Run("truncate_large_slice", func(t *testing.T) {
		input := []byte("hello world")
		maxBytes := 5
		expectedHash := sha256.Sum256(input)
		out, truncated, origLen, hash := truncateBytes(input, maxBytes)

		if !truncated {
			t.Fatalf("expected truncated=true, got false")
		}
		if origLen != len(input) {
			t.Fatalf("expected original size %d, got %d", len(input), origLen)
		}
		if string(out) != "hello" {
			t.Fatalf("expected output %q, got %q", "hello", string(out))
		}
		if hash != hex.EncodeToString(expectedHash[:]) {
			t.Fatalf("unexpected hash %q", hash)
		}
	})
}

func TestTruncateStringBytes(t *testing.T) {
	t.Run("delegates_to_shared_byte_truncator", func(t *testing.T) {
		input := "hello world"
		maxBytes := 5
		expected, expectedTruncated, expectedLen, expectedHash := truncateBytes([]byte(input), maxBytes)

		out, truncated, origLen, hash := truncateStringBytes(input, maxBytes)

		if out != string(expected) {
			t.Fatalf("expected output %q, got %q", string(expected), out)
		}
		if truncated != expectedTruncated {
			t.Fatalf("expected truncated=%v, got %v", expectedTruncated, truncated)
		}
		if origLen != expectedLen {
			t.Fatalf("expected original size %d, got %d", expectedLen, origLen)
		}
		if hash != expectedHash {
			t.Fatalf("expected hash %q, got %q", expectedHash, hash)
		}
	})

	t.Run("non_ascii_is_truncated_by_bytes", func(t *testing.T) {
		input := "😀😀" // each rune is 4 bytes
		out, _, _, _ := truncateStringBytes(input, 5)
		if len([]byte(out)) != 5 {
			t.Fatalf("expected byte length 5, got %d", len([]byte(out)))
		}
	})
}
