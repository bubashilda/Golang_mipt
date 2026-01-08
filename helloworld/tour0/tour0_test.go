package tour0

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLastFunctionName(t *testing.T) {
	h := sha256.New()
	h.Write([]byte(LastFunctionName()))

	assert.Equal(t,
		"9e258d1804965ab218e17023f248a3e1efe0ef3f8fd0693938dcd76433e96c09",
		hex.EncodeToString(h.Sum(nil)),
	)
}
