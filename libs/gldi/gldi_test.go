package gldi_test

import (
	"github.com/stretchr/testify/assert"

	"github.com/sqp/godock/libs/gldi"

	"testing"
)

func TestCrypto(t *testing.T) {
	for _, str := range []string{
		"/usr/share/cairo-dock/cairo-dock.conf",
		"http://glx-dock.org/index.php",
	} {
		out := gldi.Crypto.EncryptString(str)
		assert.Equal(t, str, gldi.Crypto.DecryptString(out), "Encrypt/Decrypt")
	}
}
