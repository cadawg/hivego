package hivego

import (
	"bytes"

	"github.com/decred/base58"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	//lint:ignore SA1019 ripemd160 is used for checksums of public keys and is required for compatibility with Hive
	"golang.org/x/crypto/ripemd160"
)

// PublicKeyPrefix is the package-level default prefix for public key strings (e.g. "STM7...").
// This is used by [GetPublicKeyString] and [DecodePublicKey].
// Override per-client via [Client.PublicKeyPrefix]; set to "TST" for the Hive public testnet.
var PublicKeyPrefix = "STM"

// KeyPair holds a secp256k1 private/public key pair derived from a Hive WIF private key.
// Create one with [KeyPairFromWif] or [KeyPairFromBytes] and pass it to broadcast methods.
type KeyPair struct {
	PrivateKey *secp256k1.PrivateKey
	PublicKey  *secp256k1.PublicKey
}

// KeyPairFromWif creates a KeyPair from a WIF-encoded private key string.
func KeyPairFromWif(wif string) (*KeyPair, error) {
	privKey, _, err := GphBase58CheckDecode(wif)
	if err != nil {
		return nil, err
	}
	prvKey := secp256k1.PrivKeyFromBytes(privKey)
	return &KeyPair{prvKey, prvKey.PubKey()}, nil
}

// KeyPairFromBytes creates a KeyPair from a raw 32-byte private key.
func KeyPairFromBytes(privKeyBytes []byte) (*KeyPair, error) {
	if len(privKeyBytes) != 32 {
		return nil, ErrInvalidKeyLength
	}
	prvKey := secp256k1.PrivKeyFromBytes(privKeyBytes)
	return &KeyPair{prvKey, prvKey.PubKey()}, nil
}

// DecodePublicKey decodes a Hive public key string (e.g. "STM7...") to a secp256k1 public key.
// Uses the package-level PublicKeyPrefix; for a different prefix use DecodePublicKeyWithPrefix.
func DecodePublicKey(pubKey string) (*secp256k1.PublicKey, error) {
	return DecodePublicKeyWithPrefix(pubKey, PublicKeyPrefix)
}

// DecodePublicKeyWithPrefix decodes a Hive public key string with a specific prefix (e.g. "TST").
func DecodePublicKeyWithPrefix(pubKey, prefix string) (*secp256k1.PublicKey, error) {
	if len(pubKey) < len(prefix) || pubKey[:len(prefix)] != prefix {
		return nil, ErrInvalidPrefix
	}

	decoded := base58.Decode(pubKey[len(prefix):])
	if len(decoded) < 4 {
		return nil, ErrInvalidPublicKey
	}

	pubKeyBytes := decoded[:len(decoded)-4]
	checksum := decoded[len(decoded)-4:]

	hasher := ripemd160.New()
	if _, err := hasher.Write(pubKeyBytes); err != nil {
		return nil, err
	}
	if !bytes.Equal(checksum, hasher.Sum(nil)[:4]) {
		return nil, ErrChecksumMismatch
	}

	return secp256k1.ParsePubKey(pubKeyBytes)
}

// GetPublicKeyString returns the Hive public key string for a secp256k1 public key.
// Uses the package-level PublicKeyPrefix.
func GetPublicKeyString(pubKey *secp256k1.PublicKey) string {
	return GetPublicKeyStringWithPrefix(pubKey, PublicKeyPrefix)
}

// GetPublicKeyStringWithPrefix returns the Hive public key string with a specific prefix.
func GetPublicKeyStringWithPrefix(pubKey *secp256k1.PublicKey, prefix string) string {
	if pubKey == nil {
		return ""
	}

	pubKeyBytes := pubKey.SerializeCompressed()

	hasher := ripemd160.New()
	if _, err := hasher.Write(pubKeyBytes); err != nil {
		return ""
	}
	checksum := hasher.Sum(nil)[:4]

	pubKeyBytes = append(pubKeyBytes, checksum...)
	return prefix + base58.Encode(pubKeyBytes)
}

// GetPublicKeyString returns the Hive public key string using this KeyPair's public key.
// Uses the package-level PublicKeyPrefix.
func (kp *KeyPair) GetPublicKeyString() string {
	return GetPublicKeyString(kp.PublicKey)
}
