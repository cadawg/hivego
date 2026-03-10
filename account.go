package hivego

import "encoding/json"

// GetAccounts returns account data for the given list of account names.
// Balance fields on AccountData are strings (e.g. "1.000 HIVE"); use ParseAsset to convert.
func (d DatabaseAPI) GetAccounts(names []string) ([]AccountData, error) {
	q := hrpcQuery{
		method: "condenser_api.get_accounts",
		params: []interface{}{names},
	}
	res, err := d.client.rpcExecWithFailover(q)
	if err != nil {
		return nil, err
	}

	var accounts []AccountData
	if err := json.Unmarshal(res, &accounts); err != nil {
		return nil, err
	}
	return accounts, nil
}
