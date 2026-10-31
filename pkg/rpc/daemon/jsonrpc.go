package daemon

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	methodGenerateBlocks         = "generateblocks"
	methodGetAlternateChains     = "get_alternate_chains"
	methodGetBans                = "get_bans"
	methodGetBlock               = "get_block"
	methodGetBlockCount          = "get_block_count"
	methodGetBlockHeadersRange   = "get_block_headers_range"
	methodGetBlockHeaderByHash   = "get_block_header_by_hash"
	methodGetBlockHeaderByHeight = "get_block_header_by_height"
	methodGetBlockTemplate       = "get_block_template"
	methodGetCoinbaseTxSum       = "get_coinbase_tx_sum"
	methodGetConnections         = "get_connections"
	methodGetFeeEstimate         = "get_fee_estimate"
	methodGetInfo                = "get_info"
	methodGetLastBlockHeader     = "get_last_block_header"
	methodGetVersion             = "get_version"
	methodHardForkInfo           = "hard_fork_info"
	methodOnGetBlockHash         = "on_get_block_hash"
	methodRPCAccessTracking      = "rpc_access_tracking"
	methodRelayTx                = "relay_tx"
	methodSetBans                = "set_bans"
	methodSyncInfo               = "sync_info"
	methodAddAuxPow              = "add_aux_pow"
	methodBanned                 = "banned"
	methodCalcPow                = "calc_pow"
	methodFlushCache             = "flush_cache"
	methodFlushTxpool            = "flush_txpool"
	methodGetMinerData           = "get_miner_data"
	methodGetOutputDistribution  = "get_output_distribution"
	methodGetOutputHistogram     = "get_output_histogram"
	methodGetTxpoolBacklog       = "get_txpool_backlog"
	methodPruneBlockchain        = "prune_blockchain"
	methodSubmitBlock            = "submit_block"
)

// GetAlternateChains displays alternative chains seen by the node.
//
// (restricted).
func (c *Client) GetAlternateChains(
	ctx context.Context,
) (*GetAlternateChainsResult, error) {
	resp := &GetAlternateChainsResult{}

	err := c.JSONRPC(ctx, methodGetAlternateChains, nil, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// RPCAccessTracking retrieves statistics that the monero daemon keeps track of
// about the use of each RPC method and endpoint.
//
// (restricted).
func (c *Client) RPCAccessTracking(
	ctx context.Context,
) (*RPCAccessTrackingResult, error) {
	resp := &RPCAccessTrackingResult{}

	err := c.JSONRPC(ctx, methodRPCAccessTracking, nil, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// HardForkInfo looks up informaiton about the last hard fork.
func (c *Client) HardForkInfo(
	ctx context.Context,
) (*HardForkInfoResult, error) {
	resp := &HardForkInfoResult{}

	err := c.JSONRPC(ctx, methodHardForkInfo, nil, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetBans retrieves the list of banned IPs.
//
// (restricted).
func (c *Client) GetBans(ctx context.Context) (*GetBansResult, error) {
	resp := &GetBansResult{}

	err := c.JSONRPC(ctx, methodGetBans, nil, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

type SetBansBan struct {
	Host    string `json:"host"`
	Ban     bool   `json:"ban"`
	Seconds int64  `json:"seconds"`
}

type SetBansRequestParameters struct {
	Bans []SetBansBan `json:"bans"`
}

// SetBans bans a particular host.
//
// (restricted).
func (c *Client) SetBans(
	ctx context.Context, params SetBansRequestParameters,
) (*SetBansResult, error) {
	resp := &SetBansResult{}

	err := c.JSONRPC(ctx, methodSetBans, params, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetVersion retrieves the version of monerod that the node uses.
//
// (restricted).
func (c *Client) GetVersion(ctx context.Context) (*GetVersionResult, error) {
	resp := &GetVersionResult{}

	err := c.JSONRPC(ctx, methodGetVersion, nil, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GenerateBlocksRequestParameters is the set of parameters to be passed to the
// GenerateBlocks RPC method.
type GenerateBlocksRequestParameters struct {
	// AmountOfBlocks is the number of blocks to be generated.
	//
	AmountOfBlocks uint64 `json:"amount_of_blocks,omitempty"`

	// WalletAddress is the address of the wallet that will get the rewards
	// of the coinbase transaction for such the blocks generates.
	//
	WalletAddress string `json:"wallet_address,omitempty"`

	// PreviousBlock TODO
	//
	PreviousBlock string `json:"prev_block,omitempty"`

	// StartingNonce TODO
	//
	StartingNonce uint32 `json:"starting_nonce,omitempty"`
}

// GenerateBlocks combines functionality from `GetBlockTemplate` and
// `SubmitBlock` RPC calls to allow rapid block creation.
//
// Difficulty is set permanently to 1 for regtest.
//
// (restricted).
func (c *Client) GenerateBlocks(
	ctx context.Context, params GenerateBlocksRequestParameters,
) (*GenerateBlocksResult, error) {
	resp := &GenerateBlocksResult{}

	err := c.JSONRPC(ctx, methodGenerateBlocks, params, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

func (c *Client) GetBlockCount(
	ctx context.Context,
) (*GetBlockCountResult, error) {
	resp := &GetBlockCountResult{}

	err := c.JSONRPC(ctx, methodGetBlockCount, nil, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

func (c *Client) OnGetBlockHash(
	ctx context.Context, height uint64,
) (string, error) {
	resp := ""
	params := []uint64{height}

	err := c.JSONRPC(ctx, methodOnGetBlockHash, params, &resp)
	if err != nil {
		return "", fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

func (c *Client) RelayTx(
	ctx context.Context, txns []string,
) (*RelayTxResult, error) {
	resp := &RelayTxResult{}
	params := map[string]interface{}{
		"txids": txns,
	}

	err := c.JSONRPC(ctx, methodRelayTx, params, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetBlockTemplate gets a block template on which mining a new block.
func (c *Client) GetBlockTemplate(
	ctx context.Context, walletAddress string, reserveSize uint,
) (*GetBlockTemplateResult, error) {
	resp := &GetBlockTemplateResult{}
	params := map[string]interface{}{
		"wallet_address": walletAddress,
		"reserve_size":   reserveSize,
	}

	err := c.JSONRPC(ctx, methodGetBlockTemplate, params, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

func (c *Client) GetConnections(
	ctx context.Context,
) (*GetConnectionsResult, error) {
	resp := &GetConnectionsResult{}

	err := c.JSONRPC(ctx, methodGetConnections, nil, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetInfo retrieves general information about the state of the node and the
// network.
func (c *Client) GetInfo(ctx context.Context) (*GetInfoResult, error) {
	resp := &GetInfoResult{}

	if err := c.JSONRPC(ctx, methodGetInfo, nil, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

func (c *Client) GetLastBlockHeader(
	ctx context.Context,
) (*GetLastBlockHeaderResult, error) {
	resp := &GetLastBlockHeaderResult{}

	err := c.JSONRPC(ctx, methodGetLastBlockHeader, nil, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

func (c *Client) GetCoinbaseTxSum(
	ctx context.Context, height, count uint64,
) (*GetCoinbaseTxSumResult, error) {
	resp := &GetCoinbaseTxSumResult{}
	params := map[string]uint64{
		"height": height,
		"count":  count,
	}

	err := c.JSONRPC(ctx, methodGetCoinbaseTxSum, params, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// InnerJSON parses the content of the JSON embedded in `GetBlockResult`.
func (j *GetBlockResult) InnerJSON() (*GetBlockResultJSON, error) {
	res := &GetBlockResultJSON{}

	err := json.Unmarshal([]byte(j.JSON), res)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return res, nil
}

func (c *Client) GetBlockHeadersRange(
	ctx context.Context, start, end uint64,
) (*GetBlockHeadersRangeResult, error) {
	resp := &GetBlockHeadersRangeResult{}
	params := map[string]interface{}{
		"start_height": start,
		"end_height":   end,
	}

	err := c.JSONRPC(ctx, methodGetBlockHeadersRange, params, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetBlockHeaderByHeight retrieves block header information for either one or
// multiple blocks.
func (c *Client) GetBlockHeaderByHeight(
	ctx context.Context, height uint64,
) (*GetBlockHeaderByHeightResult, error) {
	resp := &GetBlockHeaderByHeightResult{}
	params := map[string]interface{}{
		"height": height,
	}

	err := c.JSONRPC(ctx, methodGetBlockHeaderByHeight, params, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetBlockHeaderByHash retrieves block header information for either one or
// multiple blocks.
func (c *Client) GetBlockHeaderByHash(
	ctx context.Context, hashes []string,
) (*GetBlockHeaderByHashResult, error) {
	resp := &GetBlockHeaderByHashResult{}
	params := map[string]interface{}{
		"hashes": hashes,
	}

	err := c.JSONRPC(ctx, methodGetBlockHeaderByHash, params, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetBlockRequestParameters represents the set of possible parameters that can
// be used for submitting a call to the `get_block` jsonrpc method.
type GetBlockRequestParameters struct {
	Height uint64 `json:"height,omitempty"`
	Hash   string `json:"hash,omitempty"`
}

// GetBlock fetches full block information from a block at a particular hash OR
// height.
func (c *Client) GetBlock(
	ctx context.Context, params GetBlockRequestParameters,
) (*GetBlockResult, error) {
	resp := &GetBlockResult{}

	err := c.JSONRPC(ctx, methodGetBlock, params, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

func (c *Client) GetFeeEstimate(
	ctx context.Context, graceBlocks uint64,
) (*GetFeeEstimateResult, error) {
	resp := &GetFeeEstimateResult{}
	params := map[string]uint64{
		"grace_blocks": graceBlocks,
	}

	err := c.JSONRPC(ctx, methodGetFeeEstimate, params, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

func (c *Client) SyncInfo(ctx context.Context) (*SyncInfoResult, error) {
	resp := &SyncInfoResult{}

	err := c.JSONRPC(ctx, methodSyncInfo, nil, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetTxpoolBacklog returns every transaction-pool entry as packed weight,
// fee, and time-in-pool records.
//
// On a restricted RPC port, monerod still serves this method and reports
// time_in_pool as 0.
func (c *Client) GetTxpoolBacklog(
	ctx context.Context,
) (*GetTxpoolBacklogResult, error) {
	resp := &GetTxpoolBacklogResult{}

	err := c.JSONRPC(ctx, methodGetTxpoolBacklog, nil, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetMinerData returns the template fields p2pool and other miners need:
// version, height, previous id, seed hash, difficulty, median weight, coins
// already generated, and the mineable transaction backlog.
func (c *Client) GetMinerData(ctx context.Context) (*GetMinerDataResult, error) {
	resp := &GetMinerDataResult{}

	err := c.JSONRPC(ctx, methodGetMinerData, nil, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// BannedRequestParameters is the input for Banned.
type BannedRequestParameters struct {
	// Address is the host or IP address to look up.
	Address string `json:"address"`
}

// Banned reports whether the daemon currently bans address, and for how
// many seconds the ban remains.
//
// (restricted).
func (c *Client) Banned(
	ctx context.Context, params BannedRequestParameters,
) (*BannedResult, error) {
	resp := &BannedResult{}

	err := c.JSONRPC(ctx, methodBanned, params, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// FlushTxpoolRequestParameters is the input for FlushTxpool.
type FlushTxpoolRequestParameters struct {
	// Txids lists transactions to drop. An empty list drops every
	// transaction in the pool.
	Txids []string `json:"txids"`
}

// FlushTxpool drops transactions from the daemon pool.
//
// This call changes daemon state. An empty Txids list removes the whole
// pool.
//
// (restricted).
func (c *Client) FlushTxpool(
	ctx context.Context, params FlushTxpoolRequestParameters,
) (*FlushTxpoolResult, error) {
	if params.Txids == nil {
		params.Txids = []string{}
	}

	resp := &FlushTxpoolResult{}

	err := c.JSONRPC(ctx, methodFlushTxpool, params, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetOutputHistogramRequestParameters is the input for GetOutputHistogram.
type GetOutputHistogramRequestParameters struct {
	Amounts      []uint64 `json:"amounts,omitempty"`
	MinCount     uint64   `json:"min_count,omitempty"`
	MaxCount     uint64   `json:"max_count,omitempty"`
	Unlocked     bool     `json:"unlocked,omitempty"`
	RecentCutoff uint64   `json:"recent_cutoff,omitempty"`
}

// GetOutputHistogram returns how many outputs exist for the requested
// amounts. RingCT outputs are amount 0.
func (c *Client) GetOutputHistogram(
	ctx context.Context, params GetOutputHistogramRequestParameters,
) (*GetOutputHistogramResult, error) {
	resp := &GetOutputHistogramResult{}

	err := c.JSONRPC(ctx, methodGetOutputHistogram, params, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetOutputDistributionRequestParameters is the input for
// GetOutputDistribution.
type GetOutputDistributionRequestParameters struct {
	// Amounts are the cleartext amounts to tally. 0 selects RingCT outputs.
	Amounts []uint64 `json:"amounts"`

	// Cumulative includes every earlier output of that amount in each count.
	Cumulative bool `json:"cumulative"`

	// FromHeight is the first block height to include.
	FromHeight uint64 `json:"from_height"`

	// ToHeight is the last block height to include. 0 means through the tip.
	ToHeight uint64 `json:"to_height"`

	// Binary asks for monerod's packed little-endian blob instead of a JSON
	// array. Monerod defaults this to true. The zero value sends false so
	// the response is a JSON array.
	Binary bool `json:"binary"`

	// Compress selects monerod's compressed integer encoding. Monerod
	// ignores it when Binary is false. Counts does not expand compressed
	// payloads.
	Compress bool `json:"compress"`
}

// GetOutputDistribution returns per-block output counts for the given
// amounts.
//
// On a restricted RPC port, monerod only accepts Amounts equal to []uint64{0}.
func (c *Client) GetOutputDistribution(
	ctx context.Context, params GetOutputDistributionRequestParameters,
) (*GetOutputDistributionResult, error) {
	if params.Amounts == nil {
		params.Amounts = []uint64{}
	}

	resp := &GetOutputDistributionResult{}

	err := c.JSONRPC(ctx, methodGetOutputDistribution, params, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// SubmitBlockRequest is the JSON-RPC params array of mined block blobs.
type SubmitBlockRequest []string

// SubmitBlock submits mined block blobs to the daemon.
//
// This call changes daemon and network state when a blob is accepted.
func (c *Client) SubmitBlock(
	ctx context.Context, blobs SubmitBlockRequest,
) (*SubmitBlockResult, error) {
	if blobs == nil {
		blobs = SubmitBlockRequest{}
	}

	resp := &SubmitBlockResult{}

	err := c.JSONRPC(ctx, methodSubmitBlock, blobs, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// PruneBlockchainRequestParameters is the input for PruneBlockchain.
type PruneBlockchainRequestParameters struct {
	// Check reports pruning status and does not delete data when true.
	// False prunes the blockchain and changes daemon state.
	Check bool `json:"check"`
}

// PruneBlockchain prunes historical transaction data, or reports whether
// the chain is already pruned.
//
// This call changes daemon state unless Check is true.
//
// (restricted).
func (c *Client) PruneBlockchain(
	ctx context.Context, params PruneBlockchainRequestParameters,
) (*PruneBlockchainResult, error) {
	resp := &PruneBlockchainResult{}

	err := c.JSONRPC(ctx, methodPruneBlockchain, params, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// CalcPowRequestParameters is the input for CalcPow.
type CalcPowRequestParameters struct {
	MajorVersion uint8  `json:"major_version"`
	Height       uint64 `json:"height"`
	BlockBlob    string `json:"block_blob"`
	SeedHash     string `json:"seed_hash"`
}

// CalcPow asks the daemon for the proof-of-work hash of a block candidate.
// The JSON-RPC result is the hex hash itself.
//
// (restricted).
func (c *Client) CalcPow(
	ctx context.Context, params CalcPowRequestParameters,
) (string, error) {
	var resp string

	err := c.JSONRPC(ctx, methodCalcPow, params, &resp)
	if err != nil {
		return "", fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// FlushCacheRequestParameters is the input for FlushCache.
type FlushCacheRequestParameters struct {
	// BadBlocks drops the invalid-block cache when true. False leaves that
	// cache in place. monerod 0.18 does not accept a separate bad-tx flag.
	BadBlocks bool `json:"bad_blocks"`
}

// FlushCache drops the daemon's cached invalid blocks when requested.
//
// This call changes daemon state when BadBlocks is true. When BadBlocks is
// false, monerod 0.18 leaves the cache unchanged and still returns status OK.
//
// (restricted).
func (c *Client) FlushCache(
	ctx context.Context, params FlushCacheRequestParameters,
) (*FlushCacheResult, error) {
	resp := &FlushCacheResult{}

	err := c.JSONRPC(ctx, methodFlushCache, params, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// AuxPow is one auxiliary chain's id and hash for merge mining.
type AuxPow struct {
	ID   string `json:"id"`
	Hash string `json:"hash"`
}

// AddAuxPowRequestParameters is the input for AddAuxPow.
type AddAuxPowRequestParameters struct {
	BlocktemplateBlob string   `json:"blocktemplate_blob"`
	AuxPow            []AuxPow `json:"aux_pow"`
}

// AddAuxPow folds auxiliary proof-of-work hashes into a block template so
// the template can be merge-mined. The daemon returns the updated blobs and
// the aux merkle root. Submitting the resulting block later is what changes
// chain state.
func (c *Client) AddAuxPow(
	ctx context.Context, params AddAuxPowRequestParameters,
) (*AddAuxPowResult, error) {
	if params.AuxPow == nil {
		params.AuxPow = []AuxPow{}
	}

	resp := &AddAuxPowResult{}

	err := c.JSONRPC(ctx, methodAddAuxPow, params, resp)
	if err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}
