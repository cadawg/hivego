package hivego

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestRPCErrorString(t *testing.T) {
	e := &RPCError{Code: -32003, Message: "Assert Exception"}
	got := e.Error()
	want := "rpc error -32003: Assert Exception"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestAuthorityMarshalJSON(t *testing.T) {
	kp, err := KeyPairFromWif("5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W")
	if err != nil {
		t.Fatal(err)
	}
	auth := Authority{
		WeightThreshold: 1,
		AccountAuths:    []AccountAuth{{Account: "bob", Weight: 5}},
		KeyAuths:        []KeyAuth{{Key: kp.PublicKey, Weight: 1}},
	}
	b, err := json.Marshal(auth)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if len(b) == 0 {
		t.Error("expected non-empty JSON")
	}
}

func TestSignNilKey(t *testing.T) {
	client := NewClient("https://api.hive.blog").WithNoBroadcast()
	tx := &Transaction{
		RefBlockNum: 1, RefBlockPrefix: 12345,
		Expiration: "2025-01-01T00:00:00",
		Operations: []HiveOperation{VoteOperation{"alice", "bob", "p", 10000}},
	}
	err := client.Sign(tx, nil)
	if !errors.Is(err, ErrNilKey) {
		t.Fatalf("expected ErrNilKey, got %v", err)
	}
}

func TestAssetStringNegativeAmount(t *testing.T) {
	// Negative amount: exercises frac < 0 branch in Asset.String().
	a := Asset{Amount: -1500, Precision: 3, Symbol: "HIVE"}
	got := a.String()
	want := "-1.500 HIVE"
	if got != want {
		t.Errorf("Asset.String() = %q, want %q", got, want)
	}
}

func TestKeyPairFromBytesValid(t *testing.T) {
	key32 := make([]byte, 32)
	key32[31] = 1
	kp, err := KeyPairFromBytes(key32)
	if err != nil {
		t.Fatalf("KeyPairFromBytes: %v", err)
	}
	if kp.PrivateKey == nil || kp.PublicKey == nil {
		t.Fatal("nil key in pair")
	}
}

func TestKeyPairFromBytesWrongLength(t *testing.T) {
	_, err := KeyPairFromBytes(make([]byte, 16))
	if !errors.Is(err, ErrInvalidKeyLength) {
		t.Fatalf("expected ErrInvalidKeyLength, got %v", err)
	}
}

func TestDecodePublicKeyWithPrefixWrongPrefix(t *testing.T) {
	_, err := DecodePublicKeyWithPrefix("STM5nD5Sn9avUfQ1i3dZQRizopgNko19p3GrzgmwxA2YMapHrL8wx", "TST")
	if !errors.Is(err, ErrInvalidPrefix) {
		t.Fatalf("expected ErrInvalidPrefix, got %v", err)
	}
}

func TestDecodePublicKeyWithPrefixBadChecksum(t *testing.T) {
	// Flip last char of a valid key string to corrupt the checksum.
	_, err := DecodePublicKeyWithPrefix("STM5nD5Sn9avUfQ1i3dZQRizopgNko19p3GrzgmwxA2YMapHrL8wz", "STM")
	if !errors.Is(err, ErrChecksumMismatch) {
		t.Fatalf("expected ErrChecksumMismatch, got %v", err)
	}
}

func TestGetPublicKeyStringWithPrefixNil(t *testing.T) {
	got := GetPublicKeyStringWithPrefix(nil, "STM")
	if got != "" {
		t.Errorf("expected empty string for nil key, got %q", got)
	}
}

func TestKeyPairGetPublicKeyString(t *testing.T) {
	kp, err := KeyPairFromWif("5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W")
	if err != nil {
		t.Fatal(err)
	}
	got := kp.GetPublicKeyString()
	if got == "" {
		t.Error("expected non-empty public key string")
	}
	decoded, err := DecodePublicKey(got)
	if err != nil {
		t.Fatalf("DecodePublicKey(%q): %v", got, err)
	}
	if !decoded.IsEqual(kp.PublicKey) {
		t.Error("round-trip public key mismatch")
	}
}

// TestOpNames verifies every operation type returns its correct name.
// OpName() is only called via prepareJson → BroadcastTx, so it's not
// exercised by no-broadcast integration tests — test it directly here.
func TestOpNames(t *testing.T) {
	kp, _ := KeyPairFromWif("5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W")
	auth := Authority{WeightThreshold: 1, AccountAuths: []AccountAuth{}, KeyAuths: []KeyAuth{{Key: kp.PublicKey, Weight: 1}}}
	cases := []struct {
		op   HiveOperation
		want string
	}{
		{VoteOperation{}, "vote"},
		{CustomJsonOperation{}, "custom_json"},
		{TransferOperation{}, "transfer"},
		{ClaimRewardBalanceOperation{}, "claim_reward_balance"},
		{CommentOperation{}, "comment"},
		{CommentOptionsOperation{}, "comment_options"},
		{DeleteCommentOperation{}, "delete_comment"},
		{TransferToVestingOperation{}, "transfer_to_vesting"},
		{WithdrawVestingOperation{}, "withdraw_vesting"},
		{DelegateVestingSharesOperation{}, "delegate_vesting_shares"},
		{AccountWitnessVoteOperation{}, "account_witness_vote"},
		{TransferToSavingsOperation{}, "transfer_to_savings"},
		{TransferFromSavingsOperation{}, "transfer_from_savings"},
		{SetWithdrawVestingRouteOperation{}, "set_withdraw_vesting_route"},
		{CancelTransferFromSavingsOperation{}, "cancel_transfer_from_savings"},
		{RecurrentTransferOperation{}, "recurrent_transfer"},
		{AccountUpdate2Operation{}, "account_update2"},
		{ConvertOperation{}, "convert"},
		{CollateralizedConvertOperation{}, "collateralized_convert"},
		{LimitOrderCreateOperation{}, "limit_order_create"},
		{LimitOrderCancelOperation{}, "limit_order_cancel"},
		{AccountWitnessProxyOperation{}, "account_witness_proxy"},
		{CreateProposalOperation{}, "create_proposal"},
		{UpdateProposalVotesOperation{}, "update_proposal_votes"},
		{RemoveProposalOperation{}, "remove_proposal"},
		{UpdateProposalOperation{}, "update_proposal"},
		{ClaimAccountOperation{}, "claim_account"},
		{FeedPublishOperation{}, "feed_publish"},
		{AccountCreateOperation{}, "account_create"},
		{CreateClaimedAccountOperation{}, "create_claimed_account"},
		{AccountUpdateOperation{}, "account_update"},
		{WitnessUpdateOperation{}, "witness_update"},
		{WitnessSetPropertiesOperation{}, "witness_set_properties"},
		{RequestAccountRecoveryOperation{}, "request_account_recovery"},
		{RecoverAccountOperation{}, "recover_account"},
		{ChangeRecoveryAccountOperation{}, "change_recovery_account"},
		{EscrowTransferOperation{}, "escrow_transfer"},
		{EscrowApproveOperation{}, "escrow_approve"},
		{EscrowDisputeOperation{}, "escrow_dispute"},
		{EscrowReleaseOperation{}, "escrow_release"},
		{DeclineVotingRightsOperation{}, "decline_voting_rights"},
		{CustomOperation{}, "custom"},
		{CustomBinaryOperation{}, "custom_binary"},
		// ops with custom MarshalJSON — verify OpName too
		{AccountCreateOperation{Owner: auth, Active: auth, Posting: auth, MemoKey: kp.PublicKey}, "account_create"},
		{CreateClaimedAccountOperation{Owner: auth, Active: auth, Posting: auth, MemoKey: kp.PublicKey}, "create_claimed_account"},
	}
	for _, tc := range cases {
		if got := tc.op.OpName(); got != tc.want {
			t.Errorf("%T.OpName() = %q, want %q", tc.op, got, tc.want)
		}
	}
}

// TestMarshalJSONOps verifies JSON encoding for ops that have custom MarshalJSON.
func TestMarshalJSONOps(t *testing.T) {
	kp, _ := KeyPairFromWif("5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W")
	auth := Authority{
		WeightThreshold: 1,
		AccountAuths:    []AccountAuth{},
		KeyAuths:        []KeyAuth{{Key: kp.PublicKey, Weight: 1}},
	}
	import_json := func(op HiveOperation) {
		t.Helper()
		tx := &Transaction{Operations: []HiveOperation{op}}
		tx.prepareJson()
		if len(tx.OperationsJs) == 0 {
			t.Errorf("%T: prepareJson produced no operations", op)
		}
	}

	import_json(AccountCreateOperation{
		Fee: Asset{3000, 3, "HIVE"}, Creator: "alice", NewAccountName: "bob",
		Owner: auth, Active: auth, Posting: auth, MemoKey: kp.PublicKey,
	})
	import_json(CreateClaimedAccountOperation{
		Creator: "alice", NewAccountName: "bob",
		Owner: auth, Active: auth, Posting: auth, MemoKey: kp.PublicKey,
	})
	import_json(AccountUpdateOperation{
		Account: "alice", Active: &auth, MemoKey: kp.PublicKey,
	})
	import_json(WitnessUpdateOperation{
		Owner: "alice", Url: "https://example.com", BlockSigningKey: kp.PublicKey,
		Props: ChainProperties{AccountCreationFee: Asset{3000, 3, "HIVE"}, MaximumBlockSize: 131072},
		Fee:   Asset{0, 3, "HIVE"},
	})
	import_json(WitnessSetPropertiesOperation{
		Owner: "alice",
		Props: map[string][]byte{"key": kp.PublicKey.SerializeCompressed()},
	})
	import_json(AccountUpdate2Operation{
		Account: "alice", Active: &auth, MemoKey: kp.PublicKey,
	})
}
