package action

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"
)

// uuidNamespaces maps the namespace shortcuts accepted by uuid5 to their
// RFC 4122 namespace UUIDs.
var uuidNamespaces = map[string]string{
	"dns":  "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	"url":  "6ba7b811-9dad-11d1-80b4-00c04fd430c8",
	"oid":  "6ba7b812-9dad-11d1-80b4-00c04fd430c8",
	"x500": "6ba7b814-9dad-11d1-80b4-00c04fd430c8",
}

var uuid5Action = New(Definition[[]byte, []byte]{
	Doc:          "Create a deterministic UUID v5 from the input name, the namespace is dns, url, oid, x500 or a UUID",
	Names:        []string{"uuid5"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Parameters:   []ActionParameter{{StringParameter, "namespace: dns, url, oid, x500 or a UUID"}},
	Func: func(a Action, in []byte) ([]byte, error) {
		p, ok := a.InputParameters()[0].(string)
		if !ok {
			return nil, fmt.Errorf("namespace parameter is not a string")
		}
		ns, err := parseNamespace(p)
		if err != nil {
			return nil, err
		}
		return []byte(uuidV5(ns, in)), nil
	},
})

// parseNamespace resolves a namespace shortcut or a UUID to its 16 bytes.
func parseNamespace(p string) ([]byte, error) {
	p = strings.ToLower(strings.TrimSpace(p))
	if full, ok := uuidNamespaces[p]; ok {
		p = full
	}
	clean := strings.ReplaceAll(p, "-", "")
	if len(clean) != 32 {
		return nil, fmt.Errorf("%q is not a namespace: use dns, url, oid, x500 or a UUID", p)
	}
	b, err := hex.DecodeString(clean)
	if err != nil {
		return nil, fmt.Errorf("%q is not a UUID: %w", p, err)
	}
	return b, nil
}

// uuidV5 hashes the namespace and the name with SHA-1, then stamps the
// version 5 and RFC 4122 variant bits.
func uuidV5(namespace, name []byte) string {
	h := sha1.New()
	h.Write(namespace)
	h.Write(name)
	sum := h.Sum(nil)[:16]
	sum[6] = (sum[6] & 0x0f) | 0x50
	sum[8] = (sum[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", sum[0:4], sum[4:6], sum[6:8], sum[8:10], sum[10:16])
}

const (
	// tokenMaxBytes caps the random token size.
	tokenMaxBytes = 1024
	// bcryptMinCost and bcryptMaxCost bound the bcrypt work factor.
	bcryptMinCost = bcrypt.MinCost
	bcryptMaxCost = bcrypt.MaxCost
	// encryptSaltLen and encryptNonceLen size the sealed envelope.
	encryptSaltLen  = 16
	encryptNonceLen = 12
	// encryptPrefix versions the envelope so the format can evolve.
	encryptPrefix = "ovr1:"
	// scryptN is the work factor of the key derivation (about 100ms).
	scryptN = 1 << 15
	// rsaMinBits and rsaMaxBits bound the generated key size.
	rsaMinBits = 2048
	rsaMaxBits = 8192
)

var tokenAction = newGenerator(Definition[[]byte, []byte]{
	Doc:          "New random token, encoding is hex, base64 or base64url, e.g. 32 bytes as hex",
	Names:        []string{"token"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Parameters: []ActionParameter{
		{IntParameter, "number of random bytes, 1 to 1024"},
		{StringParameter, "encoding: hex, base64 or base64url"},
	},
	Func: func(a Action, in []byte) ([]byte, error) {
		params := a.InputParameters()
		n, ok := params[0].(int)
		if !ok {
			return nil, fmt.Errorf("size parameter is not an int")
		}
		encoding, _ := params[1].(string)
		if n < 1 || n > tokenMaxBytes {
			return nil, fmt.Errorf("token size must be between 1 and %d bytes", tokenMaxBytes)
		}

		b := make([]byte, n)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		switch strings.ToLower(strings.TrimSpace(encoding)) {
		case "hex":
			return []byte(hex.EncodeToString(b)), nil
		case "base64", "b64":
			return []byte(base64.StdEncoding.EncodeToString(b)), nil
		case "base64url", "b64url", "url":
			return []byte(base64.RawURLEncoding.EncodeToString(b)), nil
		default:
			return nil, fmt.Errorf("unknown encoding %q, use hex, base64 or base64url", encoding)
		}
	},
})

var newULAAction = newGenerator(Definition[[]byte, []byte]{
	Doc:          "New random IPv6 unique local prefix, fd00::/8 with a random 40-bit global ID, e.g. fd12:3456:789a::/48",
	Names:        []string{"newula"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		b := make([]byte, 5)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		return fmt.Appendf(nil, "fd%02x:%02x%02x:%02x%02x::/48", b[0], b[1], b[2], b[3], b[4]), nil
	},
})

var bcryptAction = New(Definition[[]byte, []byte]{
	Doc:              "Hash the input password with bcrypt, cost 4 to 31 (10 is a good default)",
	Names:            []string{"bcrypt"},
	Type:             TransformAction,
	InputFormat:      TextFormat,
	OutputFormat:     TextFormat,
	NonDeterministic: true,
	Parameters:       []ActionParameter{{IntParameter, "cost, 4 to 31"}},
	Func: func(a Action, in []byte) ([]byte, error) {
		cost, ok := a.InputParameters()[0].(int)
		if !ok {
			return nil, fmt.Errorf("cost parameter is not an int")
		}
		if cost < bcryptMinCost || cost > bcryptMaxCost {
			return nil, fmt.Errorf("cost must be between %d and %d", bcryptMinCost, bcryptMaxCost)
		}
		hash, err := bcrypt.GenerateFromPassword(in, cost)
		if err != nil {
			return nil, fmt.Errorf("can't hash the password: %w", err)
		}
		return hash, nil
	},
})

var bcryptCheckAction = New(Definition[[]byte, []byte]{
	Doc:          "Check a bcrypt hash against a password, the hash is the input, the answer is true or false",
	Names:        []string{"bcryptcheck"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Parameters:   []ActionParameter{{StringParameter, "the password to check"}},
	Func: func(a Action, in []byte) ([]byte, error) {
		p, ok := a.InputParameters()[0].(string)
		if !ok {
			return nil, fmt.Errorf("password parameter is not a string")
		}
		err := bcrypt.CompareHashAndPassword(in, []byte(p))
		if err == nil {
			return []byte("true"), nil
		}
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return []byte("false"), nil
		}
		return nil, fmt.Errorf("can't check the hash: %w", err)
	},
})

var encryptAction = New(Definition[[]byte, []byte]{
	Doc:              "Encrypt the input with AES-256-GCM and a passphrase, output is a versioned base64 envelope with its salt and nonce",
	Names:            []string{"encrypt"},
	Type:             TransformAction,
	InputFormat:      TextFormat,
	OutputFormat:     TextFormat,
	NonDeterministic: true,
	Parameters:       []ActionParameter{{StringParameter, "the passphrase"}},
	Func: func(a Action, in []byte) ([]byte, error) {
		pass, ok := a.InputParameters()[0].(string)
		if !ok {
			return nil, fmt.Errorf("passphrase parameter is not a string")
		}
		if pass == "" {
			return nil, fmt.Errorf("the passphrase can't be empty")
		}

		salt := make([]byte, encryptSaltLen)
		if _, err := rand.Read(salt); err != nil {
			return nil, err
		}
		key, err := scrypt.Key([]byte(pass), salt, scryptN, 8, 1, 32)
		if err != nil {
			return nil, fmt.Errorf("can't derive the key: %w", err)
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}
		nonce := make([]byte, encryptNonceLen)
		if _, err := rand.Read(nonce); err != nil {
			return nil, err
		}

		sealed := make([]byte, 0, encryptSaltLen+encryptNonceLen+len(in)+gcm.Overhead())
		sealed = append(sealed, salt...)
		sealed = append(sealed, nonce...)
		sealed = gcm.Seal(sealed, nonce, in, nil)

		return []byte(encryptPrefix + base64.StdEncoding.EncodeToString(sealed)), nil
	},
})

var decryptAction = New(Definition[[]byte, []byte]{
	Doc:          "Decrypt an envelope produced by encrypt with the same passphrase",
	Names:        []string{"decrypt"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Parameters:   []ActionParameter{{StringParameter, "the passphrase"}},
	Func: func(a Action, in []byte) ([]byte, error) {
		pass, ok := a.InputParameters()[0].(string)
		if !ok {
			return nil, fmt.Errorf("passphrase parameter is not a string")
		}

		s := strings.TrimSpace(string(in))
		rest, ok := strings.CutPrefix(s, encryptPrefix)
		if !ok {
			return nil, fmt.Errorf("input is not an encrypted envelope, it should start with %s", encryptPrefix)
		}
		sealed, err := base64.StdEncoding.DecodeString(rest)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 envelope: %w", err)
		}
		if len(sealed) < encryptSaltLen+encryptNonceLen {
			return nil, fmt.Errorf("the envelope is truncated")
		}

		salt := sealed[:encryptSaltLen]
		nonce := sealed[encryptSaltLen : encryptSaltLen+encryptNonceLen]
		ciphertext := sealed[encryptSaltLen+encryptNonceLen:]

		key, err := scrypt.Key([]byte(pass), salt, scryptN, 8, 1, 32)
		if err != nil {
			return nil, fmt.Errorf("can't derive the key: %w", err)
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}
		out, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return nil, fmt.Errorf("can't decrypt: wrong passphrase or corrupted data")
		}
		return out, nil
	},
})

var rsaAction = newGenerator(Definition[[]byte, []byte]{
	Doc:          "New RSA key pair as PEM, the private PKCS#8 key followed by the public key, bits 2048 to 8192",
	Names:        []string{"newrsa"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Parameters:   []ActionParameter{{IntParameter, "key size in bits, 2048 to 8192"}},
	Func: func(a Action, in []byte) ([]byte, error) {
		bits, ok := a.InputParameters()[0].(int)
		if !ok {
			return nil, fmt.Errorf("bits parameter is not an int")
		}
		if bits < rsaMinBits || bits > rsaMaxBits {
			return nil, fmt.Errorf("key size must be between %d and %d bits", rsaMinBits, rsaMaxBits)
		}
		key, err := rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			return nil, fmt.Errorf("can't generate the key: %w", err)
		}

		priv, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return nil, err
		}
		pub, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if err != nil {
			return nil, err
		}

		privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: priv})
		pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pub})
		return append(privPEM, pubPEM...), nil
	},
})
