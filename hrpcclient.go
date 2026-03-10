package hivego

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"strconv"

	"github.com/cadawg/jsonrpc2client"
)

// HiveMainnetChainID is the chain ID for Hive mainnet.
const HiveMainnetChainID = "beeab0de00000000000000000000000000000000000000000000000000000000"

// Client is the Hive blockchain RPC client.
//
//	client := hivego.NewClient("https://api.hive.blog", "https://rpc.ecency.com")
//	txid, err  := client.Broadcast.Vote(...)
//	block, err := client.Database.GetBlock(...)
type Client struct {
	// Broadcast provides methods for submitting operations to the blockchain.
	Broadcast BroadcastAPI
	// Database provides methods for reading data from the blockchain.
	Database DatabaseAPI

	nodes           []string
	MaxConn         int
	MaxBatch        int
	NoBroadcast     bool
	ChainID         string // hex-encoded chain ID; defaults to HiveMainnetChainID
	PublicKeyPrefix string // public key prefix; defaults to "STM"
}

// BroadcastAPI provides methods for submitting signed operations to the Hive blockchain.
// Access via client.Broadcast.
type BroadcastAPI struct{ client *Client }

// DatabaseAPI provides methods for reading data from the Hive blockchain.
// Access via client.Database.
type DatabaseAPI struct{ client *Client }

// HiveRpcNode is an alias for Client for backward compatibility.
type HiveRpcNode = Client

// NewClient creates a Client connecting to one or more API node addresses.
// Nodes are tried in order on failure, providing automatic failover.
//
//	client := hivego.NewClient("https://api.hive.blog", "https://rpc.ecency.com")
func NewClient(nodes ...string) *Client {
	if len(nodes) == 0 {
		panic("hivego: at least one node address required")
	}
	c := &Client{
		nodes:           nodes,
		MaxConn:         1,
		MaxBatch:        4,
		ChainID:         HiveMainnetChainID,
		PublicKeyPrefix: "STM",
	}
	c.Broadcast = BroadcastAPI{c}
	c.Database = DatabaseAPI{c}
	return c
}

func (h *Client) chainIDBytes() []byte {
	id, _ := hex.DecodeString(h.ChainID)
	return id
}

func (h *Client) rpcExecWithFailover(query hrpcQuery) ([]byte, error) {
	var lastErr error
	for _, node := range h.nodes {
		result, err := h.rpcExec(node, query)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func (h *Client) rpcExec(endpoint string, query hrpcQuery) ([]byte, error) {
	rpcClient := jsonrpc2client.NewClientWithOpts(endpoint, h.MaxConn, h.MaxBatch)
	jr2query := &jsonrpc2client.RpcRequest{Method: query.method, JsonRpc: "2.0", Id: 1, Params: query.params}
	resp, err := rpcClient.CallRaw(jr2query)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, errors.New(strconv.Itoa(resp.Error.Code) + "    " + resp.Error.Message)
	}
	return resp.Result, nil
}

func (h *Client) rpcExecBatch(endpoint string, queries []hrpcQuery) ([]json.RawMessage, error) {
	rpcClient := jsonrpc2client.NewClientWithOpts(endpoint, h.MaxConn, h.MaxBatch)

	var jr2queries jsonrpc2client.RPCRequests
	for i, query := range queries {
		jr2query := &jsonrpc2client.RpcRequest{Method: query.method, JsonRpc: "2.0", Id: i, Params: query.params}
		jr2queries = append(jr2queries, jr2query)
	}

	resps, err := rpcClient.CallBatchRaw(jr2queries)
	if err != nil {
		return nil, err
	}

	var batchResult []json.RawMessage
	for _, resp := range resps {
		thisResult := json.RawMessage{}
		if err := json.Unmarshal(resp.Result, &thisResult); err != nil {
			log.Println("err unmarshalling res.result")
			log.Println(err)
			log.Println(resp)
		}
		batchResult = append(batchResult, thisResult)
	}
	return batchResult, nil
}

func (h *Client) rpcExecBatchFast(endpoint string, queries []hrpcQuery) ([][]byte, error) {
	rpcClient := jsonrpc2client.NewClientWithOpts(endpoint, h.MaxConn, h.MaxBatch)

	var jr2queries jsonrpc2client.RPCRequests
	for i, query := range queries {
		jr2query := &jsonrpc2client.RpcRequest{Method: query.method, JsonRpc: "2.0", Id: i, Params: query.params}
		jr2queries = append(jr2queries, jr2query)
	}

	resps, err := rpcClient.CallBatchFast(jr2queries)
	if err != nil {
		return nil, err
	}

	var batchResult [][]byte
	batchResult = append(batchResult, resps...)
	return batchResult, nil
}

type globalProps struct {
	HeadBlockNumber int    `json:"head_block_number"`
	HeadBlockId     string `json:"head_block_id"`
	Time            string `json:"time"`
}

type hrpcQuery struct {
	method string
	params interface{}
}
