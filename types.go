package hivego

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// HiveTime is a [time.Time] that marshals/unmarshals Hive's timestamp format ("2006-01-02T15:04:05").
// Hive timestamps carry no timezone suffix and are always UTC, making them incompatible with
// Go's default RFC3339 JSON handling. Use [HiveTime.Time] to get the underlying [time.Time].
type HiveTime time.Time

const hiveTimeLayout = "2006-01-02T15:04:05"

func (ht *HiveTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	t, err := time.Parse(hiveTimeLayout, s)
	if err != nil {
		return err
	}
	*ht = HiveTime(t)
	return nil
}

func (ht HiveTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(ht).Format(hiveTimeLayout) + `"`), nil
}

// Time returns the underlying time.Time value.
func (ht HiveTime) Time() time.Time { return time.Time(ht) }

// RC holds a resource-credit manabar as returned by the API (e.g. VotingManabar, DownvoteManabar).
// CurrentMana is in raw units (not percentage); divide by the account's VestingShares to get
// the effective percentage. LastUpdateTime is a Unix timestamp.
type RC struct {
	CurrentMana    int64 `json:"current_mana"`
	LastUpdateTime int64 `json:"last_update_time"`
}

// Asset represents a Hive asset amount such as "1.000 HIVE" or "5.321 HBD".
// Use [ParseAsset] to create one from a string. The binary serializer requires all three
// fields, so constructing an Asset directly is only needed for zero values or test fixtures.
//
// Precisions: HIVE and HBD use 3 decimal places; VESTS uses 6.
type Asset struct {
	Amount    int64
	Precision uint8
	Symbol    string
}

// ParseAsset parses a Hive asset string (e.g. "1.000 HIVE") into an Asset.
func ParseAsset(s string) (Asset, error) {
	parts := strings.SplitN(strings.TrimSpace(s), " ", 2)
	if len(parts) != 2 {
		return Asset{}, fmt.Errorf("%q: %w", s, ErrInvalidAsset)
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

// MarshalJSON encodes Asset as Hive's canonical string form (e.g. "1.000 HIVE").
func (a Asset) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

// UnmarshalJSON decodes Asset from either canonical string form or legacy object form.
func (a *Asset) UnmarshalJSON(b []byte) error {
	var asString string
	if err := json.Unmarshal(b, &asString); err == nil {
		parsed, err := ParseAsset(asString)
		if err != nil {
			return err
		}
		*a = parsed
		return nil
	}

	var legacy struct {
		Amount    int64  `json:"amount"`
		Precision uint8  `json:"precision"`
		Symbol    string `json:"symbol"`
	}
	if err := json.Unmarshal(b, &legacy); err != nil {
		return err
	}
	if strings.TrimSpace(legacy.Symbol) == "" {
		return fmt.Errorf("asset symbol required: %w", ErrInvalidAsset)
	}

	*a = Asset{Amount: legacy.Amount, Precision: legacy.Precision, Symbol: legacy.Symbol}
	return nil
}

// Block represents a Hive block as returned by block_api.get_block.
// Transactions are raw JSON — unmarshal them individually based on the operation types you care about.
// TransactionIDs is populated when the block is fetched with transaction IDs included.
type Block struct {
	BlockID               string            `json:"block_id"`
	Previous              string            `json:"previous"`
	Timestamp             HiveTime          `json:"timestamp"`
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
// Note: NextVestingWithdrawal is "1969-12-31T23:59:59" when no power-down is active.
type AccountData struct {
	ID                            uint32        `json:"id"`
	Name                          string        `json:"name"`
	Owner                         AuthorityData `json:"owner"`
	Active                        AuthorityData `json:"active"`
	Posting                       AuthorityData `json:"posting"`
	MemoKey                       string        `json:"memo_key"`
	JsonMetadata                  string        `json:"json_metadata"`
	PostingJsonMetadata           string        `json:"posting_json_metadata"`
	Proxy                         string        `json:"proxy"`
	LastOwnerUpdate               HiveTime      `json:"last_owner_update"`
	LastAccountUpdate             HiveTime      `json:"last_account_update"`
	Created                       HiveTime      `json:"created"`
	Mined                         bool          `json:"mined"`
	RecoveryAccount               string        `json:"recovery_account"`
	ResetAccount                  string        `json:"reset_account"`
	LastAccountRecovery           HiveTime      `json:"last_account_recovery"`
	PostCount                     uint32        `json:"post_count"`
	CanVote                       bool          `json:"can_vote"`
	VotingPower                   int16         `json:"voting_power"`
	LastVoteTime                  HiveTime      `json:"last_vote_time"`
	Balance                       string        `json:"balance"`
	SavingsBalance                string        `json:"savings_balance"`
	HbdBalance                    string        `json:"hbd_balance"`
	HbdSeconds                    string        `json:"hbd_seconds"`
	HbdSecondsLastUpdate          HiveTime      `json:"hbd_seconds_last_update"`
	HbdLastInterestPayment        HiveTime      `json:"hbd_last_interest_payment"`
	SavingsHbdBalance             string        `json:"savings_hbd_balance"`
	SavingsHbdSeconds             string        `json:"savings_hbd_seconds"`
	SavingsHbdLastUpdate          HiveTime      `json:"savings_hbd_last_update"`
	SavingsHbdLastInterestPayment HiveTime      `json:"savings_hbd_last_interest_payment"`
	SavingsWithdrawRequests       uint32        `json:"savings_withdraw_requests"`
	RewardHbdBalance              string        `json:"reward_hbd_balance"`
	RewardHiveBalance             string        `json:"reward_hive_balance"`
	RewardVestingBalance          string        `json:"reward_vesting_balance"`
	RewardVestingHive             string        `json:"reward_vesting_hive"`
	VestingShares                 string        `json:"vesting_shares"`
	DelegatedVestingShares        string        `json:"delegated_vesting_shares"`
	ReceivedVestingShares         string        `json:"received_vesting_shares"`
	VestingWithdrawRate           string        `json:"vesting_withdraw_rate"`
	NextVestingWithdrawal         HiveTime      `json:"next_vesting_withdrawal"`
	Withdrawn                     int64         `json:"withdrawn"`
	ToWithdraw                    int64         `json:"to_withdraw"`
	WithdrawRoutes                uint32        `json:"withdraw_routes"`
	CurationRewards               int64         `json:"curation_rewards"`
	PostingRewards                int64         `json:"posting_rewards"`
	ProxiedVsfVotes               []int64       `json:"proxied_vsf_votes"`
	WitnessesVotedFor             uint32        `json:"witnesses_voted_for"`
	WitnessVotes                  []string      `json:"witness_votes"`
	LastPost                      HiveTime      `json:"last_post"`
	LastRootPost                  HiveTime      `json:"last_root_post"`
	PostVotingPower               string        `json:"post_voting_power"`
	Reputation                    int64         `json:"reputation"`
	PendingClaimedAccounts        uint32        `json:"pending_claimed_accounts"`
	PendingTransfers              uint32        `json:"pending_transfers"`
	DelayedVotes                  []interface{} `json:"delayed_votes"`
	VotingManabar                 RC            `json:"voting_manabar"`
	DownvoteManabar               RC            `json:"downvote_manabar"`
	GovernanceVoteExpirationTs    HiveTime      `json:"governance_vote_expiration_ts"`
}

// AuthorityData holds an authority structure as returned by the API (owner, active, or posting).
type AuthorityData struct {
	WeightThreshold uint32          `json:"weight_threshold"`
	AccountAuths    [][]interface{} `json:"account_auths"`
	KeyAuths        [][]interface{} `json:"key_auths"`
}

// Price represents an exchange rate between two assets (e.g. for witness feed publishing).
type Price struct {
	Base  Asset `json:"base"`
	Quote Asset `json:"quote"`
}

// ChainProperties represents witness-reported blockchain configuration for witness_update.
type ChainProperties struct {
	AccountCreationFee Asset  `json:"account_creation_fee"`
	MaximumBlockSize   uint32 `json:"maximum_block_size"`
	HbdInterestRate    uint16 `json:"hbd_interest_rate"`
}

// AccountAuth is an account name paired with a weight for use in an Authority.
type AccountAuth struct {
	Account string
	Weight  uint16
}

// KeyAuth is a public key paired with a weight for use in an Authority.
type KeyAuth struct {
	Key    *secp256k1.PublicKey
	Weight uint16
}

// Authority defines the signing requirements for an account action.
// Use in AccountUpdate2Operation to change owner, active, or posting authorities.
type Authority struct {
	WeightThreshold uint32
	AccountAuths    []AccountAuth
	KeyAuths        []KeyAuth
}

// MarshalJSON encodes Authority in the format the Hive node expects:
// account_auths and key_auths as [name/key-string, weight] tuple arrays.
func (a Authority) MarshalJSON() ([]byte, error) {
	accountAuths := make([][2]interface{}, len(a.AccountAuths))
	for i, aa := range a.AccountAuths {
		accountAuths[i] = [2]interface{}{aa.Account, aa.Weight}
	}
	keyAuths := make([][2]interface{}, len(a.KeyAuths))
	for i, ka := range a.KeyAuths {
		keyAuths[i] = [2]interface{}{GetPublicKeyString(ka.Key), ka.Weight}
	}
	return json.Marshal(struct {
		WeightThreshold uint32           `json:"weight_threshold"`
		AccountAuths    [][2]interface{} `json:"account_auths"`
		KeyAuths        [][2]interface{} `json:"key_auths"`
	}{a.WeightThreshold, accountAuths, keyAuths})
}
