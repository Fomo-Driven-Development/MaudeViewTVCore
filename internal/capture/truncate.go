package capture

import (
	"crypto/sha256"
	"encoding/hex"
)

func truncateBytes(in []byte, maxBytes int) ([]byte, bool, int, string) {
	if maxBytes <= 0 || len(in) <= maxBytes {
		return in, false, len(in), ""
	}
	sum := sha256.Sum256(in)
	return in[:maxBytes], true, len(in), hex.EncodeToString(sum[:])
}
