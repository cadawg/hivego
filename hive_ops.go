package hivego

// HiveOperation is the interface implemented by all Hive blockchain operations.
// Implement this interface to create custom operations for use with BuildTransaction.
type HiveOperation interface {
	SerializeOp() ([]byte, error)
	OpName() string
}

// VoteOperation votes on a post or comment.
// Weight ranges from -10000 (100% downvote) to 10000 (100% upvote).
type VoteOperation struct {
	Voter    string `json:"voter"`
	Author   string `json:"author"`
	Permlink string `json:"permlink"`
	Weight   int16  `json:"weight"`
}

func (o VoteOperation) OpName() string { return "vote" }

// CustomJsonOperation broadcasts a custom_json operation.
// Use RequiredAuths for active-key actions and RequiredPostingAuths for posting-key actions.
type CustomJsonOperation struct {
	RequiredAuths        []string `json:"required_auths"`
	RequiredPostingAuths []string `json:"required_posting_auths"`
	Id                   string   `json:"id"`
	Json                 string   `json:"json"`
}

func (o CustomJsonOperation) OpName() string { return "custom_json" }

// TransferOperation transfers HIVE or HBD between accounts.
type TransferOperation struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount Asset  `json:"amount"`
	Memo   string `json:"memo"`
}

func (o TransferOperation) OpName() string { return "transfer" }

// ClaimRewardBalanceOperation claims pending reward balances into the account.
type ClaimRewardBalanceOperation struct {
	Account     string `json:"account"`
	RewardHive  Asset  `json:"reward_hive"`
	RewardHbd   Asset  `json:"reward_hbd"`
	RewardVests Asset  `json:"reward_vests"`
}

func (o ClaimRewardBalanceOperation) OpName() string { return "claim_reward_balance" }

// Vote submits a vote on a post or comment.
// weight ranges from -10000 (100% downvote) to 10000 (100% upvote).
func (b BroadcastAPI) Vote(voter, author, permlink string, weight int, wif string) (string, error) {
	op := VoteOperation{voter, author, permlink, int16(weight)}
	return b.client.broadcast([]HiveOperation{op}, wif)
}

// CustomJson broadcasts a custom_json operation.
// Use reqAuth for active-key actions and reqPostAuth for posting-key actions.
func (b BroadcastAPI) CustomJson(reqAuth, reqPostAuth []string, id, cj, wif string) (string, error) {
	op := CustomJsonOperation{reqAuth, reqPostAuth, id, cj}
	return b.client.broadcast([]HiveOperation{op}, wif)
}

// Transfer sends HIVE or HBD from one account to another.
// amount should be created with ParseAsset, e.g. ParseAsset("1.000 HIVE").
func (b BroadcastAPI) Transfer(from, to string, amount Asset, memo, wif string) (string, error) {
	op := TransferOperation{from, to, amount, memo}
	return b.client.broadcast([]HiveOperation{op}, wif)
}

// ClaimRewards claims pending reward balances (HIVE, HBD, VESTS) into the account.
func (b BroadcastAPI) ClaimRewards(account string, hive, hbd, vests Asset, wif string) (string, error) {
	op := ClaimRewardBalanceOperation{account, hive, hbd, vests}
	return b.client.broadcast([]HiveOperation{op}, wif)
}

func getHiveOpId(op string) uint64 {
	return getHiveOpIds()[op+"_operation"]
}

func getHiveOpIds() map[string]uint64 {
	return map[string]uint64{
		"vote_operation":                           0,
		"comment_operation":                        1,
		"transfer_operation":                       2,
		"transfer_to_vesting_operation":            3,
		"withdraw_vesting_operation":               4,
		"limit_order_create_operation":             5,
		"limit_order_cancel_operation":             6,
		"feed_publish_operation":                   7,
		"convert_operation":                        8,
		"account_create_operation":                 9,
		"account_update_operation":                 10,
		"witness_update_operation":                 11,
		"account_witness_vote_operation":           12,
		"account_witness_proxy_operation":          13,
		"pow_operation":                            14,
		"custom_operation":                         15,
		"report_over_production_operation":         16,
		"delete_comment_operation":                 17,
		"custom_json_operation":                    18,
		"comment_options_operation":                19,
		"set_withdraw_vesting_route_operation":     20,
		"limit_order_create2_operation":            21,
		"claim_account_operation":                  22,
		"create_claimed_account_operation":         23,
		"request_account_recovery_operation":       24,
		"recover_account_operation":                25,
		"change_recovery_account_operation":        26,
		"escrow_transfer_operation":                27,
		"escrow_dispute_operation":                 28,
		"escrow_release_operation":                 29,
		"pow2_operation":                           30,
		"escrow_approve_operation":                 31,
		"transfer_to_savings_operation":            32,
		"transfer_from_savings_operation":          33,
		"cancel_transfer_from_savings_operation":   34,
		"custom_binary_operation":                  35,
		"decline_voting_rights_operation":          36,
		"reset_account_operation":                  37,
		"set_reset_account_operation":              38,
		"claim_reward_balance_operation":           39,
		"delegate_vesting_shares_operation":        40,
		"account_create_with_delegation_operation": 41,
		"witness_set_properties_operation":         42,
		"account_update2_operation":                43,
		"create_proposal_operation":                44,
		"update_proposal_votes_operation":          45,
		"remove_proposal_operation":                46,
		"update_proposal_operation":                47,
		"collateralized_convert_operation":         48,
		"recurrent_transfer_operation":             49,
	}
}
