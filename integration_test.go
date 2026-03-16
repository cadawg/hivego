//go:build integration

package hivego

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

const integrationNode = "https://api.deathwing.me"

func integrationClient() *Client {
	return NewClient(integrationNode)
}

func TestIntegrationGetBlock(t *testing.T) {
	client := integrationClient()
	block, err := client.Database.GetBlock(1)
	if err != nil {
		t.Fatalf("GetBlock(1): %v", err)
	}
	if block == nil {
		t.Fatal("GetBlock(1): got nil block")
	}
	if block.BlockID == "" {
		t.Error("block.BlockID is empty")
	}
	if block.Witness == "" {
		t.Error("block.Witness is empty")
	}
	if block.Timestamp == "" {
		t.Error("block.Timestamp is empty")
	}
	t.Logf("block 1: id=%s witness=%s ts=%s", block.BlockID, block.Witness, block.Timestamp)
}

func TestIntegrationGetBlockKnown(t *testing.T) {
	// Block 88,000,000 is a well-known block on Hive mainnet.
	client := integrationClient()
	block, err := client.Database.GetBlock(88_000_000)
	if err != nil {
		t.Fatalf("GetBlock(88000000): %v", err)
	}
	if block == nil {
		t.Fatal("GetBlock(88000000): got nil block")
	}
	t.Logf("block 88000000: id=%s witness=%s txns=%d", block.BlockID, block.Witness, len(block.Transactions))
}

func TestIntegrationGetAccounts(t *testing.T) {
	client := integrationClient()
	accounts, err := client.Database.GetAccounts([]string{"hive"})
	if err != nil {
		t.Fatalf("GetAccounts([hive]): %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(accounts))
	}
	acc := accounts[0]
	if acc.Name != "hive" {
		t.Errorf("expected name=hive, got %q", acc.Name)
	}
	if acc.MemoKey == "" {
		t.Error("memo_key is empty")
	}
	t.Logf("account hive: id=%d balance=%s vests=%s", acc.ID, acc.Balance, acc.VestingShares)
}

func TestIntegrationGetAccountsMultiple(t *testing.T) {
	client := integrationClient()
	names := []string{"hive", "steemit"}
	accounts, err := client.Database.GetAccounts(names)
	if err != nil {
		t.Fatalf("GetAccounts: %v", err)
	}
	if len(accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(accounts))
	}
	for _, acc := range accounts {
		if acc.Name == "" {
			t.Error("got account with empty name")
		}
		t.Logf("account %s: balance=%s", acc.Name, acc.Balance)
	}
}

func TestIntegrationGetAccountsNotFound(t *testing.T) {
	client := integrationClient()
	// Valid Hive account name that almost certainly does not exist (≤16 chars).
	accounts, err := client.Database.GetAccounts([]string{"zzznotexist"})
	if err != nil {
		t.Fatalf("GetAccounts: %v", err)
	}
	// The API returns an empty array for non-existent accounts.
	if len(accounts) != 0 {
		t.Errorf("expected 0 accounts for unknown name, got %d", len(accounts))
	}
}

func TestIntegrationParseAssetFromAccount(t *testing.T) {
	client := integrationClient()
	accounts, err := client.Database.GetAccounts([]string{"hive"})
	if err != nil {
		t.Fatalf("GetAccounts: %v", err)
	}
	if len(accounts) == 0 {
		t.Fatal("no accounts returned")
	}
	acc := accounts[0]
	for _, field := range []string{acc.Balance, acc.HbdBalance, acc.VestingShares} {
		if field == "" {
			continue
		}
		asset, err := ParseAsset(field)
		if err != nil {
			t.Errorf("ParseAsset(%q): %v", field, err)
			continue
		}
		if asset.Symbol == "" {
			t.Errorf("ParseAsset(%q): empty symbol", field)
		}
	}
}

func TestIntegrationGetDynamicGlobalProperties(t *testing.T) {
	client := integrationClient()
	raw, err := client.Database.GetDynamicGlobalProperties()
	if err != nil {
		t.Fatalf("GetDynamicGlobalProperties: %v", err)
	}
	var props map[string]json.RawMessage
	if err := json.Unmarshal(raw, &props); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"head_block_number", "head_block_id", "time", "current_witness"} {
		if _, ok := props[key]; !ok {
			t.Errorf("missing expected field %q in dynamic global properties", key)
		}
	}
	t.Logf("head_block_number=%s", props["head_block_number"])
}

func TestIntegrationBlockHasValidTransactionIDs(t *testing.T) {
	// Fetch a busy block and verify transaction IDs are valid hex strings.
	client := integrationClient()
	block, err := client.Database.GetBlock(50_000_000)
	if err != nil {
		t.Fatalf("GetBlock: %v", err)
	}
	if block == nil {
		t.Fatal("nil block")
	}
	if len(block.Transactions) != len(block.TransactionIDs) {
		t.Errorf("transactions count %d != transaction_ids count %d",
			len(block.Transactions), len(block.TransactionIDs))
	}
	for i, txid := range block.TransactionIDs {
		if len(txid) != 40 {
			t.Errorf("tx[%d] id %q: expected 40 hex chars, got %d", i, txid, len(txid))
		}
		if strings.ContainsAny(txid, "ghijklmnopqrstuvwxyz") {
			t.Errorf("tx[%d] id %q contains non-hex chars", i, txid)
		}
	}
	t.Logf("block 50000000: %d transactions", len(block.Transactions))
}

func TestIntegrationNoBroadcastSign(t *testing.T) {
	// Verify that WithNoBroadcast allows building and signing a transaction
	// without network broadcast. Uses a well-known test key.
	client := integrationClient().WithNoBroadcast()

	kp, err := KeyPairFromWif("5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W")
	if err != nil {
		t.Fatalf("KeyPairFromWif: %v", err)
	}

	op := TransferOperation{
		From:   "alice",
		To:     "bob",
		Amount: Asset{Amount: 1000, Precision: 3, Symbol: "HIVE"},
		Memo:   "integration test",
	}

	tx, txid, err := client.BroadcastOps([]HiveOperation{op}, kp)
	if err != nil {
		t.Fatalf("BroadcastOps (no-broadcast): %v", err)
	}
	if tx == nil {
		t.Fatal("tx is nil")
	}
	if txid == "" {
		t.Fatal("txid is empty")
	}
	if len(tx.Signatures) == 0 {
		t.Fatal("no signatures on transaction")
	}
	t.Logf("txid=%s sig=%s", txid, tx.Signatures[0])
}

// --- NoBroadcast tests for all BroadcastAPI convenience methods ---

// nbClient returns a no-broadcast client backed by api.deathwing.me (needs ref block).
func nbClient(t *testing.T) (*Client, *KeyPair) {
	t.Helper()
	kp, err := KeyPairFromWif("5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W")
	if err != nil {
		t.Fatalf("KeyPairFromWif: %v", err)
	}
	return integrationClient().WithNoBroadcast(), kp
}

func assertSigned(t *testing.T, tx *Transaction, txid string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("BroadcastOps: %v", err)
	}
	if tx == nil {
		t.Fatal("tx is nil")
	}
	if txid == "" {
		t.Fatal("txid is empty")
	}
	if len(tx.Signatures) == 0 {
		t.Fatal("no signatures")
	}
}

func nbAuth(kp *KeyPair) Authority {
	return Authority{
		WeightThreshold: 1,
		AccountAuths:    []AccountAuth{},
		KeyAuths:        []KeyAuth{{Key: kp.PublicKey, Weight: 1}},
	}
}

func TestIntegrationNoBroadcastVote(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.Vote("alice", "bob", "my-post", 10000, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastCustomJson(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.CustomJson(nil, []string{"alice"}, "follow", `{"what":["blog"]}`, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastTransfer(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.Transfer("alice", "bob", Asset{1000, 3, "HIVE"}, "memo", kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastClaimRewards(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.ClaimRewards("alice",
		Asset{100, 3, "HIVE"}, Asset{0, 3, "HBD"}, Asset{1000000, 6, "VESTS"}, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastComment(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.Comment("", "hive", "alice", "my-post", "Title", "Body", "{}", kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastCommentOptions(t *testing.T) {
	c, kp := nbClient(t)
	ops := []HiveOperation{
		CommentOperation{"", "hive", "alice", "my-post", "Title", "Body", "{}"},
		CommentOptionsOperation{
			Author:               "alice",
			Permlink:             "my-post",
			MaxAcceptedPayout:    Asset{1000000000, 3, "HBD"},
			PercentHbd:           10000,
			AllowVotes:           true,
			AllowCurationRewards: true,
			Beneficiaries:        []Beneficiary{{Account: "bob", Weight: 500}},
		},
	}
	tx, txid, err := c.BroadcastOps(ops, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastDeleteComment(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.DeleteComment("alice", "my-post", kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastPowerUp(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.PowerUp("alice", "alice", Asset{1000, 3, "HIVE"}, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastPowerDown(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.PowerDown("alice", Asset{1000000, 6, "VESTS"}, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastDelegate(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.Delegate("alice", "bob", Asset{1000000, 6, "VESTS"}, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastVoteWitness(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.VoteWitness("alice", "gtg", true, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastTransferToSavings(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.TransferToSavings("alice", "alice", Asset{1000, 3, "HIVE"}, "", kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastTransferFromSavings(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.TransferFromSavings("alice", 1, "alice", Asset{1000, 3, "HIVE"}, "", kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastSetWithdrawRoute(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.SetWithdrawRoute("alice", "bob", 5000, false, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastCancelTransferFromSavings(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.CancelTransferFromSavings("alice", 1, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastRecurrentTransfer(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.RecurrentTransfer("alice", "bob", Asset{1000, 3, "HIVE"}, "", 24, 2, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastUpdateAccount(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.UpdateAccount("alice", "{}", "{}", kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastAccountUpdate(t *testing.T) {
	// account_update (legacy) via BroadcastOps
	c, kp := nbClient(t)
	auth := nbAuth(kp)
	op := AccountUpdateOperation{
		Account:      "alice",
		Active:       &auth,
		MemoKey:      kp.PublicKey,
		JsonMetadata: "{}",
	}
	tx, txid, err := c.BroadcastOps([]HiveOperation{op}, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastAccountUpdate2WithAuthority(t *testing.T) {
	// account_update2 with an active authority change
	c, kp := nbClient(t)
	auth := nbAuth(kp)
	op := AccountUpdate2Operation{
		Account:             "alice",
		Active:              &auth,
		MemoKey:             kp.PublicKey,
		JsonMetadata:        "{}",
		PostingJsonMetadata: "{}",
	}
	tx, txid, err := c.BroadcastOps([]HiveOperation{op}, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastConvert(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.Convert("alice", 1, Asset{1000, 3, "HBD"}, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastCollateralizedConvert(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.CollateralizedConvert("alice", 1, Asset{1000, 3, "HIVE"}, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastLimitOrderCreate(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.LimitOrderCreate("alice", 1,
		Asset{1000, 3, "HIVE"}, Asset{500, 3, "HBD"}, false, "2030-01-01T00:00:00", kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastLimitOrderCancel(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.LimitOrderCancel("alice", 1, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastSetWitnessProxy(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.SetWitnessProxy("alice", "bob", kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastCreateProposal(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.CreateProposal("alice", "alice",
		"2025-01-01T00:00:00", "2025-12-31T00:00:00",
		Asset{1000, 3, "HBD"}, "My Proposal", "my-permlink", kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastUpdateProposalVotes(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.UpdateProposalVotes("alice", []int64{1, 2, 3}, true, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastRemoveProposal(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.RemoveProposal("alice", []int64{1}, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastUpdateProposal(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.UpdateProposal(1, "alice", Asset{500, 3, "HBD"}, "Updated", "my-permlink", "", kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastUpdateProposalWithEndDate(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.UpdateProposal(1, "alice", Asset{500, 3, "HBD"}, "Updated", "my-permlink", "2025-06-01T00:00:00", kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastClaimAccount(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.ClaimAccount("alice", Asset{0, 3, "HIVE"}, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastFeedPublish(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.FeedPublish("alice", Asset{1000, 3, "HBD"}, Asset{1000, 3, "HIVE"}, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastCreateAccount(t *testing.T) {
	c, kp := nbClient(t)
	auth := nbAuth(kp)
	tx, txid, err := c.Broadcast.CreateAccount(
		Asset{3000, 3, "HIVE"}, "alice", "newaccount",
		auth, auth, auth, kp.PublicKey, "{}", kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastCreateClaimedAccount(t *testing.T) {
	c, kp := nbClient(t)
	auth := nbAuth(kp)
	tx, txid, err := c.Broadcast.CreateClaimedAccount(
		"alice", "newaccount",
		auth, auth, auth, kp.PublicKey, "{}", kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastUpdateWitness(t *testing.T) {
	c, kp := nbClient(t)
	props := ChainProperties{
		AccountCreationFee: Asset{3000, 3, "HIVE"},
		MaximumBlockSize:   131072,
		HbdInterestRate:    1000,
	}
	tx, txid, err := c.Broadcast.UpdateWitness("alice", "https://example.com", kp.PublicKey, props, Asset{0, 3, "HIVE"}, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastWitnessSetProperties(t *testing.T) {
	c, kp := nbClient(t)
	// Encode the signing key as 33-byte compressed pubkey for the "key" property.
	keyBytes := kp.PublicKey.SerializeCompressed()
	// account_creation_fee as 16-byte LE asset: amount(8) + precision(1) + symbol(7)
	var feeBuf [16]byte
	binary.LittleEndian.PutUint64(feeBuf[0:], uint64(3000))
	feeBuf[8] = 3
	copy(feeBuf[9:], []byte("STEEM\x00\x00"))
	props := map[string][]byte{
		"key":                  keyBytes,
		"account_creation_fee": feeBuf[:],
	}
	tx, txid, err := c.Broadcast.WitnessSetProperties("alice", props, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastCancelDeclineVotingRights(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.CancelDeclineVotingRights("alice", kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastDeclineVotingRights(t *testing.T) {
	c, kp := nbClient(t)
	op := DeclineVotingRightsOperation{
		Account:                       "alice",
		Decline:                       true,
		IUnderstandThisIsIrreversible: true,
	}
	tx, txid, err := c.BroadcastOps([]HiveOperation{op}, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastRequestAccountRecovery(t *testing.T) {
	c, kp := nbClient(t)
	auth := nbAuth(kp)
	tx, txid, err := c.Broadcast.RequestAccountRecovery("steem", "alice", auth, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastRecoverAccount(t *testing.T) {
	c, kp := nbClient(t)
	auth := nbAuth(kp)
	tx, txid, err := c.Broadcast.RecoverAccount("alice", auth, auth, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastChangeRecoveryAccount(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.ChangeRecoveryAccount("alice", "bob", kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastEscrowTransfer(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.EscrowTransfer(
		"alice", "bob", "carol", 1,
		Asset{0, 3, "HBD"}, Asset{1000, 3, "HIVE"}, Asset{100, 3, "HIVE"},
		"2025-06-01T00:00:00", "2025-12-01T00:00:00", "{}", kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastEscrowApprove(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.EscrowApprove("alice", "bob", "carol", "bob", 1, true, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastEscrowDispute(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.EscrowDispute("alice", "bob", "carol", "alice", 1, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastEscrowRelease(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.EscrowRelease(
		"alice", "bob", "carol", "carol", "bob", 1,
		Asset{0, 3, "HBD"}, Asset{1000, 3, "HIVE"}, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationGetTransaction(t *testing.T) {
	// Use a known transaction from block 88,000,000.
	client := integrationClient()
	block, err := client.Database.GetBlock(88_000_000)
	if err != nil || block == nil || len(block.TransactionIDs) == 0 {
		t.Skip("could not fetch reference block or block has no transactions")
	}
	txid := block.TransactionIDs[0]
	raw, err := client.Database.GetTransaction(txid, true)
	if err != nil {
		t.Fatalf("GetTransaction(%s): %v", txid, err)
	}
	var result map[string]json.RawMessage
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := result["transaction_id"]; !ok {
		t.Error("response missing transaction_id field")
	}
	t.Logf("fetched transaction %s", txid)
}
func TestIntegrationRPCErrorAs(t *testing.T) {
	// Trigger an RPC error by requesting a non-existent method.
	// The node returns an error we can unwrap with errors.As.
	client := integrationClient()
	// GetAccounts with a name that's too long for the chain (>16 chars) triggers RPC error -32003.
	_, err := client.Database.GetAccounts([]string{"this-name-is-way-too-long-for-hive"})
	if err == nil {
		t.Fatal("expected RPC error, got nil")
	}
	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) {
		t.Fatalf("expected *RPCError via errors.As, got %T: %v", err, err)
	}
	if rpcErr.Code == 0 {
		t.Error("expected non-zero error code")
	}
	// Call Error() to cover the method.
	msg := rpcErr.Error()
	if msg == "" {
		t.Error("RPCError.Error() returned empty string")
	}
	t.Logf("RPCError: %s", msg)
}

func TestIntegrationFailover(t *testing.T) {
	// First node is invalid; client should failover to the second.
	client := NewClient("http://127.0.0.1:1", integrationNode)
	block, err := client.Database.GetBlock(1)
	if err != nil {
		t.Fatalf("expected failover to succeed, got: %v", err)
	}
	if block == nil || block.BlockID == "" {
		t.Fatal("got nil or empty block after failover")
	}
	t.Logf("failover succeeded: block id=%s", block.BlockID)
}

func TestIntegrationTransactionSerialize(t *testing.T) {
	// Build a tx via no-broadcast, then call Serialize() on it.
	c, kp := nbClient(t)
	tx, _, err := c.BroadcastOps([]HiveOperation{
		TransferOperation{"alice", "bob", Asset{1000, 3, "HIVE"}, ""},
	}, kp)
	if err != nil {
		t.Fatalf("BroadcastOps: %v", err)
	}
	b, err := tx.Serialize()
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if len(b) == 0 {
		t.Error("Serialize returned empty bytes")
	}
	t.Logf("serialized tx: %d bytes", len(b))
}

func TestIntegrationGetBlockRange(t *testing.T) {
	client := integrationClient()
	results, err := client.Database.GetBlockRange(1, 10)
	if err != nil {
		t.Fatalf("GetBlockRange: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results from GetBlockRange")
	}
	t.Logf("GetBlockRange returned %d batch results", len(results))
}

func TestIntegrationNoBroadcastCustomBinary(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.CustomBinary(
		nil, nil, nil, nil,
		"myapp", []byte{0xde, 0xad, 0xbe, 0xef}, kp)
	assertSigned(t, tx, txid, err)
}

func TestIntegrationNoBroadcastCustom(t *testing.T) {
	c, kp := nbClient(t)
	tx, txid, err := c.Broadcast.Custom([]string{"alice"}, 42, "0102", kp)
	assertSigned(t, tx, txid, err)
}
