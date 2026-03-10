package hivego

import (
	"encoding/hex"
)

// Transaction is a Hive blockchain transaction, ready for signing and broadcasting.
// Use BuildTransaction to create one from a set of operations, or construct it directly
// for offline/multi-sig workflows.
type Transaction struct {
	RefBlockNum    uint16           `json:"ref_block_num"`
	RefBlockPrefix uint32           `json:"ref_block_prefix"`
	Expiration     string           `json:"expiration"`
	Operations     []HiveOperation  `json:"-"`
	OperationsJs   [][2]interface{} `json:"operations"`
	Extensions     []string         `json:"extensions"`
	Signatures     []string         `json:"signatures"`
}

// GenerateTrxId computes the transaction ID (first 20 bytes of SHA256 of the serialized tx).
func (t *Transaction) GenerateTrxId() (string, error) {
	tB, err := serializeTx(*t)
	if err != nil {
		return "", err
	}
	digest := hashTx(tB)
	return hex.EncodeToString(digest)[0:40], nil
}

func (t *Transaction) prepareJson() {
	var opsContainer [][2]interface{}
	for _, op := range t.Operations {
		var opContainer [2]interface{}
		opContainer[0] = op.OpName()
		opContainer[1] = op
		opsContainer = append(opsContainer, opContainer)
	}
	if t.Extensions == nil {
		t.Extensions = []string{}
	}
	t.OperationsJs = opsContainer
}

// BuildTransaction fetches the current chain state and returns an unsigned Transaction
// populated with the correct reference block and expiration.
func (h *Client) BuildTransaction(ops []HiveOperation) (*Transaction, error) {
	signingData, err := h.getSigningData()
	if err != nil {
		return nil, err
	}
	return &Transaction{
		RefBlockNum:    signingData.refBlockNum,
		RefBlockPrefix: signingData.refBlockPrefix,
		Expiration:     signingData.expiration,
		Operations:     ops,
	}, nil
}

// Sign signs the transaction with the given WIF key and appends the signature.
// The client's ChainID is used (Hive mainnet by default; override with client.ChainID).
// Does not broadcast — call BroadcastRaw when ready.
func (h *Client) Sign(tx *Transaction, wif string) error {
	message, err := serializeTx(*tx)
	if err != nil {
		return err
	}
	digest := hashTxForSig(message, h.chainIDBytes())
	sig, err := SignDigest(digest, wif)
	if err != nil {
		return err
	}
	tx.Signatures = append(tx.Signatures, hex.EncodeToString(sig))
	return nil
}

// BroadcastRaw submits a pre-built, pre-signed Transaction to the network.
func (h *Client) BroadcastRaw(tx *Transaction) error {
	tx.prepareJson()
	var params []interface{}
	params = append(params, tx)
	q := hrpcQuery{"condenser_api.broadcast_transaction", params}
	_, err := h.rpcExecWithFailover(q)
	return err
}

func (h *Client) broadcast(ops []HiveOperation, wif string) (string, error) {
	tx, err := h.BuildTransaction(ops)
	if err != nil {
		return "", err
	}

	txId, err := tx.GenerateTrxId()
	if err != nil {
		return "", err
	}

	if err := h.Sign(tx, wif); err != nil {
		return "", err
	}

	if !h.NoBroadcast {
		if err := h.BroadcastRaw(tx); err != nil {
			return "", err
		}
	}

	return txId, nil
}
