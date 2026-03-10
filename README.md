# hivego

A Go client library for the [Hive](https://hive.io) blockchain.

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

## Broadcasting Operations

### Vote

```go
txid, err := client.Broadcast.Vote("voter", "author", "permlink", 10000, wif)
```

`weight` ranges from `-10000` (100% downvote) to `10000` (100% upvote).

### Transfer

```go
amount, _ := hivego.ParseAsset("1.000 HIVE")
txid, err := client.Broadcast.Transfer("sender", "receiver", amount, "memo", wif)
```

### Custom JSON

```go
// Active key action (required_auths)
txid, err := client.Broadcast.CustomJson([]string{"myaccount"}, []string{}, "app-id", `{"key":"value"}`, activeWif)

// Posting key action (required_posting_auths)
txid, err := client.Broadcast.CustomJson([]string{}, []string{"myaccount"}, "app-id", `{"key":"value"}`, postingWif)
```

### Claim Rewards

```go
hive,  _ := hivego.ParseAsset("0.000 HIVE")
hbd,   _ := hivego.ParseAsset("0.000 HBD")
vests, _ := hivego.ParseAsset("1234.567890 VESTS")
txid, err := client.Broadcast.ClaimRewards("myaccount", hive, hbd, vests, wif)
```

## Manual Transaction Building

For multi-operation transactions, offline signing, or custom operations:

```go
ops := []hivego.HiveOperation{
    hivego.VoteOperation{Voter: "voter", Author: "author", Permlink: "permlink", Weight: 10000},
    hivego.CustomJsonOperation{
        RequiredPostingAuths: []string{"voter"},
        Id:   "follow",
        Json: `["follow", {...}]`,
    },
}

tx, err := client.BuildTransaction(ops)
if err != nil { ... }

if err := client.Sign(tx, wif); err != nil { ... }

txid, _ := tx.GenerateTrxId()
if err := client.BroadcastRaw(tx); err != nil { ... }
```

### Custom Operations

Implement `HiveOperation` to broadcast operations not yet built into the library:

```go
type MyOperation struct {
    Account string `json:"account"`
}

func (o MyOperation) OpName() string              { return "my_operation_name" }
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

## Key Management

```go
// From WIF string
kp, err := hivego.KeyPairFromWif("5J...")

// From raw 32-byte private key
kp, err := hivego.KeyPairFromBytes(privKeyBytes)

// Public key string encoding/decoding
pubKeyStr := kp.GetPublicKeyString()              // "STM7..."
pubKey, err := hivego.DecodePublicKey("STM7...")

// Graphene base58 encode/decode
payload, version, err := hivego.GphBase58CheckDecode(wif)
encoded := hivego.GphBase58Encode(payload, version)
```

## Asset Handling

```go
a, err := hivego.ParseAsset("1.500 HIVE")
fmt.Println(a.Amount)    // 1500  (3 decimal places for HIVE/HBD, 6 for VESTS)
fmt.Println(a.Precision) // 3
fmt.Println(a.Symbol)    // "HIVE"
fmt.Println(a.String())  // "1.500 HIVE"
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
| Vote | `client.Broadcast.Vote(...)` | `hive.commit.vote(...)` | `client.broadcast.vote(...)` |
| Custom JSON | `client.Broadcast.CustomJson(...)` | `hive.commit.custom_json(...)` | `client.broadcast.customJson(...)` |
| Read block | `client.Database.GetBlock(...)` | `hive.get_block(...)` | `client.database.getBlock(...)` |
| Read account | `client.Database.GetAccounts(...)` | `Account(name)` | `client.database.getAccounts(...)` |
| Manual tx | `BuildTransaction` / `Sign` / `BroadcastRaw` | `TransactionBuilder` | `client.broadcast.sendOperations(...)` |
| Custom ops | implement `HiveOperation` | extend `Operation` | implement `Operation` |
| Block streaming | `client.Database.StreamBlocks(ctx, ...)` | `blockchain.stream_from(...)` | loop over `getBlock(...)` |