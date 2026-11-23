package install

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

func Digest(media io.Reader, rendered string) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, media); err != nil {
		return "", fmt.Errorf("hash installation media: %w", err)
	}
	mediaSum := h.Sum(nil)
	templateSum := sha256.Sum256([]byte(rendered))
	combined := sha256.Sum256(append(mediaSum, templateSum[:]...))
	return hex.EncodeToString(combined[:]), nil
}
