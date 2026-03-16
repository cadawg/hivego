# hivego

A Go client library for the [Hive](https://hive.io) blockchain. 80%+ test coverage across unit and integration suites.

## Installation

```sh
go get github.com/cadawg/hivego
```

## Quick Start

```go
import "github.com/cadawg/hivego"

// Single node
client := hivego.NewClient("https://api.hive.blog")

// Multiple nodes — automatically fails over to the next on error
client := hivego.NewClient(
    "https://api.hive.blog",
    "https://rpc.ecency.com",
    "https://api.deathwing.me",
)
```

The client exposes two namespaces:

- **`client.Broadcast`** — submit signed operations to the blockchain
- **`client.Database`** — read data from the blockchain

## Keys

All broadcast operations take a `*KeyPair`. Create one from a WIF string once and reuse it:

```go
key, err := hivego.KeyPairFromWif("5J...")
if err != nil {
    log.Fatal(err)
}
```

You can also create one from raw bytes:

```go
key, err := hivego.KeyPairFromBytes(privKeyBytes)
```

## Broadcasting Operations

All broadcast methods return `(*Transaction, string, error)` — the signed transaction object,
the transaction ID, and any error. If you only need the txid, use `_` for the transaction:

```go
_, txid, err := client.Broadcast.Vote(...)
```

### Vote

```go
_, txid, err := client.Broadcast.Vote("voter", "author", "permlink", 10000, key)
```

`weight` ranges from `-10000` (100% downvote) to `10000` (100% upvote).

### Transfer

```go
amount, _ := hivego.ParseAsset("1.000 HIVE")
_, txid, err := client.Broadcast.Transfer("sender", "receiver", amount, "memo", key)
```

### Custom JSON

```go
// Active key action (required_auths)
_, txid, err := client.Broadcast.CustomJson([]string{"myaccount"}, []string{}, "app-id", `{"key":"value"}`, activeKey)

// Posting key action (required_posting_auths)
_, txid, err := client.Broadcast.CustomJson([]string{}, []string{"myaccount"}, "app-id", `{"key":"value"}`, postingKey)
```

### Claim Rewards

```go
hive,  _ := hivego.ParseAsset("0.000 HIVE")
hbd,   _ := hivego.ParseAsset("0.000 HBD")
vests, _ := hivego.ParseAsset("1234.567890 VESTS")
_, txid, err := client.Broadcast.ClaimRewards("myaccount", hive, hbd, vests, key)
```

### Post / Comment

```go
// Top-level post (parentAuthor = "")
_, txid, err := client.Broadcast.Comment(
    "", "hive-blog",       // parentAuthor, parentPermlink (category for top-level posts)
    "alice", "my-post",   // author, permlink
    "Hello World",        // title
    "This is my post.",   // body (markdown)
    `{"tags":["blog"]}`,  // json_metadata
    postingKey,
)

// Reply to a post
_, txid, err := client.Broadcast.Comment(
    "bob", "bobs-post",   // parentAuthor, parentPermlink
    "alice", "re-bobs-post",
    "", "Great post!",    // title is empty for replies
    "{}",
    postingKey,
)

// Delete a post (must have no replies, no payout, no net votes)
_, txid, err := client.Broadcast.DeleteComment("alice", "my-post", postingKey)
```

### Post with Beneficiaries

`comment_options` must be in the same transaction as the `comment` it targets. Use `BroadcastOps`:

```go
maxPayout, _ := hivego.ParseAsset("1000000.000 HBD") // no limit

ops := []hivego.HiveOperation{
    hivego.CommentOperation{
        Author: "alice", Permlink: "my-post",
        ParentPermlink: "hive-blog", Title: "Hello", Body: "...",
        JsonMetadata: `{"tags":["blog"]}`,
    },
    hivego.CommentOptionsOperation{
        Author: "alice", Permlink: "my-post",
        MaxAcceptedPayout:    maxPayout,
        PercentHbd:           10000, // 100% HBD rewards
        AllowVotes:           true,
        AllowCurationRewards: true,
        Beneficiaries: []hivego.Beneficiary{
            {Account: "appdev", Weight: 500}, // 5%
        },
    },
}

_, txid, err := client.BroadcastOps(ops, postingKey)
```

### Hive Power

```go
// Power up (HIVE → HP)
amount, _ := hivego.ParseAsset("100.000 HIVE")
_, txid, err := client.Broadcast.PowerUp("alice", "alice", amount, activeKey)

// Power down (HP → HIVE over 13 weeks)
vests, _ := hivego.ParseAsset("100000.000000 VESTS")
_, txid, err := client.Broadcast.PowerDown("alice", vests, activeKey)

// Cancel power-down
zero, _ := hivego.ParseAsset("0.000000 VESTS")
_, txid, err := client.Broadcast.PowerDown("alice", zero, activeKey)

// Delegate HP to another account
vests, _ := hivego.ParseAsset("50000.000000 VESTS")
_, txid, err := client.Broadcast.Delegate("alice", "bob", vests, activeKey)

// Remove delegation
zero, _ := hivego.ParseAsset("0.000000 VESTS")
_, txid, err := client.Broadcast.Delegate("alice", "bob", zero, activeKey)
```

### Witnesses

```go
// Vote for a witness
_, txid, err := client.Broadcast.VoteWitness("alice", "good-witness", true, activeKey)

// Remove witness vote
_, txid, err := client.Broadcast.VoteWitness("alice", "good-witness", false, activeKey)
```

### Savings

```go
// Move funds to savings (3-day withdrawal delay)
amount, _ := hivego.ParseAsset("100.000 HBD")
_, txid, err := client.Broadcast.TransferToSavings("alice", "alice", amount, "", activeKey)

// Initiate savings withdrawal (requestId must be unique per account)
_, txid, err := client.Broadcast.TransferFromSavings("alice", 1, "alice", amount, "", activeKey)

// Cancel a pending savings withdrawal
_, txid, err := client.Broadcast.CancelTransferFromSavings("alice", 1, activeKey)
```

### Vesting Routes

```go
// Route 50% of power-down to another account, auto-powering it up there
_, txid, err := client.Broadcast.SetWithdrawRoute("alice", "bob", 5000, true, activeKey)

// Remove a route (set percent to 0)
_, txid, err := client.Broadcast.SetWithdrawRoute("alice", "bob", 0, false, activeKey)
```

### Recurrent Transfers

```go
// Send 10 HIVE every 24 hours, 7 times
amount, _ := hivego.ParseAsset("10.000 HIVE")
_, txid, err := client.Broadcast.RecurrentTransfer("alice", "bob", amount, "weekly payment", 24, 7, activeKey)

// Cancel a recurrent transfer (set amount to 0)
zero, _ := hivego.ParseAsset("0.000 HIVE")
_, txid, err := client.Broadcast.RecurrentTransfer("alice", "bob", zero, "", 24, 2, activeKey)
```

### Account Update

```go
// Update profile and posting metadata
_, txid, err := client.Broadcast.UpdateAccount(
    "alice",
    `{"profile":{"name":"Alice","about":"Hive user"}}`,
    `{"tags":["hive","blog"]}`,
    postingKey,
)
```

For authority or memo key changes, construct `AccountUpdate2Operation` directly and use `BroadcastOps`.
All authority fields and `MemoKey` are optional — set to `nil` to leave unchanged:

```go
newKey, _ := hivego.KeyPairFromWif("5J...")

// Rotate the active key
_, txid, err := client.BroadcastOps([]hivego.HiveOperation{
    hivego.AccountUpdate2Operation{
        Account: "alice",
        Active: &hivego.Authority{
            WeightThreshold: 1,
            KeyAuths: []hivego.KeyAuth{
                {Key: newKey.PublicKey, Weight: 1},
            },
        },
    },
}, activeKey)

// Add a co-signer to the posting authority (2-of-2 multisig)
_, txid, err := client.BroadcastOps([]hivego.HiveOperation{
    hivego.AccountUpdate2Operation{
        Account: "alice",
        Posting: &hivego.Authority{
            WeightThreshold: 2,
            KeyAuths: []hivego.KeyAuth{
                {Key: alicePostingKey.PublicKey, Weight: 1},
                {Key: bobPostingKey.PublicKey, Weight: 1},
            },
        },
    },
}, activeKey)
```

### Conversions

```go
// Convert HBD to HIVE (3.5-day process, requestId must be unique per account)
hbd, _ := hivego.ParseAsset("100.000 HBD")
_, txid, err := client.Broadcast.Convert("alice", 1, hbd, activeKey)

// Convert HIVE to HBD instantly using collateral
hive, _ := hivego.ParseAsset("100.000 HIVE")
_, txid, err := client.Broadcast.CollateralizedConvert("alice", 1, hive, activeKey)
```

### Limit Orders

```go
// Place a limit order: sell 10 HIVE for at least 9 HBD, expires in 1 hour
sell, _ := hivego.ParseAsset("10.000 HIVE")
buy,  _ := hivego.ParseAsset("9.000 HBD")
_, txid, err := client.Broadcast.LimitOrderCreate("alice", 1, sell, buy, false, "2026-01-01T12:00:00", activeKey)

// Cancel an open order
_, txid, err := client.Broadcast.LimitOrderCancel("alice", 1, activeKey)
```

### Witness Proxy

```go
// Delegate all witness votes to a proxy
_, txid, err := client.Broadcast.SetWitnessProxy("alice", "my-proxy", activeKey)

// Remove proxy
_, txid, err := client.Broadcast.SetWitnessProxy("alice", "", activeKey)
```

### DHF Proposals

```go
// Create a funding proposal (dailyPay must be HBD)
dailyPay, _ := hivego.ParseAsset("100.000 HBD")
_, txid, err := client.Broadcast.CreateProposal(
    "alice", "alice",
    "2026-01-01T00:00:00", "2026-06-01T00:00:00",
    dailyPay,
    "My Proposal", "my-proposal-post",
    activeKey,
)

// Vote for proposals
_, txid, err := client.Broadcast.UpdateProposalVotes("alice", []int64{0, 1, 2}, true, activeKey)

// Remove votes
_, txid, err := client.Broadcast.UpdateProposalVotes("alice", []int64{0}, false, activeKey)

// Update a proposal (pass "" for endDate to leave it unchanged)
_, txid, err := client.Broadcast.UpdateProposal(0, "alice", dailyPay, "Updated Subject", "my-proposal-post", "", activeKey)

// Remove a proposal (creator only)
_, txid, err := client.Broadcast.RemoveProposal("alice", []int64{0}, activeKey)
```

### Account Creation

```go
// Claim an account creation token using RC (free, but costs RC)
zero, _ := hivego.ParseAsset("0.000 HIVE")
_, txid, err := client.Broadcast.ClaimAccount("alice", zero, activeKey)

// Claim using a HIVE fee instead
fee, _ := hivego.ParseAsset("3.000 HIVE")
_, txid, err := client.Broadcast.ClaimAccount("alice", fee, activeKey)

// Create a new account with a HIVE fee
fee, _ = hivego.ParseAsset("3.000 HIVE")
ownerKey, _ := hivego.KeyPairFromWif("5J...")
owner := hivego.Authority{WeightThreshold: 1, KeyAuths: []hivego.KeyAuth{{Key: ownerKey.PublicKey, Weight: 1}}}
_, txid, err = client.Broadcast.CreateAccount(fee, "alice", "newaccount", owner, owner, owner, ownerKey.PublicKey, "{}", activeKey)

// Create a new account from a claimed token (no fee)
_, txid, err = client.Broadcast.CreateClaimedAccount("alice", "newaccount", owner, owner, owner, ownerKey.PublicKey, "{}", activeKey)
```

### Witness Management

```go
// Register or update a witness
fee, _ := hivego.ParseAsset("0.000 HIVE")
props := hivego.ChainProperties{
    AccountCreationFee: func() hivego.Asset { a, _ := hivego.ParseAsset("3.000 HIVE"); return a }(),
    MaximumBlockSize:   131072,
    HbdInterestRate:    1000, // 10%
}
signingKey, _ := hivego.KeyPairFromWif("5J...")
_, txid, err := client.Broadcast.UpdateWitness("alice", "https://mywitness.com", signingKey.PublicKey, props, fee, activeKey)

// Disable a witness (set signing key to nil → 33 zero bytes)
_, txid, err = client.Broadcast.UpdateWitness("alice", "https://mywitness.com", nil, props, fee, activeKey)

// Modern witness_set_properties (preferred for active witnesses)
import "encoding/binary"
keyBuf := make([]byte, 2)
binary.LittleEndian.PutUint16(keyBuf, 1000) // hbd_interest_rate = 10%
_, txid, err = client.Broadcast.WitnessSetProperties("alice", map[string][]byte{
    "hbd_interest_rate": keyBuf,
    "key": signingKey.PublicKey.SerializeCompressed(),
}, activeKey)
```

### Publish Price Feed

```go
// Witnesses only: publish a HIVE/HBD price feed
base, _  := hivego.ParseAsset("1.000 HBD")
quote, _ := hivego.ParseAsset("3.500 HIVE")
_, txid, err := client.Broadcast.FeedPublish("mywitness", base, quote, witnessKey)
```

### Account Recovery

```go
// Step 1: recovery account submits the new owner authority
newOwner := hivego.Authority{WeightThreshold: 1, KeyAuths: []hivego.KeyAuth{{Key: newKey.PublicKey, Weight: 1}}}
_, txid, err := client.Broadcast.RequestAccountRecovery("recovery-account", "lost-account", newOwner, recoveryKey)

// Step 2: lost account owner completes recovery (sign with BOTH new and a recent old owner key)
recentOwner := hivego.Authority{WeightThreshold: 1, KeyAuths: []hivego.KeyAuth{{Key: oldKey.PublicKey, Weight: 1}}}
tx, err := client.BuildTransaction([]hivego.HiveOperation{
    hivego.RecoverAccountOperation{"lost-account", newOwner, recentOwner},
})
client.Sign(tx, newKey)
client.Sign(tx, oldKey)
client.BroadcastTx(tx)

// Change recovery account (takes 30 days to take effect)
_, txid, err = client.Broadcast.ChangeRecoveryAccount("alice", "new-recovery-acct", ownerKey)
```

### Escrow

```go
// Step 1: lock funds with an agent
hbd,  _ := hivego.ParseAsset("100.000 HBD")
zero, _ := hivego.ParseAsset("0.000 HIVE")
fee,  _ := hivego.ParseAsset("1.000 HBD")
_, txid, err := client.Broadcast.EscrowTransfer(
    "alice", "bob", "agent",
    1,                    // escrow ID
    hbd, zero, fee,
    "2026-04-01T00:00:00", // ratification deadline
    "2026-06-01T00:00:00", // escrow expiration
    `{}`,
    activeKey,
)

// Step 2a: to-party or agent approves
_, txid, err = client.Broadcast.EscrowApprove("alice", "bob", "agent", "bob", 1, true, bobKey)

// Step 2b: raise a dispute (hands release authority to the agent)
_, txid, err = client.Broadcast.EscrowDispute("alice", "bob", "agent", "bob", 1, bobKey)

// Step 3: release funds to bob
_, txid, err = client.Broadcast.EscrowRelease("alice", "bob", "agent", "agent", "bob", 1, hbd, zero, agentKey)
```

### Custom

`custom_operation` broadcasts an integer-keyed operation with arbitrary string data, authorized by active keys.

```go
_, txid, err := client.Broadcast.Custom([]string{"myaccount"}, 42, "my data", activeKey)
```

### Custom Binary

`custom_binary_operation` broadcasts raw binary data with full authority control (owner, active, posting, and authority objects). The `Id` field is a short string (< 32 chars); `Data` is arbitrary bytes.

```go
// Simple case — no authority restrictions
_, txid, err := client.Broadcast.CustomBinary(nil, nil, nil, nil, "myapp", []byte{0x01, 0x02}, activeKey)

// With active-key restriction
_, txid, err = client.Broadcast.CustomBinary(
    nil,                    // requiredOwnerAuths
    []string{"myaccount"}, // requiredActiveAuths
    nil,                    // requiredPostingAuths
    nil,                    // requiredAuths (Authority objects)
    "myapp", []byte{0xde, 0xad, 0xbe, 0xef},
    activeKey,
)
```

### Decline Voting Rights

Declining voting rights is **permanent and irreversible**. There is no convenience method for this
path — you must construct the operation directly to make the intent unambiguous:

```go
// PERMANENT — the account can never vote again
_, txid, err := client.BroadcastOps([]hivego.HiveOperation{
    hivego.DeclineVotingRightsOperation{
        Account: "alice",
        Decline: true,
        IUnderstandThisIsIrreversible: true,
    },
}, ownerKey)

// Cancel a pending (not yet effective) decline request
_, txid, err = client.Broadcast.CancelDeclineVotingRights("alice", ownerKey)
```

## Multi-Operation Transactions

To combine multiple operations in one transaction, use `BroadcastOps` directly:

```go
key, _ := hivego.KeyPairFromWif("5J...")

ops := []hivego.HiveOperation{
    hivego.CommentOperation{
        Author: "alice", Permlink: "my-post",
        ParentPermlink: "hive-blog", Title: "Hello", Body: "...",
        JsonMetadata: `{"tags":["blog"]}`,
    },
    hivego.CommentOptionsOperation{
        Author: "alice", Permlink: "my-post",
        MaxAcceptedPayout:    maxPayout,
        PercentHbd:           10000,
        AllowVotes:           true,
        AllowCurationRewards: true,
    },
}

tx, txid, err := client.BroadcastOps(ops, key)
```

### Advanced: Offline Signing and Multi-sig

For offline signing, multi-sig, or other workflows where you need control over each step,
use `BuildTransaction`, `Sign`, and `BroadcastTx` separately:

```go
tx, err := client.BuildTransaction(ops)
if err != nil { ... }

// sign with multiple keys for multi-sig operations
if err := client.Sign(tx, activeKey); err != nil { ... }
if err := client.Sign(tx, postingKey); err != nil { ... }

txid, _ := tx.GenerateTrxId()
if err := client.BroadcastTx(tx); err != nil { ... }
```

### Implementing Your Own Operation Type

Implement `HiveOperation` to broadcast operations not yet built into the library:

```go
type MyOperation struct {
    Account string `json:"account"`
}

func (o MyOperation) OpName() string               { return "my_operation_name" }
func (o MyOperation) SerializeOp() ([]byte, error) { /* binary serialization */ }
```

## Reading Chain Data

### Get a Block

```go
block, err := client.Database.GetBlock(88000000)
fmt.Println(block.Timestamp, block.Witness)
// block.Transactions is []json.RawMessage — unmarshal individual transactions as needed
```

### Stream Blocks

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

blocks, errc := client.Database.StreamBlocks(ctx, 88000000, 3*time.Second)
for block := range blocks {
    fmt.Println(block.BlockID, block.Timestamp)
}
if err := <-errc; err != nil {
    log.Fatal(err)
}
```

### Get Accounts

```go
accounts, err := client.Database.GetAccounts([]string{"alice", "bob"})
for _, acc := range accounts {
    fmt.Println(acc.Name, acc.Balance)
    bal, _ := hivego.ParseAsset(acc.Balance)
    fmt.Println(bal) // "1.500 HIVE"
}
```

### Get a Transaction

```go
raw, err := client.Database.GetTransaction("abc123...", true)
```

### Raw Block Range (bulk processing)

```go
// Returns raw bytes — fastest option for high-throughput block processing
rawBlocks, err := client.Database.GetBlockRangeFast(startBlock, count)

// Returns parsed json.RawMessage
blocks, err := client.Database.GetBlockRange(startBlock, count)
```

> **Warning:** Do not run bulk block fetching against public API nodes at high rates. Run your own node for block processing workloads.

## Asset Handling

Assets are parsed explicitly rather than accepted as raw strings. This is required because
transactions are binary-serialized locally before signing, and the serialization needs
the amount, precision, and symbol as separate fields.

```go
a, err := hivego.ParseAsset("1.500 HIVE")
fmt.Println(a.Amount)    // 1500  (3 decimal places for HIVE/HBD, 6 for VESTS)
fmt.Println(a.Precision) // 3
fmt.Println(a.Symbol)    // "HIVE"
fmt.Println(a.String())  // "1.500 HIVE"
```

## Error Handling

All errors are sentinel values so you can use `errors.Is` and `errors.As`:

```go
key, err := hivego.KeyPairFromWif(wif)
if errors.Is(err, hivego.ErrInvalidFormat) {
    // bad WIF string
}
if errors.Is(err, hivego.ErrChecksumMismatch) {
    // WIF decoded but checksum failed
}

_, err = hivego.ParseAsset("bad input")
if errors.Is(err, hivego.ErrInvalidAsset) {
    // not a valid asset string
}

// RPC errors expose the node's error code and message
_, err = client.Database.GetBlock(0)
var rpcErr *hivego.RPCError
if errors.As(err, &rpcErr) {
    fmt.Println(rpcErr.Code, rpcErr.Message)
}
```

Available sentinels: `ErrNilKey`, `ErrInvalidKeyLength`, `ErrInvalidPrefix`, `ErrInvalidPublicKey`, `ErrChecksumMismatch`, `ErrInvalidFormat`, `ErrInvalidAsset`.

## Public Key Utilities

```go
// Encode/decode public keys
pubKeyStr := key.GetPublicKeyString()              // "STM7..."
pubKey, err := hivego.DecodePublicKey("STM7...")

// Graphene base58 encode/decode (for advanced use)
payload, version, err := hivego.GphBase58CheckDecode(wif)
encoded := hivego.GphBase58Encode(payload, version)
```

## Testnet / Custom Chain

```go
client := hivego.NewClient("https://testnet.openhive.network")
client.ChainID = "18dcf0a285365fc58b71f18b3d3fec954aa0c141c44e4e5cb4cf777b9eab274e"
client.PublicKeyPrefix = "TST"

// Decode testnet public keys
pubKey, err := hivego.DecodePublicKeyWithPrefix("TST7...", "TST")
str := hivego.GetPublicKeyStringWithPrefix(pubKey, "TST")
```

## Dry-run / Inspecting Transactions

`WithNoBroadcast` enables dry-run mode: transactions are built and signed but not submitted.
The returned `*Transaction` can be inspected, logged, or serialized:

```go
client := hivego.NewClient("https://api.hive.blog").WithNoBroadcast()

tx, txid, err := client.Broadcast.Vote("voter", "author", "permlink", 10000, key)

// predicted transaction ID
fmt.Println(txid)

// signatures
fmt.Println(tx.Signatures)

// binary wire representation (what the node verifies the signature against)
b, err := tx.Serialize()
fmt.Printf("%x\n", b)

// JSON representation (what is sent to the node)
json.NewEncoder(os.Stdout).Encode(tx)
```

## Client Options

All fields are public and can be set after construction:

```go
client := hivego.NewClient("https://api.hive.blog")
client.MaxConn     = 4     // max concurrent HTTP connections per node (default 1)
client.MaxBatch    = 8     // max requests per batch RPC call (default 4)
client.NoBroadcast = true  // dry-run mode: sign but don't submit transactions
```

## Comparison with Other Libraries

| Concept | hivego | Beem (Python) | dhive (TypeScript) |
|---|---|---|---|
| Client | `Client` | `Hive` | `Client` |
| Multi-node failover | `NewClient("node1", "node2", ...)` | `NodeList` | `new Client(["node1", "node2"])` |
| Keys | `KeyPairFromWif("5J...")` | WIF string | `PrivateKey.fromString("5J...")` |
| Vote | `client.Broadcast.Vote(..., key)` | `hive.commit.vote(...)` | `client.broadcast.vote(...)` |
| Custom JSON | `client.Broadcast.CustomJson(..., key)` | `hive.commit.custom_json(...)` | `client.broadcast.customJson(...)` |
| Custom | `client.Broadcast.Custom(..., key)` | `hive.commit.custom(...)` | — |
| Custom Binary | `client.Broadcast.CustomBinary(..., key)` | `hive.commit.custom_binary(...)` | — |
| Read block | `client.Database.GetBlock(...)` | `hive.get_block(...)` | `client.database.getBlock(...)` |
| Read account | `client.Database.GetAccounts(...)` | `Account(name)` | `client.database.getAccounts(...)` |
| Multi-op tx | `BroadcastOps(ops, key)` | `TransactionBuilder` | `client.broadcast.sendOperations(...)` |
| Custom ops | implement `HiveOperation` | extend `Operation` | implement `Operation` |
| Block streaming | `client.Database.StreamBlocks(ctx, ...)` | `blockchain.stream_from(...)` | loop over `getBlock(...)` |