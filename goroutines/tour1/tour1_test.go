package tour1

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLastImplementedTypeName(t *testing.T) {
	h := sha256.New()
	h.Write([]byte(LastImplementedTypeName()))

	assert.Equal(t,
		"1aa4cb0bcca76e92e30677e809bb3d4b5c066715ef4d558184e319496bcc5125",
		hex.EncodeToString(h.Sum(nil)),
	)

}
