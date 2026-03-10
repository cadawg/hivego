package hivego

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Asset represents a Hive asset amount such as "1.000 HIVE" or "5.321 HBD".
// Use ParseAsset to create one from a string, or construct directly with Amount, Precision, and Symbol.
type Asset struct {
	Amount    int64
	Precision uint8
	Symbol    string
}

// ParseAsset parses a Hive asset string (e.g. "1.000 HIVE") into an Asset.
func ParseAsset(s string) (Asset, error) {
	parts := strings.SplitN(strings.TrimSpace(s), " ", 2)
	if len(parts) != 2 {
		return Asset{}, errors.New("invalid asset: " + s)
	}
	symbol := parts[1]
	numStr := parts[0]

	var precision uint8
	var intStr string
	dotIdx := strings.Index(numStr, ".")
	if dotIdx == -1 {
		intStr = numStr
	} else {
		precision = uint8(len(numStr) - dotIdx - 1)
		intStr = numStr[:dotIdx] + numStr[dotIdx+1:]
	}

	amount, err := strconv.ParseInt(intStr, 10, 64)
	if err != nil {
		return Asset{}, fmt.Errorf("invalid asset amount: %w", err)
	}

	return Asset{Amount: amount, Precision: precision, Symbol: symbol}, nil
}

// String formats the Asset back to its canonical string form (e.g. "1.000 HIVE").
func (a Asset) String() string {
	if a.Precision == 0 {
		return fmt.Sprintf("%d %s", a.Amount, a.Symbol)
	}
	divisor := int64(1)
	for i := uint8(0); i < a.Precision; i++ {
		divisor *= 10
	}
	whole := a.Amount / divisor
	frac := a.Amount % divisor
	if frac < 0 {
		frac = -frac
	}
	return fmt.Sprintf("%d.%0*d %s", whole, int(a.Precision), frac, a.Symbol)
}

// Block represents a Hive block as returned by block_api.get_block.
// Transactions are returned as raw JSON; use json.Unmarshal to decode individual transactions.
type Block struct {
	BlockID               string            `json:"block_id"`
	Previous              string            `json:"previous"`
	Timestamp             string            `json:"timestamp"`
	Witness               string            `json:"witness"`
	TransactionMerkleRoot string            `json:"transaction_merkle_root"`
	WitnessSignature      string            `json:"witness_signature"`
	Extensions            []interface{}     `json:"extensions"`
	Transactions          []json.RawMessage `json:"transactions"`
	TransactionIDs        []string          `json:"transaction_ids"`
}

// AccountData holds account information as returned by condenser_api.get_accounts.
// Balance fields (e.g. Balance, HbdBalance) are returned as strings like "1.000 HIVE";
// use ParseAsset to convert them if needed.
type AccountData struct {
	ID                     uint32        `json:"id"`
	Name                   string        `json:"name"`
	Owner                  AuthorityData `json:"owner"`
	Active                 AuthorityData `json:"active"`
	Posting                AuthorityData `json:"posting"`
	MemoKey                string        `json:"memo_key"`
	JsonMetadata           string        `json:"json_metadata"`
	PostingJsonMetadata    string        `json:"posting_json_metadata"`
	Proxy                  string        `json:"proxy"`
	Created                string        `json:"created"`
	Mined                  bool          `json:"mined"`
	RecoveryAccount        string        `json:"recovery_account"`
	ResetAccount           string        `json:"reset_account"`
	PostCount              uint32        `json:"post_count"`
	CanVote                bool          `json:"can_vote"`
	Balance                string        `json:"balance"`
	SavingsBalance         string        `json:"savings_balance"`
	HbdBalance             string        `json:"hbd_balance"`
	SavingsHbdBalance      string        `json:"savings_hbd_balance"`
	RewardHbdBalance       string        `json:"reward_hbd_balance"`
	RewardHiveBalance      string        `json:"reward_hive_balance"`
	RewardVestingBalance   string        `json:"reward_vesting_balance"`
	RewardVestingHive      string        `json:"reward_vesting_hive"`
	VestingShares          string        `json:"vesting_shares"`
	DelegatedVestingShares string        `json:"delegated_vesting_shares"`
	ReceivedVestingShares  string        `json:"received_vesting_shares"`
	VestingWithdrawRate    string        `json:"vesting_withdraw_rate"`
	NextVestingWithdrawal  string        `json:"next_vesting_withdrawal"`
	CurationRewards        int64         `json:"curation_rewards"`
	PostingRewards         int64         `json:"posting_rewards"`
	WitnessesVotedFor      uint32        `json:"witnesses_voted_for"`
	LastPost               string        `json:"last_post"`
	LastRootPost           string        `json:"last_root_post"`
	LastVoteTime           string        `json:"last_vote_time"`
	PendingClaimedAccounts uint32        `json:"pending_claimed_accounts"`
}

// AuthorityData holds an authority structure (owner, active, or posting).
type AuthorityData struct {
	WeightThreshold uint32          `json:"weight_threshold"`
	AccountAuths    [][]interface{} `json:"account_auths"`
	KeyAuths        [][]interface{} `json:"key_auths"`
}
