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

// --- Comment operations ---

// CommentOperation creates or edits a post or comment.
// For a top-level post, set ParentAuthor to "" and ParentPermlink to the category tag.
// For a reply, set ParentAuthor and ParentPermlink to the parent post's values.
type CommentOperation struct {
	ParentAuthor   string `json:"parent_author"`
	ParentPermlink string `json:"parent_permlink"`
	Author         string `json:"author"`
	Permlink       string `json:"permlink"`
	Title          string `json:"title"`
	Body           string `json:"body"`
	JsonMetadata   string `json:"json_metadata"`
}

func (o CommentOperation) OpName() string { return "comment" }

// Beneficiary is a share of a post's author reward routed to another account.
// Weight is in basis points out of 10000 (e.g. 500 = 5%).
// The sum of all beneficiary weights must not exceed 10000.
type Beneficiary struct {
	Account string `json:"account"`
	Weight  uint16 `json:"weight"`
}

// CommentOptionsOperation sets payout and beneficiary options on a post.
// It must be submitted in the same transaction as the CommentOperation it targets.
// Use client.BuildTransaction to combine them.
//
// MaxAcceptedPayout: ParseAsset("1000000.000 HBD") for no limit, or ParseAsset("0.000 HBD") to decline.
// PercentHbd: portion of author reward paid as HBD, in basis points (10000 = 100%).
type CommentOptionsOperation struct {
	Author               string        `json:"author"`
	Permlink             string        `json:"permlink"`
	MaxAcceptedPayout    Asset         `json:"max_accepted_payout"`
	PercentHbd           uint16        `json:"percent_hbd"`
	AllowVotes           bool          `json:"allow_votes"`
	AllowCurationRewards bool          `json:"allow_curation_rewards"`
	Beneficiaries        []Beneficiary `json:"beneficiaries"`
}

func (o CommentOptionsOperation) OpName() string { return "comment_options" }

// DeleteCommentOperation deletes a post or comment.
// The post must have no replies, no pending payout, and no net votes.
type DeleteCommentOperation struct {
	Author   string `json:"author"`
	Permlink string `json:"permlink"`
}

func (o DeleteCommentOperation) OpName() string { return "delete_comment" }

// Comment publishes a post or reply.
// For a top-level post, set parentAuthor to "" and parentPermlink to the category tag.
func (b BroadcastAPI) Comment(parentAuthor, parentPermlink, author, permlink, title, body, jsonMetadata, wif string) (string, error) {
	op := CommentOperation{parentAuthor, parentPermlink, author, permlink, title, body, jsonMetadata}
	return b.client.broadcast([]HiveOperation{op}, wif)
}

// DeleteComment deletes a post or comment.
func (b BroadcastAPI) DeleteComment(author, permlink, wif string) (string, error) {
	op := DeleteCommentOperation{author, permlink}
	return b.client.broadcast([]HiveOperation{op}, wif)
}

// --- Vesting (HP) operations ---

// TransferToVestingOperation converts HIVE to Hive Power (HP).
// Set To to "" to power up to the same account as From.
type TransferToVestingOperation struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount Asset  `json:"amount"`
}

func (o TransferToVestingOperation) OpName() string { return "transfer_to_vesting" }

// WithdrawVestingOperation initiates a power-down, converting HP back to HIVE over 13 weeks.
// Set VestingShares to ParseAsset("0.000000 VESTS") to cancel an in-progress power-down.
type WithdrawVestingOperation struct {
	Account       string `json:"account"`
	VestingShares Asset  `json:"vesting_shares"`
}

func (o WithdrawVestingOperation) OpName() string { return "withdraw_vesting" }

// DelegateVestingSharesOperation delegates HP to another account.
// Set VestingShares to ParseAsset("0.000000 VESTS") to remove an existing delegation.
type DelegateVestingSharesOperation struct {
	Delegator     string `json:"delegator"`
	Delegatee     string `json:"delegatee"`
	VestingShares Asset  `json:"vesting_shares"`
}

func (o DelegateVestingSharesOperation) OpName() string { return "delegate_vesting_shares" }

// PowerUp converts HIVE to Hive Power. Set to "" to power up to the same account as from.
func (b BroadcastAPI) PowerUp(from, to string, amount Asset, wif string) (string, error) {
	op := TransferToVestingOperation{from, to, amount}
	return b.client.broadcast([]HiveOperation{op}, wif)
}

// PowerDown initiates a 13-week power-down of vestingShares.
// Use ParseAsset("0.000000 VESTS") to cancel.
func (b BroadcastAPI) PowerDown(account string, vestingShares Asset, wif string) (string, error) {
	op := WithdrawVestingOperation{account, vestingShares}
	return b.client.broadcast([]HiveOperation{op}, wif)
}

// Delegate delegates vestingShares of HP from delegator to delegatee.
// Use ParseAsset("0.000000 VESTS") to remove an existing delegation.
func (b BroadcastAPI) Delegate(delegator, delegatee string, vestingShares Asset, wif string) (string, error) {
	op := DelegateVestingSharesOperation{delegator, delegatee, vestingShares}
	return b.client.broadcast([]HiveOperation{op}, wif)
}

// --- Witness operations ---

// AccountWitnessVoteOperation votes for or removes a vote from a witness.
type AccountWitnessVoteOperation struct {
	Account string `json:"account"`
	Witness string `json:"witness"`
	Approve bool   `json:"approve"`
}

func (o AccountWitnessVoteOperation) OpName() string { return "account_witness_vote" }

// VoteWitness approves or removes a vote for a witness.
func (b BroadcastAPI) VoteWitness(account, witness string, approve bool, wif string) (string, error) {
	op := AccountWitnessVoteOperation{account, witness, approve}
	return b.client.broadcast([]HiveOperation{op}, wif)
}

// --- Savings operations ---

// TransferToSavingsOperation moves HIVE or HBD into the savings balance.
// Savings have a 3-day withdrawal delay.
type TransferToSavingsOperation struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount Asset  `json:"amount"`
	Memo   string `json:"memo"`
}

func (o TransferToSavingsOperation) OpName() string { return "transfer_to_savings" }

// TransferFromSavingsOperation initiates a withdrawal from savings.
// RequestId must be unique per account — use an incrementing counter.
type TransferFromSavingsOperation struct {
	From      string `json:"from"`
	RequestId uint32 `json:"request_id"`
	To        string `json:"to"`
	Amount    Asset  `json:"amount"`
	Memo      string `json:"memo"`
}

func (o TransferFromSavingsOperation) OpName() string { return "transfer_from_savings" }

// TransferToSavings moves HIVE or HBD into the savings balance (3-day withdrawal delay).
func (b BroadcastAPI) TransferToSavings(from, to string, amount Asset, memo, wif string) (string, error) {
	op := TransferToSavingsOperation{from, to, amount, memo}
	return b.client.broadcast([]HiveOperation{op}, wif)
}

// TransferFromSavings initiates a savings withdrawal. requestId must be unique per account.
func (b BroadcastAPI) TransferFromSavings(from string, requestId uint32, to string, amount Asset, memo, wif string) (string, error) {
	op := TransferFromSavingsOperation{from, requestId, to, amount, memo}
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
