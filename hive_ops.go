package hivego

import (
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// HiveOperation is the interface implemented by all Hive blockchain operations.
// Implement this interface to broadcast operations not built into the library.
//
// SerializeOp returns the binary representation used for local signing.
// OpName returns the Hive operation type name (e.g. "vote", "custom_json").
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
func (b BroadcastAPI) Vote(voter, author, permlink string, weight int, key *KeyPair) (*Transaction, string, error) {
	op := VoteOperation{voter, author, permlink, int16(weight)}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// CustomJson broadcasts a custom_json operation.
// Use reqAuth for active-key actions and reqPostAuth for posting-key actions.
func (b BroadcastAPI) CustomJson(reqAuth, reqPostAuth []string, id, cj string, key *KeyPair) (*Transaction, string, error) {
	op := CustomJsonOperation{reqAuth, reqPostAuth, id, cj}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// Transfer sends HIVE or HBD from one account to another.
// amount should be created with ParseAsset, e.g. ParseAsset("1.000 HIVE").
func (b BroadcastAPI) Transfer(from, to string, amount Asset, memo string, key *KeyPair) (*Transaction, string, error) {
	op := TransferOperation{from, to, amount, memo}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// ClaimRewards claims pending reward balances (HIVE, HBD, VESTS) into the account.
func (b BroadcastAPI) ClaimRewards(account string, hive, hbd, vests Asset, key *KeyPair) (*Transaction, string, error) {
	op := ClaimRewardBalanceOperation{account, hive, hbd, vests}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
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
func (b BroadcastAPI) Comment(parentAuthor, parentPermlink, author, permlink, title, body, jsonMetadata string, key *KeyPair) (*Transaction, string, error) {
	op := CommentOperation{parentAuthor, parentPermlink, author, permlink, title, body, jsonMetadata}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// DeleteComment deletes a post or comment.
func (b BroadcastAPI) DeleteComment(author, permlink string, key *KeyPair) (*Transaction, string, error) {
	op := DeleteCommentOperation{author, permlink}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
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
func (b BroadcastAPI) PowerUp(from, to string, amount Asset, key *KeyPair) (*Transaction, string, error) {
	op := TransferToVestingOperation{from, to, amount}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// PowerDown initiates a 13-week power-down of vestingShares.
// Use ParseAsset("0.000000 VESTS") to cancel.
func (b BroadcastAPI) PowerDown(account string, vestingShares Asset, key *KeyPair) (*Transaction, string, error) {
	op := WithdrawVestingOperation{account, vestingShares}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// Delegate delegates vestingShares of HP from delegator to delegatee.
// Use ParseAsset("0.000000 VESTS") to remove an existing delegation.
func (b BroadcastAPI) Delegate(delegator, delegatee string, vestingShares Asset, key *KeyPair) (*Transaction, string, error) {
	op := DelegateVestingSharesOperation{delegator, delegatee, vestingShares}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
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
func (b BroadcastAPI) VoteWitness(account, witness string, approve bool, key *KeyPair) (*Transaction, string, error) {
	op := AccountWitnessVoteOperation{account, witness, approve}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
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
func (b BroadcastAPI) TransferToSavings(from, to string, amount Asset, memo string, key *KeyPair) (*Transaction, string, error) {
	op := TransferToSavingsOperation{from, to, amount, memo}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// TransferFromSavings initiates a savings withdrawal. requestId must be unique per account.
func (b BroadcastAPI) TransferFromSavings(from string, requestId uint32, to string, amount Asset, memo string, key *KeyPair) (*Transaction, string, error) {
	op := TransferFromSavingsOperation{from, requestId, to, amount, memo}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// --- Vesting route ---

// SetWithdrawVestingRouteOperation routes a percentage of power-down payouts to another account.
// Percent is in basis points (10000 = 100%). Set to 0 to remove a route.
// AutoVest controls whether the routed amount is automatically powered up in the destination account.
type SetWithdrawVestingRouteOperation struct {
	FromAccount string `json:"from_account"`
	ToAccount   string `json:"to_account"`
	Percent     uint16 `json:"percent"`
	AutoVest    bool   `json:"auto_vest"`
}

func (o SetWithdrawVestingRouteOperation) OpName() string { return "set_withdraw_vesting_route" }

// SetWithdrawRoute routes percent of power-down payouts to toAccount.
// percent is in basis points (10000 = 100%). Set to 0 to remove the route.
func (b BroadcastAPI) SetWithdrawRoute(from, to string, percent uint16, autoVest bool, key *KeyPair) (*Transaction, string, error) {
	op := SetWithdrawVestingRouteOperation{from, to, percent, autoVest}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// --- Cancel savings withdrawal ---

// CancelTransferFromSavingsOperation cancels a pending savings withdrawal by request ID.
type CancelTransferFromSavingsOperation struct {
	From      string `json:"from"`
	RequestId uint32 `json:"request_id"`
}

func (o CancelTransferFromSavingsOperation) OpName() string { return "cancel_transfer_from_savings" }

// CancelTransferFromSavings cancels a pending savings withdrawal. requestId must match the original request.
func (b BroadcastAPI) CancelTransferFromSavings(from string, requestId uint32, key *KeyPair) (*Transaction, string, error) {
	op := CancelTransferFromSavingsOperation{from, requestId}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// --- Recurrent transfers ---

// RecurrentTransferOperation schedules a recurring transfer of HIVE or HBD.
// Recurrence is the number of hours between each transfer.
// Executions is the number of times to execute (minimum 2).
// To cancel a recurrent transfer, set Amount to ParseAsset("0.000 HIVE") (or HBD) and Executions to 2.
type RecurrentTransferOperation struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Amount     Asset  `json:"amount"`
	Memo       string `json:"memo"`
	Recurrence uint16 `json:"recurrence"`
	Executions uint16 `json:"executions"`
}

func (o RecurrentTransferOperation) OpName() string { return "recurrent_transfer" }

// RecurrentTransfer schedules a recurring transfer of HIVE or HBD.
// recurrence is the number of hours between each execution; executions is the total number of
// transfers (minimum 2). To cancel, set amount to zero and executions to 2.
func (b BroadcastAPI) RecurrentTransfer(from, to string, amount Asset, memo string, recurrence, executions uint16, key *KeyPair) (*Transaction, string, error) {
	op := RecurrentTransferOperation{from, to, amount, memo, recurrence, executions}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// --- Account update ---

// AccountUpdate2Operation updates account metadata and/or authorities.
// All authority fields and MemoKey are optional — set to nil to leave unchanged.
type AccountUpdate2Operation struct {
	Account             string               `json:"-"`
	Owner               *Authority           `json:"-"`
	Active              *Authority           `json:"-"`
	Posting             *Authority           `json:"-"`
	MemoKey             *secp256k1.PublicKey `json:"-"`
	JsonMetadata        string               `json:"-"`
	PostingJsonMetadata string               `json:"-"`
}

func (o AccountUpdate2Operation) OpName() string { return "account_update2" }

// MarshalJSON encodes AccountUpdate2Operation for the Hive node, rendering MemoKey
// as a public key string and omitting absent optional fields.
func (o AccountUpdate2Operation) MarshalJSON() ([]byte, error) {
	type plain struct {
		Account             string     `json:"account"`
		Owner               *Authority `json:"owner,omitempty"`
		Active              *Authority `json:"active,omitempty"`
		Posting             *Authority `json:"posting,omitempty"`
		MemoKey             string     `json:"memo_key,omitempty"`
		JsonMetadata        string     `json:"json_metadata"`
		PostingJsonMetadata string     `json:"posting_json_metadata"`
	}
	p := plain{
		Account:             o.Account,
		Owner:               o.Owner,
		Active:              o.Active,
		Posting:             o.Posting,
		JsonMetadata:        o.JsonMetadata,
		PostingJsonMetadata: o.PostingJsonMetadata,
	}
	if o.MemoKey != nil {
		p.MemoKey = GetPublicKeyString(o.MemoKey)
	}
	return json.Marshal(p)
}

// UpdateAccount updates the json_metadata and/or posting_json_metadata for an account.
// For authority or memo key changes, construct AccountUpdate2Operation directly and use BroadcastOps.
func (b BroadcastAPI) UpdateAccount(account, jsonMetadata, postingJsonMetadata string, key *KeyPair) (*Transaction, string, error) {
	op := AccountUpdate2Operation{
		Account:             account,
		JsonMetadata:        jsonMetadata,
		PostingJsonMetadata: postingJsonMetadata,
	}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// --- Conversions ---

// ConvertOperation converts HBD to HIVE via the 3.5-day conversion process.
// RequestId must be unique per account.
type ConvertOperation struct {
	Owner     string `json:"owner"`
	RequestId uint32 `json:"request_id"`
	Amount    Asset  `json:"amount"`
}

func (o ConvertOperation) OpName() string { return "convert" }

// CollateralizedConvertOperation converts HIVE to HBD instantly using collateral.
// RequestId must be unique per account.
type CollateralizedConvertOperation struct {
	Owner     string `json:"owner"`
	RequestId uint32 `json:"request_id"`
	Amount    Asset  `json:"amount"`
}

func (o CollateralizedConvertOperation) OpName() string { return "collateralized_convert" }

// Convert converts HBD to HIVE via the 3.5-day conversion process. requestId must be unique per account.
func (b BroadcastAPI) Convert(owner string, requestId uint32, amount Asset, key *KeyPair) (*Transaction, string, error) {
	op := ConvertOperation{owner, requestId, amount}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// CollateralizedConvert converts HIVE to HBD instantly using collateral. requestId must be unique per account.
func (b BroadcastAPI) CollateralizedConvert(owner string, requestId uint32, amount Asset, key *KeyPair) (*Transaction, string, error) {
	op := CollateralizedConvertOperation{owner, requestId, amount}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// --- Limit orders ---

// LimitOrderCreateOperation places a limit order on the internal DEX.
// Expiration must be in "2006-01-02T15:04:05" format.
// Set FillOrKill to true to cancel the order if it cannot be immediately filled.
type LimitOrderCreateOperation struct {
	Owner        string `json:"owner"`
	OrderId      uint32 `json:"orderid"`
	AmountToSell Asset  `json:"amount_to_sell"`
	MinToReceive Asset  `json:"min_to_receive"`
	FillOrKill   bool   `json:"fill_or_kill"`
	Expiration   string `json:"expiration"`
}

func (o LimitOrderCreateOperation) OpName() string { return "limit_order_create" }

// LimitOrderCancelOperation cancels an open limit order by ID.
type LimitOrderCancelOperation struct {
	Owner   string `json:"owner"`
	OrderId uint32 `json:"orderid"`
}

func (o LimitOrderCancelOperation) OpName() string { return "limit_order_cancel" }

// LimitOrderCreate places a limit order on the internal DEX.
// expiration must be in "2006-01-02T15:04:05" format.
func (b BroadcastAPI) LimitOrderCreate(owner string, orderId uint32, amountToSell, minToReceive Asset, fillOrKill bool, expiration string, key *KeyPair) (*Transaction, string, error) {
	op := LimitOrderCreateOperation{owner, orderId, amountToSell, minToReceive, fillOrKill, expiration}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// LimitOrderCancel cancels an open limit order by ID.
func (b BroadcastAPI) LimitOrderCancel(owner string, orderId uint32, key *KeyPair) (*Transaction, string, error) {
	op := LimitOrderCancelOperation{owner, orderId}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// --- Witness proxy ---

// AccountWitnessProxyOperation delegates witness votes to a proxy account.
// Set Proxy to "" to remove an existing proxy.
type AccountWitnessProxyOperation struct {
	Account string `json:"account"`
	Proxy   string `json:"proxy"`
}

func (o AccountWitnessProxyOperation) OpName() string { return "account_witness_proxy" }

// SetWitnessProxy delegates witness votes to proxy. Set proxy to "" to remove.
func (b BroadcastAPI) SetWitnessProxy(account, proxy string, key *KeyPair) (*Transaction, string, error) {
	op := AccountWitnessProxyOperation{account, proxy}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// --- DHF proposals ---

// CreateProposalOperation submits a funding proposal to the Decentralised Hive Fund.
// StartDate and EndDate must be in "2006-01-02T15:04:05" format. DailyPay must be HBD.
type CreateProposalOperation struct {
	Creator   string `json:"creator"`
	Receiver  string `json:"receiver"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	DailyPay  Asset  `json:"daily_pay"`
	Subject   string `json:"subject"`
	Permlink  string `json:"permlink"`
}

func (o CreateProposalOperation) OpName() string { return "create_proposal" }

// UpdateProposalVotesOperation approves or removes votes for one or more proposals.
type UpdateProposalVotesOperation struct {
	Voter       string  `json:"voter"`
	ProposalIds []int64 `json:"proposal_ids"`
	Approve     bool    `json:"approve"`
}

func (o UpdateProposalVotesOperation) OpName() string { return "update_proposal_votes" }

// RemoveProposalOperation removes one or more proposals. Only the proposal creator can remove.
type RemoveProposalOperation struct {
	ProposalOwner string  `json:"proposal_owner"`
	ProposalIds   []int64 `json:"proposal_ids"`
}

func (o RemoveProposalOperation) OpName() string { return "remove_proposal" }

// UpdateProposalOperation updates an existing proposal's pay, subject, permlink, or end date.
// DailyPay can only be lowered, not increased — the chain rejects any value above the current pay.
// EndDate is optional — set to "" to leave unchanged. Must be in "2006-01-02T15:04:05" format if provided.
type UpdateProposalOperation struct {
	ProposalId int64  `json:"proposal_id"`
	Creator    string `json:"creator"`
	DailyPay   Asset  `json:"daily_pay"`
	Subject    string `json:"subject"`
	Permlink   string `json:"permlink"`
	EndDate    string `json:"end_date,omitempty"`
}

func (o UpdateProposalOperation) OpName() string { return "update_proposal" }

// CreateProposal submits a DHF funding proposal. startDate/endDate in "2006-01-02T15:04:05" format, dailyPay must be HBD.
func (b BroadcastAPI) CreateProposal(creator, receiver, startDate, endDate string, dailyPay Asset, subject, permlink string, key *KeyPair) (*Transaction, string, error) {
	op := CreateProposalOperation{creator, receiver, startDate, endDate, dailyPay, subject, permlink}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// UpdateProposalVotes approves or removes votes for the given proposal IDs.
func (b BroadcastAPI) UpdateProposalVotes(voter string, proposalIds []int64, approve bool, key *KeyPair) (*Transaction, string, error) {
	op := UpdateProposalVotesOperation{voter, proposalIds, approve}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// RemoveProposal removes proposals by ID. Only the proposal creator can remove.
func (b BroadcastAPI) RemoveProposal(owner string, proposalIds []int64, key *KeyPair) (*Transaction, string, error) {
	op := RemoveProposalOperation{owner, proposalIds}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// UpdateProposal updates an existing proposal. endDate is optional — pass "" to leave unchanged.
func (b BroadcastAPI) UpdateProposal(proposalId int64, creator string, dailyPay Asset, subject, permlink, endDate string, key *KeyPair) (*Transaction, string, error) {
	op := UpdateProposalOperation{proposalId, creator, dailyPay, subject, permlink, endDate}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// --- Account creation ---

// ClaimAccountOperation claims an account creation token using RC or a HIVE fee.
// Set Fee to ParseAsset("0.000 HIVE") to use RC instead of a HIVE fee.
type ClaimAccountOperation struct {
	Creator string `json:"creator"`
	Fee     Asset  `json:"fee"`
}

func (o ClaimAccountOperation) OpName() string { return "claim_account" }

// ClaimAccount claims an account creation token. Use ParseAsset("0.000 HIVE") as fee to claim via RC.
func (b BroadcastAPI) ClaimAccount(creator string, fee Asset, key *KeyPair) (*Transaction, string, error) {
	op := ClaimAccountOperation{creator, fee}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// --- Price feed ---

// FeedPublishOperation publishes a HIVE/HBD price feed for consensus. Witnesses only.
// ExchangeRate.Base should be HBD and ExchangeRate.Quote should be HIVE, representing
// the price of 1 HBD in HIVE (e.g. base="1.000 HBD", quote="3.500 HIVE" means 1 HBD = 3.5 HIVE).
type FeedPublishOperation struct {
	Publisher    string `json:"publisher"`
	ExchangeRate Price  `json:"exchange_rate"`
}

func (o FeedPublishOperation) OpName() string { return "feed_publish" }

// FeedPublish publishes a HIVE/HBD price feed. base should be HBD, quote should be HIVE.
// Only witnesses with an active signing key should call this.
func (b BroadcastAPI) FeedPublish(publisher string, base, quote Asset, key *KeyPair) (*Transaction, string, error) {
	op := FeedPublishOperation{publisher, Price{base, quote}}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// --- Account creation ---

// AccountCreateOperation creates a new account with a HIVE fee.
// All authority fields and MemoKey are required.
type AccountCreateOperation struct {
	Fee            Asset                `json:"-"`
	Creator        string               `json:"-"`
	NewAccountName string               `json:"-"`
	Owner          Authority            `json:"-"`
	Active         Authority            `json:"-"`
	Posting        Authority            `json:"-"`
	MemoKey        *secp256k1.PublicKey `json:"-"`
	JsonMetadata   string               `json:"-"`
}

func (o AccountCreateOperation) OpName() string { return "account_create" }

func (o AccountCreateOperation) MarshalJSON() ([]byte, error) {
	memoKey := ""
	if o.MemoKey != nil {
		memoKey = GetPublicKeyString(o.MemoKey)
	}
	return json.Marshal(struct {
		Fee            Asset     `json:"fee"`
		Creator        string    `json:"creator"`
		NewAccountName string    `json:"new_account_name"`
		Owner          Authority `json:"owner"`
		Active         Authority `json:"active"`
		Posting        Authority `json:"posting"`
		MemoKey        string    `json:"memo_key"`
		JsonMetadata   string    `json:"json_metadata"`
	}{o.Fee, o.Creator, o.NewAccountName, o.Owner, o.Active, o.Posting, memoKey, o.JsonMetadata})
}

// CreateClaimedAccountOperation creates a new account from a previously claimed token (no fee).
// All authority fields and MemoKey are required.
type CreateClaimedAccountOperation struct {
	Creator        string               `json:"-"`
	NewAccountName string               `json:"-"`
	Owner          Authority            `json:"-"`
	Active         Authority            `json:"-"`
	Posting        Authority            `json:"-"`
	MemoKey        *secp256k1.PublicKey `json:"-"`
	JsonMetadata   string               `json:"-"`
}

func (o CreateClaimedAccountOperation) OpName() string { return "create_claimed_account" }

func (o CreateClaimedAccountOperation) MarshalJSON() ([]byte, error) {
	memoKey := ""
	if o.MemoKey != nil {
		memoKey = GetPublicKeyString(o.MemoKey)
	}
	return json.Marshal(struct {
		Creator        string    `json:"creator"`
		NewAccountName string    `json:"new_account_name"`
		Owner          Authority `json:"owner"`
		Active         Authority `json:"active"`
		Posting        Authority `json:"posting"`
		MemoKey        string    `json:"memo_key"`
		JsonMetadata   string    `json:"json_metadata"`
	}{o.Creator, o.NewAccountName, o.Owner, o.Active, o.Posting, memoKey, o.JsonMetadata})
}

// CreateAccount creates a new account with a HIVE fee.
func (b BroadcastAPI) CreateAccount(fee Asset, creator, newAccountName string, owner, active, posting Authority, memoKey *secp256k1.PublicKey, jsonMetadata string, key *KeyPair) (*Transaction, string, error) {
	op := AccountCreateOperation{fee, creator, newAccountName, owner, active, posting, memoKey, jsonMetadata}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// CreateClaimedAccount creates a new account from a previously claimed token (no fee).
func (b BroadcastAPI) CreateClaimedAccount(creator, newAccountName string, owner, active, posting Authority, memoKey *secp256k1.PublicKey, jsonMetadata string, key *KeyPair) (*Transaction, string, error) {
	op := CreateClaimedAccountOperation{creator, newAccountName, owner, active, posting, memoKey, jsonMetadata}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// --- Account update (legacy) ---

// AccountUpdateOperation updates account authorities and memo key (legacy account_update operation).
// All authority fields are optional — set to nil to leave unchanged. MemoKey is required.
// For metadata-only updates (json_metadata, posting_json_metadata), prefer [AccountUpdate2Operation].
type AccountUpdateOperation struct {
	Account      string               `json:"-"`
	Owner        *Authority           `json:"-"`
	Active       *Authority           `json:"-"`
	Posting      *Authority           `json:"-"`
	MemoKey      *secp256k1.PublicKey `json:"-"` // required
	JsonMetadata string               `json:"-"`
}

func (o AccountUpdateOperation) OpName() string { return "account_update" }

func (o AccountUpdateOperation) MarshalJSON() ([]byte, error) {
	memoKey := ""
	if o.MemoKey != nil {
		memoKey = GetPublicKeyString(o.MemoKey)
	}
	type plain struct {
		Account      string     `json:"account"`
		Owner        *Authority `json:"owner,omitempty"`
		Active       *Authority `json:"active,omitempty"`
		Posting      *Authority `json:"posting,omitempty"`
		MemoKey      string     `json:"memo_key"`
		JsonMetadata string     `json:"json_metadata"`
	}
	return json.Marshal(plain{o.Account, o.Owner, o.Active, o.Posting, memoKey, o.JsonMetadata})
}

// --- Witness operations ---

// WitnessUpdateOperation updates a witness's URL, signing key, chain properties, and fee.
// Set BlockSigningKey to nil to write 33 zero bytes, which disables the witness.
type WitnessUpdateOperation struct {
	Owner           string               `json:"-"`
	Url             string               `json:"-"`
	BlockSigningKey *secp256k1.PublicKey `json:"-"`
	Props           ChainProperties      `json:"-"`
	Fee             Asset                `json:"-"`
}

func (o WitnessUpdateOperation) OpName() string { return "witness_update" }

func (o WitnessUpdateOperation) MarshalJSON() ([]byte, error) {
	signingKey := ""
	if o.BlockSigningKey != nil {
		signingKey = GetPublicKeyString(o.BlockSigningKey)
	}
	return json.Marshal(struct {
		Owner           string          `json:"owner"`
		Url             string          `json:"url"`
		BlockSigningKey string          `json:"block_signing_key"`
		Props           ChainProperties `json:"props"`
		Fee             Asset           `json:"fee"`
	}{o.Owner, o.Url, signingKey, o.Props, o.Fee})
}

// WitnessSetPropertiesOperation updates witness properties using the modern key-value format.
// Props values are raw binary — use encoding/binary to build property values.
// Keys must be one of: "key", "new_signing_key", "account_creation_fee", "maximum_block_size",
// "hbd_interest_rate", "hbd_exchange_rate", "url", "account_subsidy_budget", "account_subsidy_decay".
type WitnessSetPropertiesOperation struct {
	Owner string            `json:"owner"`
	Props map[string][]byte `json:"-"` // serialized as [["key", "hexvalue"], ...]
}

func (o WitnessSetPropertiesOperation) OpName() string { return "witness_set_properties" }

func (o WitnessSetPropertiesOperation) MarshalJSON() ([]byte, error) {
	keys := make([]string, 0, len(o.Props))
	for k := range o.Props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([][2]string, len(keys))
	for i, k := range keys {
		pairs[i] = [2]string{k, hex.EncodeToString(o.Props[k])}
	}
	return json.Marshal(struct {
		Owner      string      `json:"owner"`
		Props      [][2]string `json:"props"`
		Extensions []string    `json:"extensions"`
	}{o.Owner, pairs, []string{}})
}

// UpdateWitness updates a witness registration (witness_update).
func (b BroadcastAPI) UpdateWitness(owner, url string, signingKey *secp256k1.PublicKey, props ChainProperties, fee Asset, key *KeyPair) (*Transaction, string, error) {
	op := WitnessUpdateOperation{owner, url, signingKey, props, fee}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// WitnessSetProperties updates witness properties using the modern key-value format.
func (b BroadcastAPI) WitnessSetProperties(owner string, props map[string][]byte, key *KeyPair) (*Transaction, string, error) {
	op := WitnessSetPropertiesOperation{Owner: owner, Props: props}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// --- Account recovery ---

// RequestAccountRecoveryOperation initiates account recovery. Submitted by the recovery account.
type RequestAccountRecoveryOperation struct {
	RecoveryAccount   string    `json:"recovery_account"`
	AccountToRecover  string    `json:"account_to_recover"`
	NewOwnerAuthority Authority `json:"new_owner_authority"`
}

func (o RequestAccountRecoveryOperation) OpName() string { return "request_account_recovery" }

// RecoverAccountOperation completes account recovery using both the new and a recent old owner authority.
type RecoverAccountOperation struct {
	AccountToRecover     string    `json:"account_to_recover"`
	NewOwnerAuthority    Authority `json:"new_owner_authority"`
	RecentOwnerAuthority Authority `json:"recent_owner_authority"`
}

func (o RecoverAccountOperation) OpName() string { return "recover_account" }

// ChangeRecoveryAccountOperation changes the designated recovery account. Takes 30 days to take effect.
type ChangeRecoveryAccountOperation struct {
	AccountToRecover   string `json:"account_to_recover"`
	NewRecoveryAccount string `json:"new_recovery_account"`
}

func (o ChangeRecoveryAccountOperation) OpName() string { return "change_recovery_account" }

// CancelDeclineVotingRights cancels a pending (not yet effective) decline-voting-rights request.
// To initiate a decline, construct DeclineVotingRightsOperation directly with
// Decline: true and IUnderstandThisIsIrreversible: true, then use BroadcastOps.
func (b BroadcastAPI) CancelDeclineVotingRights(account string, key *KeyPair) (*Transaction, string, error) {
	op := DeclineVotingRightsOperation{Account: account, Decline: false}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// RequestAccountRecovery initiates recovery. Must be submitted by the recovery account.
func (b BroadcastAPI) RequestAccountRecovery(recoveryAccount, accountToRecover string, newOwnerAuthority Authority, key *KeyPair) (*Transaction, string, error) {
	op := RequestAccountRecoveryOperation{recoveryAccount, accountToRecover, newOwnerAuthority}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// RecoverAccount completes account recovery. Must be signed with both the new and a recent old owner key.
func (b BroadcastAPI) RecoverAccount(accountToRecover string, newOwnerAuthority, recentOwnerAuthority Authority, key *KeyPair) (*Transaction, string, error) {
	op := RecoverAccountOperation{accountToRecover, newOwnerAuthority, recentOwnerAuthority}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// ChangeRecoveryAccount changes the recovery account. Takes 30 days to take effect.
func (b BroadcastAPI) ChangeRecoveryAccount(accountToRecover, newRecoveryAccount string, key *KeyPair) (*Transaction, string, error) {
	op := ChangeRecoveryAccountOperation{accountToRecover, newRecoveryAccount}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// --- Escrow ---

// EscrowTransferOperation locks funds with a third-party agent until conditions are met.
// RatificationDeadline and EscrowExpiration must be in "2006-01-02T15:04:05" format.
type EscrowTransferOperation struct {
	From                 string `json:"from"`
	To                   string `json:"to"`
	Agent                string `json:"agent"`
	EscrowId             uint32 `json:"escrow_id"`
	HbdAmount            Asset  `json:"hbd_amount"`
	HiveAmount           Asset  `json:"hive_amount"`
	Fee                  Asset  `json:"fee"`
	RatificationDeadline string `json:"ratification_deadline"`
	EscrowExpiration     string `json:"escrow_expiration"`
	JsonMeta             string `json:"json_meta"`
}

func (o EscrowTransferOperation) OpName() string { return "escrow_transfer" }

// EscrowApproveOperation approves or rejects an escrow by the to-party or agent.
type EscrowApproveOperation struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Agent    string `json:"agent"`
	Who      string `json:"who"`
	EscrowId uint32 `json:"escrow_id"`
	Approve  bool   `json:"approve"`
}

func (o EscrowApproveOperation) OpName() string { return "escrow_approve" }

// EscrowDisputeOperation raises a dispute, transferring release authority to the agent.
type EscrowDisputeOperation struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Agent    string `json:"agent"`
	Who      string `json:"who"`
	EscrowId uint32 `json:"escrow_id"`
}

func (o EscrowDisputeOperation) OpName() string { return "escrow_dispute" }

// EscrowReleaseOperation releases escrowed funds to the receiver.
type EscrowReleaseOperation struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Agent      string `json:"agent"`
	Who        string `json:"who"`
	Receiver   string `json:"receiver"`
	EscrowId   uint32 `json:"escrow_id"`
	HbdAmount  Asset  `json:"hbd_amount"`
	HiveAmount Asset  `json:"hive_amount"`
}

func (o EscrowReleaseOperation) OpName() string { return "escrow_release" }

// EscrowTransfer locks funds in escrow with an agent. ratificationDeadline and escrowExpiration in "2006-01-02T15:04:05" format.
func (b BroadcastAPI) EscrowTransfer(from, to, agent string, escrowId uint32, hbdAmount, hiveAmount, fee Asset, ratificationDeadline, escrowExpiration, jsonMeta string, key *KeyPair) (*Transaction, string, error) {
	op := EscrowTransferOperation{from, to, agent, escrowId, hbdAmount, hiveAmount, fee, ratificationDeadline, escrowExpiration, jsonMeta}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// EscrowApprove approves or rejects an escrow. who must be the to-party or agent.
func (b BroadcastAPI) EscrowApprove(from, to, agent, who string, escrowId uint32, approve bool, key *KeyPair) (*Transaction, string, error) {
	op := EscrowApproveOperation{from, to, agent, who, escrowId, approve}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// EscrowDispute raises a dispute on an escrow, handing release authority to the agent.
func (b BroadcastAPI) EscrowDispute(from, to, agent, who string, escrowId uint32, key *KeyPair) (*Transaction, string, error) {
	op := EscrowDisputeOperation{from, to, agent, who, escrowId}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// EscrowRelease releases escrowed funds to receiver.
func (b BroadcastAPI) EscrowRelease(from, to, agent, who, receiver string, escrowId uint32, hbdAmount, hiveAmount Asset, key *KeyPair) (*Transaction, string, error) {
	op := EscrowReleaseOperation{from, to, agent, who, receiver, escrowId, hbdAmount, hiveAmount}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// --- Custom ---

// CustomOperation broadcasts a custom operation with an integer id and arbitrary string data.
// RequiredAuths must contain accounts with active key authority.
type CustomOperation struct {
	RequiredAuths []string `json:"required_auths"`
	Id            uint16   `json:"id"`
	Data          string   `json:"data"`
}

func (o CustomOperation) OpName() string { return "custom" }

// Custom broadcasts a custom operation.
func (b BroadcastAPI) Custom(requiredAuths []string, id uint16, data string, key *KeyPair) (*Transaction, string, error) {
	op := CustomOperation{requiredAuths, id, data}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// --- Custom binary ---

// CustomBinaryOperation broadcasts arbitrary binary data with optional authority requirements.
// Id must be less than 32 characters. Data is arbitrary binary.
//
// Note: beem's implementation of this operation is incomplete (it uses uint16 for id and omits
// all auth fields). The wire format here follows the Hive C++ protocol spec.
type CustomBinaryOperation struct {
	RequiredOwnerAuths   []string    `json:"-"`
	RequiredActiveAuths  []string    `json:"-"`
	RequiredPostingAuths []string    `json:"-"`
	RequiredAuths        []Authority `json:"-"`
	Id                   string      `json:"-"`
	Data                 []byte      `json:"-"`
}

func (o CustomBinaryOperation) OpName() string { return "custom_binary" }

func (o CustomBinaryOperation) MarshalJSON() ([]byte, error) {
	ownerAuths := o.RequiredOwnerAuths
	if ownerAuths == nil {
		ownerAuths = []string{}
	}
	activeAuths := o.RequiredActiveAuths
	if activeAuths == nil {
		activeAuths = []string{}
	}
	postingAuths := o.RequiredPostingAuths
	if postingAuths == nil {
		postingAuths = []string{}
	}
	reqAuths := o.RequiredAuths
	if reqAuths == nil {
		reqAuths = []Authority{}
	}
	return json.Marshal(struct {
		RequiredOwnerAuths   []string    `json:"required_owner_auths"`
		RequiredActiveAuths  []string    `json:"required_active_auths"`
		RequiredPostingAuths []string    `json:"required_posting_auths"`
		RequiredAuths        []Authority `json:"required_auths"`
		Id                   string      `json:"id"`
		Data                 string      `json:"data"` // hex-encoded
	}{ownerAuths, activeAuths, postingAuths, reqAuths, o.Id, hex.EncodeToString(o.Data)})
}

// CustomBinary broadcasts arbitrary binary data.
func (b BroadcastAPI) CustomBinary(requiredOwnerAuths, requiredActiveAuths, requiredPostingAuths []string, requiredAuths []Authority, id string, data []byte, key *KeyPair) (*Transaction, string, error) {
	op := CustomBinaryOperation{requiredOwnerAuths, requiredActiveAuths, requiredPostingAuths, requiredAuths, id, data}
	return b.client.BroadcastOps([]HiveOperation{op}, key)
}

// --- Decline voting rights ---

// DeclineVotingRightsOperation permanently removes an account's ability to vote.
// THIS CANNOT BE UNDONE. You must set IUnderstandThisIsIrreversible to true when Decline is true.
// Set Decline to false to cancel a pending (not yet effective) decline request.
type DeclineVotingRightsOperation struct {
	Account string `json:"account"`
	Decline bool   `json:"decline"`
	// IUnderstandThisIsIrreversible must be true when Decline is true.
	// Declining voting rights permanently removes the account's ability to vote and cannot be undone.
	IUnderstandThisIsIrreversible bool `json:"-"`
}

func (o DeclineVotingRightsOperation) OpName() string { return "decline_voting_rights" }

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
