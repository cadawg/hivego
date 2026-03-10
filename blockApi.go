package hivego

import (
	"context"
	"encoding/json"
	"time"
)

type getBlockRangeQueryParams struct {
	StartingBlockNum int `json:"starting_block_num"`
	Count            int `json:"count"`
}

// GetDynamicGlobalProperties returns the current dynamic global properties as raw JSON.
func (d DatabaseAPI) GetDynamicGlobalProperties() ([]byte, error) {
	q := hrpcQuery{method: "condenser_api.get_dynamic_global_properties", params: []string{}}
	return d.client.rpcExecWithFailover(q)
}

// GetBlock fetches a single typed Block by block number.
func (d DatabaseAPI) GetBlock(blockNum int) (*Block, error) {
	type blockParams struct {
		BlockNum int `json:"block_num"`
	}
	q := hrpcQuery{method: "block_api.get_block", params: blockParams{BlockNum: blockNum}}
	res, err := d.client.rpcExecWithFailover(q)
	if err != nil {
		return nil, err
	}

	type blockResponse struct {
		Block *Block `json:"block"`
	}
	var resp blockResponse
	if err := json.Unmarshal(res, &resp); err != nil {
		return nil, err
	}
	return resp.Block, nil
}

// GetBlockRange fetches a range of blocks starting at startBlock, returning raw JSON.
// Results are batched in groups of 500. For typed results use GetBlock or StreamBlocks.
func (d DatabaseAPI) GetBlockRange(startBlock int, count int) ([]json.RawMessage, error) {
	if d.client.MaxConn == 0 {
		d.client.MaxConn = 10
	}
	if d.client.MaxBatch == 0 {
		d.client.MaxBatch = 4
	}

	var queries []hrpcQuery
	for i := startBlock; i <= startBlock+count; {
		params := getBlockRangeQueryParams{StartingBlockNum: i, Count: 500}
		queries = append(queries, hrpcQuery{method: "block_api.get_block_range", params: params})
		i += 500
	}

	return d.client.rpcExecBatch(d.client.nodes[0], queries)
}

// GetBlockRangeFast fetches a range of blocks starting at startBlock, returning raw bytes.
// Results are batched in groups of 500.
func (d DatabaseAPI) GetBlockRangeFast(startBlock int, count int) ([][]byte, error) {
	if d.client.MaxConn == 0 {
		d.client.MaxConn = 10
	}
	if d.client.MaxBatch == 0 {
		d.client.MaxBatch = 4
	}

	var queries []hrpcQuery
	for i := startBlock; i <= startBlock+count; {
		params := getBlockRangeQueryParams{StartingBlockNum: i, Count: 500}
		queries = append(queries, hrpcQuery{method: "block_api.get_block_range", params: params})
		i += 500
	}

	return d.client.rpcExecBatchFast(d.client.nodes[0], queries)
}

// StreamBlocks streams blocks sequentially starting at startBlock.
// Blocks are delivered on the returned channel. When caught up to the head block,
// StreamBlocks polls at the given interval until the next block is produced.
// Cancel ctx to stop the stream; both returned channels are closed on exit.
// Any terminal error is sent on the error channel before it closes.
//
// Example:
//
//	blocks, errc := client.Database.StreamBlocks(ctx, 88000000, 3*time.Second)
//	for block := range blocks {
//	    fmt.Println(block.BlockID, block.Timestamp)
//	}
//	if err := <-errc; err != nil {
//	    log.Fatal(err)
//	}
func (d DatabaseAPI) StreamBlocks(ctx context.Context, startBlock int, pollInterval time.Duration) (<-chan Block, <-chan error) {
	blocks := make(chan Block, 10)
	errc := make(chan error, 1)

	go func() {
		defer close(blocks)
		defer close(errc)

		current := startBlock
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			block, err := d.GetBlock(current)
			if err != nil || block == nil {
				select {
				case <-ctx.Done():
					return
				case <-time.After(pollInterval):
				}
				continue
			}

			select {
			case blocks <- *block:
				current++
			case <-ctx.Done():
				return
			}
		}
	}()

	return blocks, errc
}
