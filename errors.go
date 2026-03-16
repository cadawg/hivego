package hivego

import (
	"errors"
	"fmt"
	"strconv"
)

// Sentinel errors returned by hivego functions.
// Use errors.Is to check for specific error conditions,
// and errors.As to extract structured error details (e.g. RPCError).
var (
	ErrNilKey           = errors.New("key must not be nil")
	ErrNilMemoKey       = errors.New("memo key must not be nil")
	ErrInvalidKeyLength = errors.New("private key must be 32 bytes")
	ErrInvalidPrefix    = errors.New("invalid public key prefix")
	ErrInvalidPublicKey = errors.New("invalid public key")
	ErrChecksumMismatch = errors.New("checksum mismatch")
	ErrInvalidFormat    = errors.New("invalid format")
	ErrInvalidAsset     = errors.New("invalid asset")

	// ErrDeclineVotingRightsNotConfirmed is returned when DeclineVotingRightsOperation.Decline
	// is true but IUnderstandThisIsIrreversible is not set. Declining voting rights permanently
	// removes the account's ability to vote and cannot be undone.
	ErrDeclineVotingRightsNotConfirmed = errors.New("set IUnderstandThisIsIrreversible: true to confirm — declining voting rights is permanent and cannot be undone")
)

// RPCError is returned when the Hive node returns a JSON-RPC error response.
// Use errors.As to access the Code and Message fields.
type RPCError struct {
	Code    int
	Message string
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("rpc error %s: %s", strconv.Itoa(e.Code), e.Message)
}
