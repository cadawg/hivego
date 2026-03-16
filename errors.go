package hivego

import (
	"errors"
	"fmt"
	"strconv"
)

// Sentinel errors returned by hivego functions.
// Use [errors.Is] to check for a specific condition and [errors.As] to extract a structured
// [RPCError].
var (
	// ErrNilKey is returned when a nil *KeyPair is passed to Sign or a broadcast method.
	ErrNilKey = errors.New("key must not be nil")
	// ErrNilMemoKey is returned when a nil memo public key is passed to an operation that requires one.
	ErrNilMemoKey = errors.New("memo key must not be nil")
	// ErrInvalidKeyLength is returned by [KeyPairFromBytes] when the input is not exactly 32 bytes.
	ErrInvalidKeyLength = errors.New("private key must be 32 bytes")
	// ErrInvalidPrefix is returned by [DecodePublicKeyWithPrefix] when the key string does not start with the expected prefix.
	ErrInvalidPrefix = errors.New("invalid public key prefix")
	// ErrInvalidPublicKey is returned when the public key bytes cannot be parsed.
	ErrInvalidPublicKey = errors.New("invalid public key")
	// ErrChecksumMismatch is returned when a WIF or public key checksum does not match.
	ErrChecksumMismatch = errors.New("checksum mismatch")
	// ErrInvalidFormat is returned when a WIF string is too short or malformed.
	ErrInvalidFormat = errors.New("invalid format")
	// ErrInvalidAsset is returned by [ParseAsset] when the input string cannot be parsed.
	ErrInvalidAsset = errors.New("invalid asset")

	// ErrDeclineVotingRightsNotConfirmed is returned when [DeclineVotingRightsOperation].Decline
	// is true but IUnderstandThisIsIrreversible is not set. Declining voting rights permanently
	// removes the account's ability to vote and cannot be undone.
	ErrDeclineVotingRightsNotConfirmed = errors.New("set IUnderstandThisIsIrreversible: true to confirm — declining voting rights is permanent and cannot be undone")
)

// RPCError is returned when the Hive node responds with a JSON-RPC error.
// Use [errors.As] to extract the Code and Message:
//
//	var rpcErr *hivego.RPCError
//	if errors.As(err, &rpcErr) {
//	    fmt.Println(rpcErr.Code, rpcErr.Message)
//	}
type RPCError struct {
	Code    int
	Message string
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("rpc error %s: %s", strconv.Itoa(e.Code), e.Message)
}
