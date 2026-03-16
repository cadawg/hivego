# hivego — Claude Context

## Project Overview

Go client library for the Hive blockchain (`github.com/cadawg/hivego`). Provides:

- Querying blockchain data (blocks, accounts, transactions)
- Signing and broadcasting transactions
- Binary serialization of Hive operations
- Key management (WIF import, public key encoding/decoding)

Single package (`hivego`), no subpackages. Targets Go 1.24+.

## Tech Stack

- **Language:** Go 1.24+
- **Crypto:** `github.com/decred/dcrd/dcrec/secp256k1/v4` — secp256k1 key generation and ECDSA compact signing
- **Base58:** `github.com/decred/base58` — WIF and public key encoding
- **RPC:** `github.com/cadawg/jsonrpc2client` — JSON-RPC 2.0 HTTP client with batching
- **Hashing:** `golang.org/x/crypto/ripemd160` — public key checksums (required for Hive compatibility)

## Development Commands

```bash
go test ./...        # Run all tests
go build ./...       # Build (no binary — library only)
go vet ./...         # Static analysis
```

## File Map

| File | Responsibility |
|---|---|
| `hrpcclient.go` | `Client` struct, `NewClient`, `WithNoBroadcast`, JSON-RPC execution and failover |
| `broadcaster.go` | `Transaction` struct, `BuildTransaction`, `Sign`, `BroadcastTx`, `BroadcastOps`, `Serialize` |
| `hive_ops.go` | All `HiveOperation` structs and `BroadcastAPI` convenience methods |
| `serializer.go` | Binary serialization — `SerializeOp()` implementations and helpers |
| `signer.go` | `SignDigest`, `GphBase58CheckDecode/Encode`, signing digest helpers |
| `keys.go` | `KeyPair`, `KeyPairFromWif`, `KeyPairFromBytes`, public key encode/decode |
| `errors.go` | Sentinel errors (`ErrNilKey` etc.) and `RPCError` struct |
| `types.go` | `Asset`, `ParseAsset`, `Block`, `AccountData`, `AuthorityData`, `Authority`, `AccountAuth`, `KeyAuth` |
| `account.go` | `DatabaseAPI` account methods |
| `blockApi.go` | `DatabaseAPI` block methods (`GetBlock`, `StreamBlocks`, `GetBlockRange`) |
| `testmocks_test.go` | Shared test fixtures (unexported, test-only) |
| `misc_test.go` | Unit tests for errors, types, keys, OpName, MarshalJSON |
| `integration_test.go` | Integration + no-broadcast tests (`//go:build integration`) |
| `scripts/coverage.sh` | Runs unit + integration coverage reports |

## Key Design Decisions

**Keys**: All broadcast operations accept `*KeyPair`, never raw WIF strings. Users call `KeyPairFromWif` once at the boundary. WIF strings must not appear past key construction.

**Errors**: All errors are sentinel `var Err... = errors.New(...)` values so callers can use `errors.Is`. Dynamic errors wrap a sentinel with `fmt.Errorf("context: %w", ErrSentinel)`. RPC errors use the `RPCError` struct for `errors.As` access to code/message.

**Broadcast return**: All `BroadcastAPI` methods and `BroadcastOps` return `(*Transaction, string, error)` — the signed transaction, the transaction ID, and any error. Callers who only need the txid use `_, txid, err`.

**Binary serialization**: Transactions are serialized locally to binary for signing. The JSON sent to the node is separate. The node re-serializes the JSON it receives and verifies signatures against that. Both sides must agree on the format exactly — a wrong byte in serialization produces a valid-looking transaction that the node rejects with a signature mismatch.

**Asset**: `ParseAsset` is always required — no string shortcuts. The binary serializer needs `Amount`, `Precision`, and `Symbol` as separate fields.

## Adding a New Operation

Every operation needs three things:

**1. Struct + `OpName()` in `hive_ops.go`:**
```go
type MyOperation struct {
    Account string `json:"account"`
    // ...
}
func (o MyOperation) OpName() string { return "my_operation" }
```

**2. `SerializeOp()` in `serializer.go`:**
```go
func (o MyOperation) SerializeOp() ([]byte, error) {
    var buf bytes.Buffer
    buf.WriteByte(opIdB("my_operation"))
    appendVString(o.Account, &buf)
    // strings  → appendVString
    // uint16   → binary.LittleEndian.PutUint16 into 2-byte buf
    // uint32   → binary.LittleEndian.PutUint32 into 4-byte buf
    // Asset    → appendAssetBytes
    // bool     → WriteByte(1) or WriteByte(0)
    // optional → WriteByte(0) if absent, WriteByte(1) + value if present
    // []string → appendVStringArray
    // no extensions → buf.WriteByte(0)
    return buf.Bytes(), nil
}
```

**3. Op ID in `getHiveOpIds()` in `hive_ops.go`** — verify the ID matches the Hive chain source.

**4. Convenience method on `BroadcastAPI` (optional):**
```go
func (b BroadcastAPI) MyOp(account string, key *KeyPair) (*Transaction, string, error) {
    op := MyOperation{account}
    return b.client.BroadcastOps([]HiveOperation{op}, key)
}
```

Operations that must be combined with another op (e.g. `comment_options` with `comment`) should **not** get a convenience method — callers use `BroadcastOps` directly.

## Serialization Helpers

| Helper | Use for |
|---|---|
| `appendVString(s, buf)` | Variable-length prefixed string |
| `appendVStringArray(a, buf)` | Length-prefixed array of strings (e.g. `required_auths`) |
| `appendVBytes(data, buf)` | Variable-length prefixed raw bytes (e.g. witness property values) |
| `appendAssetBytes(a, buf)` | 8-byte LE amount + 1-byte precision + 7-byte null-padded symbol |
| `appendAuthorityBytes(a, buf)` | uint32 threshold + varint count + account_auths + varint count + key_auths |
| `appendChainPropertiesBytes(p, buf)` | asset(account_creation_fee) + uint32(max_block_size) + uint16(hbd_interest_rate) |
| `appendInt64Array(ids, buf)` | Varint count + 8-byte LE each (for proposal IDs) |
| `opIdB("op_name")` | First byte of every serialized op |

## Testing

Tests are in `_test.go` files in the same package. `testmocks_test.go` provides shared fixtures. Tests verify binary serialization output against known-good byte sequences — when adding ops, add a serialization test with a hardcoded expected output verified against another Hive library or the reference implementation.

- **Unit tests**: `go test ./...`
- **Integration tests** (require network): `go test -tags integration ./...` — gated behind `//go:build integration`, use `api.deathwing.me`; all broadcast ops tested via `WithNoBroadcast()`
- **Coverage**: `bash scripts/coverage.sh` — runs both suites and prints totals
- **Test vectors**: `testdata/generate_vectors.py` — generates and verifies serialization byte sequences using beem as the reference implementation

## Transaction Flow

```
BuildTransaction(ops)        → fetches ref block + expiration from chain
Sign(tx, key)                → serializeTx → hashTxForSig → SignCompact → appends hex sig
BroadcastTx(tx)              → prepareJson (builds OperationsJs) → condenser_api.broadcast_transaction
```

`BroadcastOps` wraps all three. `NoBroadcast` skips the final step.

## Configuration

**Mainnet (default):**
```go
client := hivego.NewClient("https://api.hive.blog")
// ChainID defaults to HiveMainnetChainID
// PublicKeyPrefix defaults to "STM"
```

**Testnet:**
```go
client := hivego.NewClient("https://testnet.openhive.network")
client.ChainID = "18dcf0a285365fc58b71f18b3d3fec954aa0c141c44e4e5cb4cf777b9eab274e"
client.PublicKeyPrefix = "TST"
```

**Dry-run mode:**
```go
client := hivego.NewClient("https://api.hive.blog").WithNoBroadcast()
```

## Hive-Specific Notes

- **Chain ID**: Mainnet is `beeab0de00000000...` (in `HiveMainnetChainID`). Testnet requires a different chain ID and `PublicKeyPrefix = "TST"`.
- **Public key checksum**: ripemd160, not SHA256 — this is intentional and required for Hive compatibility.
- **WIF checksum**: double-SHA256, same as Bitcoin.
- **Asset precision**: HIVE/HBD = 3 decimal places, VESTS = 6.
- **comment_options** must be in the same transaction as the comment it targets — the node rejects it otherwise.
- **set_withdraw_vesting_route** percent is basis points (10000 = 100%).
- **recurrent_transfer** minimum executions is 2; cancel by setting amount to 0.
- **Asset wire symbols**: Despite the Hive rebrand, the binary wire format (`condenser_api` / `pack_type::legacy`) still requires the original Steem symbol names: `HIVE` → `"STEEM"`, `HBD` → `"SBD"`. This is handled by `appendAssetBytes` in serializer.go. All other symbols (`VESTS`, custom) are written as-is. See `OBSOLETE_SYMBOL_SER` in `hive/libraries/protocol/include/hive/protocol/asset_symbol.hpp`. A post-HF26 NAI format exists but is not used here.
- **account_update2** optional fields (`Owner`, `Active`, `Posting *Authority`, `MemoKey *secp256k1.PublicKey`) serialize as `0x00` when nil, `0x01 + value` when set. `AuthorityData` in types.go is the read-side type (API responses); `Authority` is the write-side type (operations).
- **Authority binary format**: uint32 weight_threshold + varint count + account_auths (vstring + uint16 each) + varint count + key_auths (33-byte compressed pubkey + uint16 each). Handled by `appendAuthorityBytes` in serializer.go.
- **limit_order_create** expiration is a `"2006-01-02T15:04:05"` string, serialized as uint32 unix timestamp via `expTimeB`.
- **update_proposal** `EndDate` is encoded in the extensions array as a `static_variant` with type_id=1 (`update_proposal_end_date`, second type after `void_t`). When `""`: write `0x00` (empty extensions count). When set: write `0x01` (1 extension) + `0x01` (type_id) + uint32 timestamp. `DailyPay` can only be lowered — the chain enforces `o.daily_pay <= proposal.daily_pay`.
- **claim_account** fee of `"0.000 HIVE"` claims via RC; non-zero fee burns HIVE directly.
- **Proposal IDs** are `[]int64` serialized via `appendInt64Array` (varint count + 8-byte LE each).
- **witness_set_properties** `Props map[string][]byte` serializes as a sorted flat_map: varint count, then for each key (sorted): `vstring(key) + vbytes(value)`. JSON output uses `[["key","hexvalue"],...]` pairs via `MarshalJSON`. Property values are raw binary — callers build them with `encoding/binary`.
- **witness_update** nil `BlockSigningKey` serializes as 33 zero bytes (disables the witness). Same pattern for other required-but-disableable public key fields.
- **account_create / create_claimed_account** require all three authority structs and a memo key. Use `appendAuthorityBytes` for each authority.
- **account_update (legacy)** optional authorities serialize with `fc::optional` prefix byte (0x00 absent, 0x01 present). MemoKey is required (not optional).
- **account_recovery ops** use `appendAuthorityBytes` for authority fields. `recover_account` must be signed with both new and a recent old owner key — callers use `BuildTransaction` + `Sign` twice + `BroadcastTx`.
- **escrow ops** use standard `expTimeB` for deadline/expiration fields.
- **decline_voting_rights** `SerializeOp` returns `ErrDeclineVotingRightsNotConfirmed` if `Decline == true` and `IUnderstandThisIsIrreversible == false`.
- **custom_operation** (id 15): `RequiredAuths []string` (flat_set, vstring array) + `Id uint16` (LE) + `Data string` (vstring). Verified against beem.
- **custom_binary_operation** (id 35): four auth arrays (`required_owner_auths`, `required_active_auths`, `required_posting_auths` as vstring arrays, `required_auths` as `vector<authority>`) + `id string` (vstring) + `data []byte` (vbytes). **beem's implementation is wrong** — it uses `uint16` for id and omits all auth fields. Wire format follows the Hive C++ protocol spec. Bytes in test are protocol-spec derived, not beem-verified.