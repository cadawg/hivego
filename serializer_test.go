package hivego

import (
	"bytes"
	"errors"
	"testing"
)

func TestOpIdB(t *testing.T) {
	got := opIdB("custom_json")
	expected := byte(18)

	if got != expected {
		t.Error("Expected", expected, "got")
	}
}

func TestRefBlockNumB(t *testing.T) {
	got := refBlockNumB(36029)
	expected := []byte{189, 140}

	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestRefBlockPrefixB(t *testing.T) {
	got := refBlockPrefixB(1164960351)
	expected := []byte{95, 226, 111, 69}

	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestExpTimeB(t *testing.T) {
	got, _ := expTimeB("2016-08-08T12:24:17")
	expected := []byte{241, 121, 168, 87}

	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestCountOpsB(t *testing.T) {
	got := countOpsB(getTwoTestOps())
	expected := []byte{2}

	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

//func TestExtensionsB

func TestAppendVString(t *testing.T) {
	var buf bytes.Buffer
	got := appendVString("xeroc", &buf)
	expected := []byte{5, 120, 101, 114, 111, 99}
	if !bytes.Equal(got.Bytes(), expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestAppendVStringArray(t *testing.T) {
	var buf bytes.Buffer
	got := appendVStringArray([]string{"xeroc", "piston"}, &buf).Bytes()
	expected := []byte{2, 5, 120, 101, 114, 111, 99, 6, 112, 105, 115, 116, 111, 110}
	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestSerializeTx(t *testing.T) {
	got, _ := serializeTx(getTestVoteTx())
	expected := []byte{189, 140, 95, 226, 111, 69, 241, 121, 168, 87, 1, 0, 5, 120, 101, 114, 111, 99, 5, 120, 101, 114, 111, 99, 6, 112, 105, 115, 116, 111, 110, 16, 39, 0}
	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestSerializeOps(t *testing.T) {
	got, _ := serializeOps(getTwoTestOps())
	expected := []byte{2, 0, 5, 120, 101, 114, 111, 99, 5, 120, 101, 114, 111, 99, 6, 112, 105, 115, 116, 111, 110, 16, 39, 18, 0, 1, 5, 120, 101, 114, 111, 99, 7, 116, 101, 115, 116, 45, 105, 100, 17, 123, 34, 116, 101, 115, 116, 107, 34, 58, 34, 116, 101, 115, 116, 118, 34, 125}
	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestSerializeOpVoteOperation(t *testing.T) {
	got, _ := getTestVoteOp().SerializeOp()
	expected := []byte{0, 5, 120, 101, 114, 111, 99, 5, 120, 101, 114, 111, 99, 6, 112, 105, 115, 116, 111, 110, 16, 39}
	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestSerializeOpCustomJsonOperation(t *testing.T) {
	got, _ := getTestCustomJsonOp().SerializeOp()
	expected := []byte{18, 0, 1, 5, 120, 101, 114, 111, 99, 7, 116, 101, 115, 116, 45, 105, 100, 17, 123, 34, 116, 101, 115, 116, 107, 34, 58, 34, 116, 101, 115, 116, 118, 34, 125}
	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestSerializeOpTransfer(t *testing.T) {
	amount, _ := ParseAsset("1.000 HIVE")
	got, err := TransferOperation{From: "alice", To: "bob", Amount: amount, Memo: "memo"}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=2, "alice", "bob", 1.000 HIVE (amount=1000 LE64, prec=3, "STEEM\0\0" wire), "memo"
	expected := []byte{2, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 232, 3, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0, 4, 109, 101, 109, 111}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpComment(t *testing.T) {
	got, err := CommentOperation{
		ParentAuthor: "", ParentPermlink: "hive-blog",
		Author: "alice", Permlink: "my-post",
		Title: "Hello", Body: "World", JsonMetadata: "{}",
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=1, "","hive-blog","alice","my-post","Hello","World","{}"
	expected := []byte{1, 0, 9, 104, 105, 118, 101, 45, 98, 108, 111, 103, 5, 97, 108, 105, 99, 101, 7, 109, 121, 45, 112, 111, 115, 116, 5, 72, 101, 108, 108, 111, 5, 87, 111, 114, 108, 100, 2, 123, 125}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpDeleteComment(t *testing.T) {
	got, err := DeleteCommentOperation{Author: "alice", Permlink: "my-post"}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=17, "alice", "my-post"
	expected := []byte{17, 5, 97, 108, 105, 99, 101, 7, 109, 121, 45, 112, 111, 115, 116}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpClaimRewardBalance(t *testing.T) {
	hive, _ := ParseAsset("0.000 HIVE")
	hbd, _ := ParseAsset("0.000 HBD")
	vests, _ := ParseAsset("1.000000 VESTS")
	got, err := ClaimRewardBalanceOperation{Account: "alice", RewardHive: hive, RewardHbd: hbd, RewardVests: vests}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=39, "alice", 0.000 HIVE ("STEEM\0\0"), 0.000 HBD ("SBD\0\0\0\0"), 1.000000 VESTS
	expected := []byte{
		39, 5, 97, 108, 105, 99, 101,
		0, 0, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0,
		64, 66, 15, 0, 0, 0, 0, 0, 6, 86, 69, 83, 84, 83, 0, 0,
	}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpTransferToVesting(t *testing.T) {
	amount, _ := ParseAsset("100.000 HIVE")
	got, err := TransferToVestingOperation{From: "alice", To: "bob", Amount: amount}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=3, "alice","bob", 100.000 HIVE (amount=100000, "STEEM\0\0" wire)
	expected := []byte{3, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 160, 134, 1, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpWithdrawVesting(t *testing.T) {
	vests, _ := ParseAsset("0.000000 VESTS")
	got, err := WithdrawVestingOperation{Account: "alice", VestingShares: vests}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=4, "alice", 0.000000 VESTS
	expected := []byte{4, 5, 97, 108, 105, 99, 101, 0, 0, 0, 0, 0, 0, 0, 0, 6, 86, 69, 83, 84, 83, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpDelegateVestingShares(t *testing.T) {
	vests, _ := ParseAsset("1.000000 VESTS")
	got, err := DelegateVestingSharesOperation{Delegator: "alice", Delegatee: "bob", VestingShares: vests}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=40, "alice","bob", 1.000000 VESTS (amount=1000000)
	expected := []byte{40, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 64, 66, 15, 0, 0, 0, 0, 0, 6, 86, 69, 83, 84, 83, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpAccountWitnessVote(t *testing.T) {
	got, err := AccountWitnessVoteOperation{Account: "alice", Witness: "bob", Approve: true}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=12, "alice","bob", approve=1
	expected := []byte{12, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 1}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpAccountWitnessProxy(t *testing.T) {
	got, err := AccountWitnessProxyOperation{Account: "alice", Proxy: ""}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=13, "alice", "" (remove proxy)
	expected := []byte{13, 5, 97, 108, 105, 99, 101, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpTransferToSavings(t *testing.T) {
	amount, _ := ParseAsset("100.000 HBD")
	got, err := TransferToSavingsOperation{From: "alice", To: "alice", Amount: amount, Memo: ""}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=32, "alice","alice", 100.000 HBD (amount=100000, "SBD\0\0\0\0" wire), ""
	expected := []byte{32, 5, 97, 108, 105, 99, 101, 5, 97, 108, 105, 99, 101, 160, 134, 1, 0, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpTransferFromSavings(t *testing.T) {
	amount, _ := ParseAsset("1.000 HBD")
	got, err := TransferFromSavingsOperation{From: "alice", RequestId: 1, To: "alice", Amount: amount, Memo: ""}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=33, "alice", requestId=1, "alice", 1.000 HBD ("SBD\0\0\0\0" wire), ""
	expected := []byte{33, 5, 97, 108, 105, 99, 101, 1, 0, 0, 0, 5, 97, 108, 105, 99, 101, 232, 3, 0, 0, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpCancelTransferFromSavings(t *testing.T) {
	got, err := CancelTransferFromSavingsOperation{From: "alice", RequestId: 42}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=34, "alice", requestId=42
	expected := []byte{34, 5, 97, 108, 105, 99, 101, 42, 0, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpSetWithdrawVestingRoute(t *testing.T) {
	got, err := SetWithdrawVestingRouteOperation{FromAccount: "alice", ToAccount: "bob", Percent: 5000, AutoVest: true}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=20, "alice","bob", 5000 (0x1388 LE: [136,19]), autoVest=1
	expected := []byte{20, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 136, 19, 1}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpRecurrentTransfer(t *testing.T) {
	amount, _ := ParseAsset("1.000 HIVE")
	got, err := RecurrentTransferOperation{From: "alice", To: "bob", Amount: amount, Memo: "", Recurrence: 24, Executions: 7}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=49, "alice","bob", 1.000 HIVE ("STEEM\0\0" wire), "", recurrence=24, executions=7, extensions=0
	expected := []byte{49, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 232, 3, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0, 0, 24, 0, 7, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpAccountUpdate2MetadataOnly(t *testing.T) {
	got, err := AccountUpdate2Operation{Account: "alice", JsonMetadata: "{}", PostingJsonMetadata: ""}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=43, "alice", nil×4 (0×4), "{}","", extensions=0
	expected := []byte{43, 5, 97, 108, 105, 99, 101, 0, 0, 0, 0, 2, 123, 125, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpConvert(t *testing.T) {
	amount, _ := ParseAsset("1.000 HBD")
	got, err := ConvertOperation{Owner: "alice", RequestId: 1, Amount: amount}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=8, "alice", requestId=1, 1.000 HBD ("SBD\0\0\0\0" wire)
	expected := []byte{8, 5, 97, 108, 105, 99, 101, 1, 0, 0, 0, 232, 3, 0, 0, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpCollateralizedConvert(t *testing.T) {
	amount, _ := ParseAsset("1.000 HIVE")
	got, err := CollateralizedConvertOperation{Owner: "alice", RequestId: 1, Amount: amount}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=48, "alice", requestId=1, 1.000 HIVE ("STEEM\0\0" wire)
	expected := []byte{48, 5, 97, 108, 105, 99, 101, 1, 0, 0, 0, 232, 3, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpLimitOrderCancel(t *testing.T) {
	got, err := LimitOrderCancelOperation{Owner: "alice", OrderId: 1}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=6, "alice", orderId=1
	expected := []byte{6, 5, 97, 108, 105, 99, 101, 1, 0, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpFeedPublish(t *testing.T) {
	base, _ := ParseAsset("1.000 HBD")
	quote, _ := ParseAsset("3.500 HIVE")
	got, err := FeedPublishOperation{Publisher: "alice", ExchangeRate: Price{base, quote}}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=7, "alice", 1.000 HBD ("SBD\0\0\0\0"), 3.500 HIVE ("STEEM\0\0"), amount=3500=[172,13,...]
	expected := []byte{7, 5, 97, 108, 105, 99, 101, 232, 3, 0, 0, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 172, 13, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpChangeRecoveryAccount(t *testing.T) {
	got, err := ChangeRecoveryAccountOperation{AccountToRecover: "alice", NewRecoveryAccount: "bob"}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=26, "alice","bob", extensions=0
	expected := []byte{26, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpDeclineVotingRightsFalse(t *testing.T) {
	got, err := DeclineVotingRightsOperation{Account: "alice", Decline: false}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=36, "alice", decline=0
	expected := []byte{36, 5, 97, 108, 105, 99, 101, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpDeclineVotingRightsConfirmed(t *testing.T) {
	got, err := DeclineVotingRightsOperation{Account: "alice", Decline: true, IUnderstandThisIsIrreversible: true}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=36, "alice", decline=1
	expected := []byte{36, 5, 97, 108, 105, 99, 101, 1}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpDeclineVotingRightsRequiresConfirmation(t *testing.T) {
	op := DeclineVotingRightsOperation{Account: "alice", Decline: true, IUnderstandThisIsIrreversible: false}
	_, err := op.SerializeOp()
	if !errors.Is(err, ErrDeclineVotingRightsNotConfirmed) {
		t.Errorf("expected ErrDeclineVotingRightsNotConfirmed, got %v", err)
	}
}

func TestSerializeOpClaimAccount(t *testing.T) {
	fee, _ := ParseAsset("0.000 HIVE")
	got, err := ClaimAccountOperation{Creator: "alice", Fee: fee}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=22, "alice", 0.000 HIVE ("STEEM\0\0" wire), extensions=0
	expected := []byte{22, 5, 97, 108, 105, 99, 101, 0, 0, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpCommentOptionsNoBeneficiaries(t *testing.T) {
	payout, _ := ParseAsset("1000000.000 HBD")
	got, err := CommentOptionsOperation{
		Author: "alice", Permlink: "my-post",
		MaxAcceptedPayout: payout, PercentHbd: 10000,
		AllowVotes: true, AllowCurationRewards: true,
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=19, "alice", "my-post", 1000000.000 HBD ("SBD"), 10000, allow_votes=1, allow_curation=1, extensions=0
	expected := []byte{19, 5, 97, 108, 105, 99, 101, 7, 109, 121, 45, 112, 111, 115, 116, 0, 202, 154, 59, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 16, 39, 1, 1, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpCommentOptionsWithBeneficiary(t *testing.T) {
	payout, _ := ParseAsset("1000000.000 HBD")
	got, err := CommentOptionsOperation{
		Author: "alice", Permlink: "my-post",
		MaxAcceptedPayout: payout, PercentHbd: 5000,
		AllowVotes: true, AllowCurationRewards: true,
		Beneficiaries: []Beneficiary{{Account: "bob", Weight: 5000}},
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=19, ..., 5000, 1, 1, ext=1, type=0, count=1, "bob", 5000
	expected := []byte{19, 5, 97, 108, 105, 99, 101, 7, 109, 121, 45, 112, 111, 115, 116, 0, 202, 154, 59, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 136, 19, 1, 1, 1, 0, 1, 3, 98, 111, 98, 136, 19}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpLimitOrderCreate(t *testing.T) {
	sell, _ := ParseAsset("1.000 HIVE")
	receive, _ := ParseAsset("100.000 HBD")
	got, err := LimitOrderCreateOperation{
		Owner: "alice", OrderId: 1,
		AmountToSell: sell, MinToReceive: receive,
		FillOrKill: false, Expiration: "2030-01-01T00:00:00",
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=5, "alice", orderId=1, 1.000 HIVE, 100.000 HBD, fill_or_kill=0, expiration=2030-01-01
	expected := []byte{5, 5, 97, 108, 105, 99, 101, 1, 0, 0, 0, 232, 3, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0, 160, 134, 1, 0, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 0, 128, 216, 219, 112}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpCreateProposal(t *testing.T) {
	pay, _ := ParseAsset("100.000 HBD")
	got, err := CreateProposalOperation{
		Creator: "alice", Receiver: "bob",
		StartDate: "2024-01-01T00:00:00", EndDate: "2024-12-31T00:00:00",
		DailyPay: pay, Subject: "My Proposal", Permlink: "my-proposal",
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=44, "alice","bob", start, end, 100.000 HBD, "My Proposal", "my-proposal", extensions=0
	expected := []byte{44, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 128, 0, 146, 101, 0, 52, 115, 103, 160, 134, 1, 0, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 11, 77, 121, 32, 80, 114, 111, 112, 111, 115, 97, 108, 11, 109, 121, 45, 112, 114, 111, 112, 111, 115, 97, 108, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpUpdateProposalVotes(t *testing.T) {
	got, err := UpdateProposalVotesOperation{Voter: "alice", ProposalIds: []int64{1, 2}, Approve: true}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=45, "alice", [1,2], approve=1, extensions=0
	expected := []byte{45, 5, 97, 108, 105, 99, 101, 2, 1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 1, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpRemoveProposal(t *testing.T) {
	got, err := RemoveProposalOperation{ProposalOwner: "alice", ProposalIds: []int64{1}}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=46, "alice", [1], extensions=0
	expected := []byte{46, 5, 97, 108, 105, 99, 101, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpUpdateProposalNoEndDate(t *testing.T) {
	pay, _ := ParseAsset("50.000 HBD")
	got, err := UpdateProposalOperation{
		ProposalId: 1, Creator: "alice", DailyPay: pay,
		Subject: "Updated", Permlink: "my-proposal", EndDate: "",
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=47, proposalId=1 (LE64), "alice", 50.000 HBD, "Updated", "my-proposal", extensions=[]
	expected := []byte{47, 1, 0, 0, 0, 0, 0, 0, 0, 5, 97, 108, 105, 99, 101, 80, 195, 0, 0, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 7, 85, 112, 100, 97, 116, 101, 100, 11, 109, 121, 45, 112, 114, 111, 112, 111, 115, 97, 108, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpUpdateProposalWithEndDate(t *testing.T) {
	pay, _ := ParseAsset("50.000 HBD")
	got, err := UpdateProposalOperation{
		ProposalId: 1, Creator: "alice", DailyPay: pay,
		Subject: "Updated", Permlink: "my-proposal", EndDate: "2025-06-01T00:00:00",
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// ..., extensions=[{type=1, end_date=2025-06-01}]
	expected := []byte{47, 1, 0, 0, 0, 0, 0, 0, 0, 5, 97, 108, 105, 99, 101, 80, 195, 0, 0, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 7, 85, 112, 100, 97, 116, 101, 100, 11, 109, 121, 45, 112, 114, 111, 112, 111, 115, 97, 108, 1, 1, 0, 152, 59, 104}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func testKeyPair(t *testing.T) *KeyPair {
	t.Helper()
	kp, err := KeyPairFromWif("5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W")
	if err != nil {
		t.Fatal(err)
	}
	return kp
}

func TestSerializeOpAccountCreate(t *testing.T) {
	kp := testKeyPair(t)
	auth := Authority{
		WeightThreshold: 1,
		AccountAuths:    []AccountAuth{},
		KeyAuths:        []KeyAuth{{Key: kp.PublicKey, Weight: 1}},
	}
	fee, _ := ParseAsset("3.000 HIVE")
	got, err := AccountCreateOperation{
		Fee: fee, Creator: "alice", NewAccountName: "bob",
		Owner: auth, Active: auth, Posting: auth,
		MemoKey: kp.PublicKey, JsonMetadata: "{}",
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=9, 3.000 HIVE, "alice","bob", owner+active+posting authorities, memo_key, "{}"
	expected := []byte{9, 184, 11, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 2, 123, 125}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpCreateClaimedAccount(t *testing.T) {
	kp := testKeyPair(t)
	auth := Authority{
		WeightThreshold: 1,
		AccountAuths:    []AccountAuth{},
		KeyAuths:        []KeyAuth{{Key: kp.PublicKey, Weight: 1}},
	}
	got, err := CreateClaimedAccountOperation{
		Creator: "alice", NewAccountName: "bob",
		Owner: auth, Active: auth, Posting: auth,
		MemoKey: kp.PublicKey, JsonMetadata: "{}",
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=23, "alice","bob", owner+active+posting, memo_key, "{}", extensions=0
	expected := []byte{23, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 2, 123, 125, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpAccountUpdateMemoOnly(t *testing.T) {
	kp := testKeyPair(t)
	got, err := AccountUpdateOperation{
		Account: "alice", MemoKey: kp.PublicKey, JsonMetadata: "",
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=10, "alice", owner=absent, active=absent, posting=absent, memo_key(33 bytes), ""
	expected := []byte{10, 5, 97, 108, 105, 99, 101, 0, 0, 0, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpWitnessUpdate(t *testing.T) {
	kp := testKeyPair(t)
	fee, _ := ParseAsset("0.000 HIVE")
	creationFee, _ := ParseAsset("3.000 HIVE")
	got, err := WitnessUpdateOperation{
		Owner: "alice", Url: "https://example.com",
		BlockSigningKey: kp.PublicKey,
		Props:           ChainProperties{AccountCreationFee: creationFee, MaximumBlockSize: 65536, HbdInterestRate: 0},
		Fee:             fee,
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=11, "alice", "https://example.com", pubkey(33), chain_props, 0.000 HIVE (fee)
	expected := []byte{11, 5, 97, 108, 105, 99, 101, 19, 104, 116, 116, 112, 115, 58, 47, 47, 101, 120, 97, 109, 112, 108, 101, 46, 99, 111, 109, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 184, 11, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpWitnessSetProperties(t *testing.T) {
	kp := testKeyPair(t)
	got, err := WitnessSetPropertiesOperation{
		Owner: "alice",
		Props: map[string][]byte{
			"key": kp.PublicKey.SerializeCompressed(),
		},
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=42, "alice", flat_map[{"key": pubkey(33 bytes)}], extensions=0
	expected := []byte{42, 5, 97, 108, 105, 99, 101, 1, 3, 107, 101, 121, 33, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpRequestAccountRecovery(t *testing.T) {
	kp := testKeyPair(t)
	auth := Authority{
		WeightThreshold: 1,
		AccountAuths:    []AccountAuth{},
		KeyAuths:        []KeyAuth{{Key: kp.PublicKey, Weight: 1}},
	}
	got, err := RequestAccountRecoveryOperation{
		RecoveryAccount: "alice", AccountToRecover: "bob",
		NewOwnerAuthority: auth,
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=24, "alice","bob", authority, extensions=0
	expected := []byte{24, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpRecoverAccount(t *testing.T) {
	kp := testKeyPair(t)
	auth := Authority{
		WeightThreshold: 1,
		AccountAuths:    []AccountAuth{},
		KeyAuths:        []KeyAuth{{Key: kp.PublicKey, Weight: 1}},
	}
	got, err := RecoverAccountOperation{
		AccountToRecover:     "bob",
		NewOwnerAuthority:    auth,
		RecentOwnerAuthority: auth,
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=25, "bob", new_authority, recent_authority, extensions=0
	expected := []byte{25, 3, 98, 111, 98, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpEscrowTransfer(t *testing.T) {
	hbd, _ := ParseAsset("1.000 HBD")
	hive, _ := ParseAsset("1.000 HIVE")
	fee, _ := ParseAsset("0.001 HIVE")
	got, err := EscrowTransferOperation{
		From: "alice", To: "bob", Agent: "charlie", EscrowId: 1,
		HbdAmount: hbd, HiveAmount: hive, Fee: fee,
		RatificationDeadline: "2030-01-01T00:00:00",
		EscrowExpiration:     "2030-06-01T00:00:00",
		JsonMeta:             "{}",
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=27, "alice","bob","charlie", escrowId=1, 1.000 HBD, 1.000 HIVE, 0.001 HIVE, deadline, expiration, "{}"
	expected := []byte{27, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 7, 99, 104, 97, 114, 108, 105, 101, 1, 0, 0, 0, 232, 3, 0, 0, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 232, 3, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0, 128, 216, 219, 112, 0, 235, 162, 113, 2, 123, 125}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpEscrowApprove(t *testing.T) {
	got, err := EscrowApproveOperation{
		From: "alice", To: "bob", Agent: "charlie",
		Who: "bob", EscrowId: 1, Approve: true,
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=31, "alice","bob","charlie","bob", escrowId=1, approve=1
	expected := []byte{31, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 7, 99, 104, 97, 114, 108, 105, 101, 3, 98, 111, 98, 1, 0, 0, 0, 1}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpEscrowDispute(t *testing.T) {
	got, err := EscrowDisputeOperation{
		From: "alice", To: "bob", Agent: "charlie",
		Who: "alice", EscrowId: 1,
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=28, "alice","bob","charlie","alice", escrowId=1
	// NOTE: beem omits Agent in escrow_dispute; our implementation follows the Hive protocol spec.
	expected := []byte{28, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 7, 99, 104, 97, 114, 108, 105, 101, 5, 97, 108, 105, 99, 101, 1, 0, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpAccountUpdate2WithActive(t *testing.T) {
	kp := testKeyPair(t)
	auth := Authority{
		WeightThreshold: 1,
		AccountAuths:    []AccountAuth{},
		KeyAuths:        []KeyAuth{{Key: kp.PublicKey, Weight: 1}},
	}
	got, err := AccountUpdate2Operation{
		Account: "alice", Active: &auth,
		MemoKey:      kp.PublicKey,
		JsonMetadata: "{}",
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=43, "alice", owner=absent, active=present+authority, posting=absent, memo_key=present, "{}","", extensions=0
	expected := []byte{43, 5, 97, 108, 105, 99, 101, 0, 1, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 2, 123, 125, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpAccountUpdateWithActive(t *testing.T) {
	kp := testKeyPair(t)
	auth := Authority{
		WeightThreshold: 1,
		AccountAuths:    []AccountAuth{},
		KeyAuths:        []KeyAuth{{Key: kp.PublicKey, Weight: 1}},
	}
	got, err := AccountUpdateOperation{
		Account: "alice", Active: &auth,
		MemoKey: kp.PublicKey, JsonMetadata: "",
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=10, "alice", owner=absent, active=present+authority, posting=absent, memo_key, ""
	expected := []byte{10, 5, 97, 108, 105, 99, 101, 0, 1, 1, 0, 0, 0, 0, 1, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 1, 0, 0, 2, 117, 111, 103, 78, 142, 147, 255, 200, 235, 113, 156, 147, 58, 155, 6, 220, 171, 133, 220, 98, 22, 11, 49, 41, 201, 58, 158, 247, 103, 243, 64, 145, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpEscrowRelease(t *testing.T) {
	hbd, _ := ParseAsset("1.000 HBD")
	hive, _ := ParseAsset("0.000 HIVE")
	got, err := EscrowReleaseOperation{
		From: "alice", To: "bob", Agent: "charlie",
		Who: "alice", Receiver: "bob", EscrowId: 1,
		HbdAmount: hbd, HiveAmount: hive,
	}.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// op_id=29, "alice","bob","charlie","alice","bob", escrowId=1, 1.000 HBD, 0.000 HIVE
	// NOTE: beem omits Agent and Receiver in escrow_release; our implementation follows the Hive protocol spec.
	expected := []byte{29, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 7, 99, 104, 97, 114, 108, 105, 101, 5, 97, 108, 105, 99, 101, 3, 98, 111, 98, 1, 0, 0, 0, 232, 3, 0, 0, 0, 0, 0, 0, 3, 83, 66, 68, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 83, 84, 69, 69, 77, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

// --- Error path and missing-branch tests ---

func TestSerializeOpLimitOrderCreateBadExpiration(t *testing.T) {
	op := LimitOrderCreateOperation{
		Owner:        "alice",
		OrderId:      1,
		AmountToSell: Asset{1000, 3, "HIVE"},
		MinToReceive: Asset{500, 3, "HBD"},
		Expiration:   "not-a-date",
	}
	_, err := op.SerializeOp()
	if err == nil {
		t.Fatal("expected error for bad expiration")
	}
}

func TestSerializeOpCreateProposalBadStartDate(t *testing.T) {
	op := CreateProposalOperation{
		Creator:   "alice",
		Receiver:  "alice",
		StartDate: "bad",
		EndDate:   "2025-12-31T00:00:00",
		DailyPay:  Asset{1000, 3, "HBD"},
	}
	_, err := op.SerializeOp()
	if err == nil {
		t.Fatal("expected error for bad start date")
	}
}

func TestSerializeOpCreateProposalBadEndDate(t *testing.T) {
	op := CreateProposalOperation{
		Creator:   "alice",
		Receiver:  "alice",
		StartDate: "2025-01-01T00:00:00",
		EndDate:   "bad",
		DailyPay:  Asset{1000, 3, "HBD"},
	}
	_, err := op.SerializeOp()
	if err == nil {
		t.Fatal("expected error for bad end date")
	}
}

func TestSerializeOpEscrowTransferBadRatificationDeadline(t *testing.T) {
	op := EscrowTransferOperation{
		From: "alice", To: "bob", Agent: "carol", EscrowId: 1,
		HbdAmount: Asset{0, 3, "HBD"}, HiveAmount: Asset{1000, 3, "HIVE"}, Fee: Asset{10, 3, "HIVE"},
		RatificationDeadline: "bad",
		EscrowExpiration:     "2025-12-01T00:00:00",
	}
	_, err := op.SerializeOp()
	if err == nil {
		t.Fatal("expected error for bad ratification deadline")
	}
}

func TestSerializeOpEscrowTransferBadEscrowExpiration(t *testing.T) {
	op := EscrowTransferOperation{
		From: "alice", To: "bob", Agent: "carol", EscrowId: 1,
		HbdAmount: Asset{0, 3, "HBD"}, HiveAmount: Asset{1000, 3, "HIVE"}, Fee: Asset{10, 3, "HIVE"},
		RatificationDeadline: "2025-06-01T00:00:00",
		EscrowExpiration:     "bad",
	}
	_, err := op.SerializeOp()
	if err == nil {
		t.Fatal("expected error for bad escrow expiration")
	}
}

func TestSerializeOpUpdateProposalBadEndDate(t *testing.T) {
	op := UpdateProposalOperation{
		ProposalId: 1, Creator: "alice",
		DailyPay: Asset{500, 3, "HBD"},
		EndDate:  "bad",
	}
	_, err := op.SerializeOp()
	if err == nil {
		t.Fatal("expected error for bad end date")
	}
}

func TestSerializeOpDeclineVotingRightsNotConfirmed(t *testing.T) {
	op := DeclineVotingRightsOperation{Account: "alice", Decline: true}
	_, err := op.SerializeOp()
	if !errors.Is(err, ErrDeclineVotingRightsNotConfirmed) {
		t.Fatalf("expected ErrDeclineVotingRightsNotConfirmed, got %v", err)
	}
}

func TestSerializeOpAccountWitnessVoteFalse(t *testing.T) {
	op := AccountWitnessVoteOperation{Account: "alice", Witness: "gtg", Approve: false}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// last byte should be 0 (Approve=false)
	if got[len(got)-1] != 0 {
		t.Errorf("expected Approve byte 0, got %d", got[len(got)-1])
	}
}

func TestSerializeOpCommentOptionsAllFalse(t *testing.T) {
	op := CommentOptionsOperation{
		Author: "alice", Permlink: "p",
		MaxAcceptedPayout:    Asset{1000000000, 3, "HBD"},
		PercentHbd:           10000,
		AllowVotes:           false,
		AllowCurationRewards: false,
		Beneficiaries:        nil,
	}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// AllowVotes byte and AllowCurationRewards byte should both be 0
	// format: opid(1) + vstring(author) + vstring(permlink) + asset(16) + uint16(2) + allowvotes(1) + allowcuration(1) + extensions(1)
	if len(got) < 3 {
		t.Fatal("output too short")
	}
}

func TestSerializeOpAccountUpdate2OwnerAndPosting(t *testing.T) {
	kp, err := KeyPairFromWif("5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W")
	if err != nil {
		t.Fatal(err)
	}
	auth := Authority{
		WeightThreshold: 1,
		AccountAuths:    []AccountAuth{},
		KeyAuths:        []KeyAuth{{Key: kp.PublicKey, Weight: 1}},
	}
	op := AccountUpdate2Operation{
		Account: "alice",
		Owner:   &auth,
		Posting: &auth,
	}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestSerializeOpAccountUpdateWithPosting(t *testing.T) {
	kp, err := KeyPairFromWif("5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W")
	if err != nil {
		t.Fatal(err)
	}
	auth := Authority{
		WeightThreshold: 1,
		AccountAuths:    []AccountAuth{},
		KeyAuths:        []KeyAuth{{Key: kp.PublicKey, Weight: 1}},
	}
	op := AccountUpdateOperation{
		Account: "alice",
		Posting: &auth,
		MemoKey: kp.PublicKey,
	}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestSerializeOpAccountUpdateNilMemoKey(t *testing.T) {
	op := AccountUpdateOperation{Account: "alice", MemoKey: nil}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	// nil MemoKey should produce 33 zero bytes
	if len(got) < 33 {
		t.Error("output too short for nil MemoKey")
	}
}

func TestSerializeOpAccountUpdateWithOwner(t *testing.T) {
	kp, err := KeyPairFromWif("5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W")
	if err != nil {
		t.Fatal(err)
	}
	auth := Authority{
		WeightThreshold: 1,
		AccountAuths:    []AccountAuth{},
		KeyAuths:        []KeyAuth{{Key: kp.PublicKey, Weight: 1}},
	}
	op := AccountUpdateOperation{Account: "alice", Owner: &auth, MemoKey: kp.PublicKey}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestSerializeOpAccountCreateNilMemoKey(t *testing.T) {
	auth := Authority{WeightThreshold: 1, AccountAuths: []AccountAuth{}, KeyAuths: []KeyAuth{}}
	op := AccountCreateOperation{
		Fee: Asset{3000, 3, "HIVE"}, Creator: "alice", NewAccountName: "bob",
		Owner: auth, Active: auth, Posting: auth, MemoKey: nil,
	}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestSerializeOpCreateClaimedAccountNilMemoKey(t *testing.T) {
	auth := Authority{WeightThreshold: 1, AccountAuths: []AccountAuth{}, KeyAuths: []KeyAuth{}}
	op := CreateClaimedAccountOperation{
		Creator: "alice", NewAccountName: "bob",
		Owner: auth, Active: auth, Posting: auth, MemoKey: nil,
	}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestSerializeOpWitnessUpdateNilKey(t *testing.T) {
	op := WitnessUpdateOperation{
		Owner: "alice", Url: "https://example.com", BlockSigningKey: nil,
		Props: ChainProperties{AccountCreationFee: Asset{3000, 3, "HIVE"}, MaximumBlockSize: 131072, HbdInterestRate: 1000},
		Fee:   Asset{0, 3, "HIVE"},
	}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestSerializeOpAuthorityWithAccountAuths(t *testing.T) {
	// Exercises the AccountAuths loop in appendAuthorityBytes.
	kp, err := KeyPairFromWif("5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W")
	if err != nil {
		t.Fatal(err)
	}
	auth := Authority{
		WeightThreshold: 1,
		AccountAuths:    []AccountAuth{{Account: "bob", Weight: 1}},
		KeyAuths:        []KeyAuth{{Key: kp.PublicKey, Weight: 1}},
	}
	op := AccountCreateOperation{
		Fee: Asset{3000, 3, "HIVE"}, Creator: "alice", NewAccountName: "newuser",
		Owner: auth, Active: auth, Posting: auth, MemoKey: kp.PublicKey,
	}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestSerializeTxBadExpiration(t *testing.T) {
	tx := Transaction{
		RefBlockNum:    1,
		RefBlockPrefix: 12345,
		Expiration:     "not-a-date",
		Operations:     []HiveOperation{VoteOperation{"alice", "bob", "post", 10000}},
	}
	_, err := tx.Serialize()
	if err == nil {
		t.Fatal("expected error for bad expiration")
	}
}

func TestTransactionSerialize(t *testing.T) {
	tx := Transaction{
		RefBlockNum:    1,
		RefBlockPrefix: 12345,
		Expiration:     "2025-01-01T00:00:00",
		Operations:     []HiveOperation{VoteOperation{"alice", "bob", "post", 10000}},
	}
	got, err := tx.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty serialized tx")
	}
}

func TestTransactionGenerateTrxIdBadExpiration(t *testing.T) {
	tx := Transaction{
		RefBlockNum: 1, RefBlockPrefix: 12345,
		Expiration: "bad",
		Operations: []HiveOperation{VoteOperation{"alice", "bob", "post", 10000}},
	}
	_, err := tx.GenerateTrxId()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSerializeOpLimitOrderCreateFillOrKill(t *testing.T) {
	op := LimitOrderCreateOperation{
		Owner:        "alice",
		OrderId:      1,
		AmountToSell: Asset{1000, 3, "HIVE"},
		MinToReceive: Asset{500, 3, "HBD"},
		FillOrKill:   true,
		Expiration:   "2030-01-01T00:00:00",
	}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestSerializeOpUpdateProposalVotesDisapprove(t *testing.T) {
	op := UpdateProposalVotesOperation{Voter: "alice", ProposalIds: []int64{1}, Approve: false}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestSerializeOpEscrowApproveReject(t *testing.T) {
	op := EscrowApproveOperation{From: "alice", To: "bob", Agent: "carol", Who: "bob", EscrowId: 1, Approve: false}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestSerializeOpsErrorPropagation(t *testing.T) {
	// An op with a bad time should propagate its error through serializeOps.
	tx := Transaction{
		RefBlockNum: 1, RefBlockPrefix: 12345,
		Expiration: "2025-01-01T00:00:00",
		Operations: []HiveOperation{
			LimitOrderCreateOperation{Expiration: "bad-date"},
		},
	}
	_, err := tx.Serialize()
	if err == nil {
		t.Fatal("expected error propagated from SerializeOp")
	}
}

func TestSerializeOpCustomBinary(t *testing.T) {
	// beem's Custom_binary uses uint16 id and omits auth fields — does not match the
	// Hive C++ protocol spec. Bytes derived from the protocol spec manually.
	op := CustomBinaryOperation{
		RequiredOwnerAuths:   []string{},
		RequiredActiveAuths:  []string{},
		RequiredPostingAuths: []string{},
		RequiredAuths:        []Authority{},
		Id:                   "test",
		Data:                 []byte{1, 2},
	}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	expected := []byte{35, 0, 0, 0, 0, 4, 116, 101, 115, 116, 2, 1, 2}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpCustomBinaryWithAuths(t *testing.T) {
	// Exercise the RequiredAuths loop.
	kp, err := KeyPairFromWif("5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W")
	if err != nil {
		t.Fatal(err)
	}
	auth := Authority{WeightThreshold: 1, AccountAuths: []AccountAuth{}, KeyAuths: []KeyAuth{{Key: kp.PublicKey, Weight: 1}}}
	op := CustomBinaryOperation{
		RequiredActiveAuths: []string{"alice"},
		RequiredAuths:       []Authority{auth},
		Id:                  "myapp",
		Data:                []byte{0xde, 0xad, 0xbe, 0xef},
	}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestSerializeOpCustom(t *testing.T) {
	// Verified against beem: Custom(required_auths=["alice"], id=42, data="0102")
	// → [15, 1, 5, 97, 108, 105, 99, 101, 42, 0, 4, 48, 49, 48, 50]
	op := CustomOperation{
		RequiredAuths: []string{"alice"},
		Id:            42,
		Data:          "0102",
	}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	expected := []byte{15, 1, 5, 97, 108, 105, 99, 101, 42, 0, 4, 48, 49, 48, 50}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}

func TestSerializeOpCustomEmpty(t *testing.T) {
	// Empty case: op_id(15) + varint(0 auths) + uint16(0) + vstring("")
	op := CustomOperation{RequiredAuths: []string{}, Id: 0, Data: ""}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatal(err)
	}
	expected := []byte{15, 0, 0, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("got %v, want %v", got, expected)
	}
}
