package action

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAction_UUID5(t *testing.T) {
	r := NewRegistry()

	// RFC 4122 example: uuid5(DNS, "www.example.com")
	got, err := textWithParams(t, r, "uuid5", []byte("www.example.com"), "dns")
	require.NoError(t, err)
	require.Equal(t, "2ed6657d-e927-568b-95e1-2665a8aea6a2", string(got))

	// deterministic
	again, err := textWithParams(t, r, "uuid5", []byte("www.example.com"), "DNS")
	require.NoError(t, err)
	require.Equal(t, got, again)

	// explicit namespace UUID
	custom, err := textWithParams(t, r, "uuid5", []byte("x"), "6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	require.NoError(t, err)
	require.Len(t, string(custom), 36)

	_, err = textWithParams(t, r, "uuid5", []byte("x"), "nope")
	require.Error(t, err)
}

func TestAction_Bcrypt(t *testing.T) {
	r := NewRegistry()

	hash, err := textWithParams(t, r, "bcrypt", []byte("s3cret"), 4)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(hash), "$2"))

	ok, err := textWithParams(t, r, "bcryptcheck", hash, "s3cret")
	require.NoError(t, err)
	require.Equal(t, "true", string(ok))

	ko, err := textWithParams(t, r, "bcryptcheck", hash, "wrong")
	require.NoError(t, err)
	require.Equal(t, "false", string(ko))

	_, err = textWithParams(t, r, "bcrypt", []byte("x"), 3)
	require.Error(t, err)
}

func TestAction_EncryptDecrypt(t *testing.T) {
	r := NewRegistry()

	msg := []byte("attack at dawn")
	sealed, err := textWithParams(t, r, "encrypt", msg, "correct horse")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(sealed), "ovr1:"))

	back, err := textWithParams(t, r, "decrypt", sealed, "correct horse")
	require.NoError(t, err)
	require.Equal(t, msg, back)

	// a second encryption of the same input uses a fresh salt and nonce
	again, err := textWithParams(t, r, "encrypt", msg, "correct horse")
	require.NoError(t, err)
	require.NotEqual(t, sealed, again)

	_, err = textWithParams(t, r, "decrypt", sealed, "wrong")
	require.Error(t, err)

	_, err = textWithParams(t, r, "decrypt", []byte("not-an-envelope"), "x")
	require.Error(t, err)
}

func TestAction_Token(t *testing.T) {
	r := NewRegistry()

	got, err := textWithParams(t, r, "token", nil, 16, "hex")
	require.NoError(t, err)
	require.Len(t, string(got), 32)
	_, err = hex.DecodeString(string(got))
	require.NoError(t, err)

	got, err = textWithParams(t, r, "token", nil, 16, "base64url")
	require.NoError(t, err)
	raw, err := base64.RawURLEncoding.DecodeString(string(got))
	require.NoError(t, err)
	require.Len(t, raw, 16)

	_, err = textWithParams(t, r, "token", nil, 16, "rot13")
	require.Error(t, err)
	_, err = textWithParams(t, r, "token", nil, 0, "hex")
	require.Error(t, err)
}

func TestAction_NewULA(t *testing.T) {
	r := NewRegistry()

	got, err := r.TextAction("newula", nil)
	require.NoError(t, err)
	require.Regexp(t, regexp.MustCompile(`^fd[0-9a-f]{2}(:[0-9a-f]{4}){2}::/48$`), string(got))
}

func TestAction_NewRSA(t *testing.T) {
	r := NewRegistry()

	got, err := textWithParams(t, r, "newrsa", nil, 2048)
	require.NoError(t, err)

	var types []string
	rest := got
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		types = append(types, block.Type)
		switch block.Type {
		case "PRIVATE KEY":
			_, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			require.NoError(t, err)
		case "PUBLIC KEY":
			_, err := x509.ParsePKIXPublicKey(block.Bytes)
			require.NoError(t, err)
		}
	}
	require.Equal(t, []string{"PRIVATE KEY", "PUBLIC KEY"}, types)

	_, err = textWithParams(t, r, "newrsa", nil, 512)
	require.Error(t, err)
}
