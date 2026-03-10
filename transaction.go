package hivego

type transactionQueryParams struct {
	TransactionId     string `json:"id"`
	IncludeReversible bool   `json:"include_reversible"`
}

// GetTransaction fetches a transaction by ID, returning raw JSON.
func (d DatabaseAPI) GetTransaction(txId string, includeReversible bool) ([]byte, error) {
	q := hrpcQuery{
		method: "account_history_api.get_transaction",
		params: transactionQueryParams{TransactionId: txId, IncludeReversible: includeReversible},
	}
	return d.client.rpcExecWithFailover(q)
}
