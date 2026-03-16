// Package hivego is a Go client library for the Hive blockchain.
//
// It provides:
//   - Signing and broadcasting transactions (vote, transfer, custom_json, and 40+ other operations)
//   - Reading chain data (blocks, accounts, dynamic global properties)
//   - Binary serialization of Hive operations for local signing
//   - Key management (WIF import, secp256k1 key pairs, public key encoding/decoding)
//
// # Quick start
//
//	client := hivego.NewClient("https://api.hive.blog")
//
//	key, err := hivego.KeyPairFromWif("5J...")
//
//	_, txid, err := client.Broadcast.Vote("voter", "author", "permlink", 10000, key)
//
//	block, err := client.Database.GetBlock(88000000)
//
// # Multiple nodes and failover
//
// Pass multiple node addresses to NewClient. Requests are tried against each node in order
// and fall through to the next on error:
//
//	client := hivego.NewClient(
//	    "https://api.hive.blog",
//	    "https://rpc.ecency.com",
//	    "https://api.deathwing.me",
//	)
//
// # Broadcasting operations
//
// All broadcast methods on [BroadcastAPI] return (*[Transaction], string, error) — the signed
// transaction, the predicted transaction ID, and any error. Use _ to discard the transaction
// when you only need the txid:
//
//	_, txid, err := client.Broadcast.Transfer("alice", "bob", amount, "memo", key)
//
// To combine multiple operations in one transaction, use [Client.BroadcastOps] directly:
//
//	ops := []hivego.HiveOperation{commentOp, commentOptionsOp}
//	_, txid, err := client.BroadcastOps(ops, postingKey)
//
// # Dry-run mode
//
// [Client.WithNoBroadcast] enables dry-run mode: transactions are built and signed but not
// submitted to the network. Useful for testing, fee estimation, and offline signing:
//
//	client := hivego.NewClient("https://api.hive.blog").WithNoBroadcast()
//	tx, txid, err := client.Broadcast.Vote(...)
//	b, _ := tx.Serialize() // inspect the binary wire format
//
// # Assets
//
// Use [ParseAsset] to create [Asset] values — the binary serializer requires the amount,
// precision, and symbol as separate fields, so raw strings are not accepted:
//
//	amount, err := hivego.ParseAsset("1.000 HIVE")
//
// # Testnet
//
//	client := hivego.NewClient("https://testnet.openhive.network")
//	client.ChainID = "18dcf0a285365fc58b71f18b3d3fec954aa0c141c44e4e5cb4cf777b9eab274e"
//	client.PublicKeyPrefix = "TST"
//
// # Error handling
//
// All errors are sentinel values — use [errors.Is] for specific conditions and [errors.As]
// to extract structured RPC errors:
//
//	var rpcErr *hivego.RPCError
//	if errors.As(err, &rpcErr) {
//	    fmt.Println(rpcErr.Code, rpcErr.Message)
//	}
package hivego
