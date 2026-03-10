package hivego

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/decred/base58"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

type signingDataFromChain struct {
	refBlockNum    uint16
	refBlockPrefix uint32
	expiration     string
}

func (h *Client) getSigningData() (signingDataFromChain, error) {
	q := hrpcQuery{method: "condenser_api.get_dynamic_global_properties", params: []string{}}
	propsB, err := h.rpcExecWithFailover(q)
	if err != nil {
		return signingDataFromChain{}, err
	}

	var props globalProps
	err = json.Unmarshal(propsB, &props)
	if err != nil {
		return signingDataFromChain{}, err
	}

	refBlockNum := uint16(props.HeadBlockNumber & 0xffff)
	hbidB, err := hex.DecodeString(props.HeadBlockId)
	if err != nil {
		return signingDataFromChain{}, err
	}
	refBlockPrefix := binary.LittleEndian.Uint32(hbidB[4:])

	exp, err := time.Parse("2006-01-02T15:04:05", props.Time)
	if err != nil {
		return signingDataFromChain{}, err
	}
	exp = exp.Add(30 * time.Second)
	expStr := exp.Format("2006-01-02T15:04:05")

	return signingDataFromChain{refBlockNum, refBlockPrefix, expStr}, nil
}

// hashTxForSig computes the signing digest: SHA256(chainID + serialized tx bytes).
func hashTxForSig(tx []byte, chainID []byte) []byte {
	var message bytes.Buffer
	message.Write(chainID)
	message.Write(tx)

	digest := sha256.New()
	digest.Write(message.Bytes())
	return digest.Sum(nil)
}

func hashTx(tx []byte) []byte {
	digest := sha256.New()
	digest.Write(tx)
	return digest.Sum(nil)
}

// SignDigest signs a digest with the given WIF private key using secp256k1 compact signing.
func SignDigest(digest []byte, wif string) ([]byte, error) {
	keyPair, err := KeyPairFromWif(wif)
	if err != nil {
		return nil, err
	}
	return ecdsa.SignCompact(keyPair.PrivateKey, digest, true), nil
}

// GphBase58CheckDecode decodes a Graphene/Hive base58-encoded string with a double-SHA256 checksum.
// Returns the payload bytes, the version byte, and any error.
func GphBase58CheckDecode(input string) ([]byte, [1]byte, error) {
	decoded := base58.Decode(input)
	if len(decoded) < 6 {
		return nil, [1]byte{0}, errors.New("invalid format: version and/or checksum bytes missing")
	}
	version := [1]byte{decoded[0]}
	dataLen := len(decoded) - 4
	decodedChecksum := decoded[dataLen:]
	calculatedChecksum := checksum(decoded[:dataLen])
	if !bytes.Equal(decodedChecksum, calculatedChecksum[:]) {
		return nil, [1]byte{0}, errors.New("checksum error")
	}
	payload := decoded[1:dataLen]
	return payload, version, nil
}

// GphBase58Encode encodes a payload with a version byte into a Graphene/Hive base58 string
// with a double-SHA256 checksum. This is the inverse of GphBase58CheckDecode.
func GphBase58Encode(payload []byte, version [1]byte) string {
	input := append([]byte{version[0]}, payload...)
	cs := checksum(input)
	input = append(input, cs[:]...)
	return base58.Encode(input)
}

func checksum(input []byte) [4]byte {
	var calculatedChecksum [4]byte
	intermediateHash := sha256.Sum256(input)
	finalHash := sha256.Sum256(intermediateHash[:])
	copy(calculatedChecksum[:], finalHash[:])
	return calculatedChecksum
}
