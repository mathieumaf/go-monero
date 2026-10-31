package daemon

import (
	"context"
	"fmt"
)

// Binary endpoints (*.bin) are not implemented. They use epee portable
// storage rather than JSON and are a separate follow-up.

const (
	endpointGetAltBlocksHashes       = "/get_alt_blocks_hashes"
	endpointGetTransactionPoolHashes = "/get_transaction_pool_hashes"
	endpointInPeers                  = "/in_peers"
	endpointIsKeyImageSpent          = "/is_key_image_spent"
	endpointOutPeers                 = "/out_peers"
	endpointPopBlocks                = "/pop_blocks"
	endpointSaveBC                   = "/save_bc"
	endpointSendRawTransaction       = "/send_raw_transaction"
	endpointSetBootstrapDaemon       = "/set_bootstrap_daemon"
	endpointSetLogHashRate           = "/set_log_hash_rate"
	endpointStopDaemon               = "/stop_daemon"
	endpointUpdate                   = "/update"
)

// Update commands accepted by monerod's /update endpoint.
const (
	// UpdateCommandCheck asks whether a newer monerod release exists.
	UpdateCommandCheck = "check"
	// UpdateCommandDownload fetches that release onto the daemon host.
	// State-changing.
	UpdateCommandDownload = "download"
)

// Key image spent statuses returned by /is_key_image_spent. Values match
// monerod's COMMAND_RPC_IS_KEY_IMAGE_SPENT::STATUS.
const (
	KeyImageUnspent           KeyImageSpentStatus = 0
	KeyImageSpentInBlockchain KeyImageSpentStatus = 1
	KeyImageSpentInPool       KeyImageSpentStatus = 2
)

// KeyImageSpentStatus is one entry of /is_key_image_spent's spent_status.
type KeyImageSpentStatus uint64

// InPeersRequestParameters is the request for /in_peers.
//
// Set is always encoded. Monerod treats a missing set as true and would
// apply InPeers, including a zero value that disables incoming peers.
type InPeersRequestParameters struct {
	// Set applies InPeers when true. When false, the current limit is
	// returned and left unchanged.
	Set bool `json:"set"`
	// InPeers is the maximum number of incoming public peers. Applied
	// only when Set is true. Zero disables incoming public peers.
	InPeers uint32 `json:"in_peers"`
}

// InPeersResult is the result of a call to /in_peers.
type InPeersResult struct {
	InPeers uint32 `json:"in_peers"`

	RPCResultFooter `json:",inline"`
}

// OutPeersRequestParameters is the request for /out_peers.
//
// Set is always encoded. Monerod treats a missing set as true and would
// apply OutPeers, including a zero value that disables outgoing peers.
type OutPeersRequestParameters struct {
	// Set applies OutPeers when true. When false, the current limit is
	// returned and left unchanged.
	Set bool `json:"set"`
	// OutPeers is the maximum number of outgoing public peers. Applied
	// only when Set is true. Zero disables outgoing public peers.
	OutPeers uint32 `json:"out_peers"`
}

// OutPeersResult is the result of a call to /out_peers.
type OutPeersResult struct {
	OutPeers uint32 `json:"out_peers"`

	RPCResultFooter `json:",inline"`
}

// UpdateRequestParameters is the request for /update.
type UpdateRequestParameters struct {
	// Command is UpdateCommandCheck or UpdateCommandDownload.
	Command string `json:"command"`
	// Path is where a download is written. Used when Command is
	// UpdateCommandDownload.
	Path string `json:"path,omitempty"`
}

// UpdateResult is the result of a call to /update.
type UpdateResult struct {
	// Update reports whether a newer release is available.
	Update bool `json:"update"`
	// Version is the release monerod can download.
	Version string `json:"version"`
	// UserURI is the download URI presented to an operator.
	UserURI string `json:"user_uri"`
	// AutoURI is the download URI monerod uses itself.
	AutoURI string `json:"auto_uri"`
	// Hash is the expected digest of the downloaded archive.
	Hash string `json:"hash"`
	// Path is where the archive was written, when a download ran.
	Path string `json:"path"`

	RPCResultFooter `json:",inline"`
}

// GetTransactionPoolHashesResult is the result of
// /get_transaction_pool_hashes.
type GetTransactionPoolHashesResult struct {
	// TxHashes are the mempool transaction ids, hex-encoded.
	TxHashes []string `json:"tx_hashes"`

	RPCResultFooter `json:",inline"`
}

// GetAltBlocksHashesResult is the result of /get_alt_blocks_hashes.
type GetAltBlocksHashesResult struct {
	// BlksHashes are hashes of blocks that are not on the main chain.
	BlksHashes []string `json:"blks_hashes"`

	RPCResultFooter `json:",inline"`
}

// IsKeyImageSpentRequestParameters is the request for /is_key_image_spent.
type IsKeyImageSpentRequestParameters struct {
	// KeyImages are hex-encoded key images to check.
	KeyImages []string `json:"key_images"`
}

// IsKeyImageSpentResult is the result of /is_key_image_spent.
// SpentStatus aligns with KeyImages: 0 unspent, 1 spent on chain, 2 spent
// in the pool.
type IsKeyImageSpentResult struct {
	SpentStatus []KeyImageSpentStatus `json:"spent_status"`

	RPCResultFooter `json:",inline"`
}

// SendRawTransactionRequestParameters is the request for
// /send_raw_transaction.
type SendRawTransactionRequestParameters struct {
	// TxAsHex is the full transaction blob, hex-encoded.
	TxAsHex string `json:"tx_as_hex"`
	// DoNotRelay accepts the transaction without broadcasting it.
	DoNotRelay bool `json:"do_not_relay,omitempty"`
	// DoSanityChecks runs tx_sanity_check before the daemon accepts the
	// transaction. Nil omits the field so monerod keeps its default
	// (true). Set it to false to skip those checks.
	DoSanityChecks *bool `json:"do_sanity_checks,omitempty"`
}

// SendRawTransactionResult is the result of /send_raw_transaction.
type SendRawTransactionResult struct {
	// Reason is extra detail. Monerod leaves it empty, or sets it to
	// "Not relayed" when the transaction was accepted but not broadcast.
	Reason            string `json:"reason"`
	NotRelayed        bool   `json:"not_relayed"`
	LowMixin          bool   `json:"low_mixin"`
	DoubleSpend       bool   `json:"double_spend"`
	InvalidInput      bool   `json:"invalid_input"`
	InvalidOutput     bool   `json:"invalid_output"`
	TooBig            bool   `json:"too_big"`
	Overspend         bool   `json:"overspend"`
	FeeTooLow         bool   `json:"fee_too_low"`
	TooFewOutputs     bool   `json:"too_few_outputs"`
	SanityCheckFailed bool   `json:"sanity_check_failed"`
	TxExtraTooBig     bool   `json:"tx_extra_too_big"`
	NonzeroUnlockTime bool   `json:"nonzero_unlock_time"`

	RPCResultFooter `json:",inline"`
}

// SaveBCResult is the result of /save_bc.
type SaveBCResult struct {
	RPCResultFooter `json:",inline"`
}

// StopDaemonResult is the result of /stop_daemon.
type StopDaemonResult struct {
	RPCResultFooter `json:",inline"`
}

// SetBootstrapDaemonRequestParameters is the request for
// /set_bootstrap_daemon.
type SetBootstrapDaemonRequestParameters struct {
	// Address is the bootstrap daemon host:port. Empty clears it.
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
	// Proxy is an optional proxy URL used to reach Address.
	Proxy string `json:"proxy"`
}

// SetBootstrapDaemonResult is the result of /set_bootstrap_daemon.
// Monerod returns status and does not include the usual untrusted flag.
type SetBootstrapDaemonResult struct {
	Status string `json:"status"`
}

// PopBlocksRequestParameters is the request for /pop_blocks.
type PopBlocksRequestParameters struct {
	// Nblocks is how many blocks to remove from the tip.
	Nblocks uint64 `json:"nblocks"`
	// KeepTxs returns non-coinbase transactions from popped blocks to
	// the mempool when true. Nil omits the field so monerod keeps its
	// default (true). Set it to false to drop those transactions.
	// Monerod 0.18.4 and earlier ignore this field.
	KeepTxs *bool `json:"keep_txs,omitempty"`
}

// PopBlocksResult is the result of /pop_blocks.
type PopBlocksResult struct {
	// Height is the chain height after the pop.
	Height uint64 `json:"height"`

	RPCResultFooter `json:",inline"`
}

// SetLogHashRateRequestParameters is the request for /set_log_hash_rate.
type SetLogHashRateRequestParameters struct {
	// Visible shows hash-rate lines in the daemon log when true.
	Visible bool `json:"visible"`
}

// SetLogHashRateResult is the result of /set_log_hash_rate.
// Status is "OK", or "NOT MINING" when the daemon is not mining.
type SetLogHashRateResult struct {
	RPCResultFooter `json:",inline"`
}

// InPeers reads or sets the maximum number of incoming public peers.
//
// State-changing when params.Set is true.
func (c *Client) InPeers(
	ctx context.Context, params InPeersRequestParameters,
) (*InPeersResult, error) {
	resp := &InPeersResult{}

	err := c.RawRequest(ctx, endpointInPeers, params, resp)
	if err != nil {
		return nil, fmt.Errorf("raw request: %w", err)
	}

	return resp, nil
}

// OutPeers reads or sets the maximum number of outgoing public peers.
//
// State-changing when params.Set is true.
func (c *Client) OutPeers(
	ctx context.Context, params OutPeersRequestParameters,
) (*OutPeersResult, error) {
	resp := &OutPeersResult{}

	err := c.RawRequest(ctx, endpointOutPeers, params, resp)
	if err != nil {
		return nil, fmt.Errorf("raw request: %w", err)
	}

	return resp, nil
}

// Update checks for a monerod release or downloads one.
//
// State-changing when params.Command is UpdateCommandDownload: the
// daemon writes the archive onto its own host.
func (c *Client) Update(
	ctx context.Context, params UpdateRequestParameters,
) (*UpdateResult, error) {
	resp := &UpdateResult{}

	err := c.RawRequest(ctx, endpointUpdate, params, resp)
	if err != nil {
		return nil, fmt.Errorf("raw request: %w", err)
	}

	return resp, nil
}

// GetTransactionPoolHashes returns mempool transaction ids without the
// transaction bodies.
func (c *Client) GetTransactionPoolHashes(
	ctx context.Context,
) (*GetTransactionPoolHashesResult, error) {
	resp := &GetTransactionPoolHashesResult{}

	err := c.RawRequest(ctx, endpointGetTransactionPoolHashes, nil, resp)
	if err != nil {
		return nil, fmt.Errorf("raw request: %w", err)
	}

	return resp, nil
}

// GetAltBlocksHashes returns hashes of blocks that are not on the main
// chain.
func (c *Client) GetAltBlocksHashes(
	ctx context.Context,
) (*GetAltBlocksHashesResult, error) {
	resp := &GetAltBlocksHashesResult{}

	err := c.RawRequest(ctx, endpointGetAltBlocksHashes, nil, resp)
	if err != nil {
		return nil, fmt.Errorf("raw request: %w", err)
	}

	return resp, nil
}

// IsKeyImageSpent reports whether each key image is unspent, spent on
// chain, or spent in the pool.
func (c *Client) IsKeyImageSpent(
	ctx context.Context, params IsKeyImageSpentRequestParameters,
) (*IsKeyImageSpentResult, error) {
	resp := &IsKeyImageSpentResult{}

	err := c.RawRequest(ctx, endpointIsKeyImageSpent, params, resp)
	if err != nil {
		return nil, fmt.Errorf("raw request: %w", err)
	}

	return resp, nil
}

// SendRawTransaction submits a hex-encoded transaction to the daemon.
//
// State-changing: unless DoNotRelay is set, a transaction the daemon
// accepts is broadcast to the network.
func (c *Client) SendRawTransaction(
	ctx context.Context, params SendRawTransactionRequestParameters,
) (*SendRawTransactionResult, error) {
	resp := &SendRawTransactionResult{}

	err := c.RawRequest(ctx, endpointSendRawTransaction, params, resp)
	if err != nil {
		return nil, fmt.Errorf("raw request: %w", err)
	}

	return resp, nil
}

// SaveBC flushes the blockchain to disk.
//
// State-changing: it syncs the filesystem cache. The chain is already
// saved as it changes; this call only forces that flush.
func (c *Client) SaveBC(ctx context.Context) (*SaveBCResult, error) {
	resp := &SaveBCResult{}

	err := c.RawRequest(ctx, endpointSaveBC, nil, resp)
	if err != nil {
		return nil, fmt.Errorf("raw request: %w", err)
	}

	return resp, nil
}

// StopDaemon asks the daemon to disconnect and shut down.
//
// State-changing: a successful call stops the monerod process.
func (c *Client) StopDaemon(ctx context.Context) (*StopDaemonResult, error) {
	resp := &StopDaemonResult{}

	err := c.RawRequest(ctx, endpointStopDaemon, nil, resp)
	if err != nil {
		return nil, fmt.Errorf("raw request: %w", err)
	}

	return resp, nil
}

// SetBootstrapDaemon sets the daemon used to proxy RPC while this node
// is syncing.
//
// State-changing: later RPC calls may be served by that remote daemon
// and reported as untrusted. An empty address clears the setting.
func (c *Client) SetBootstrapDaemon(
	ctx context.Context, params SetBootstrapDaemonRequestParameters,
) (*SetBootstrapDaemonResult, error) {
	resp := &SetBootstrapDaemonResult{}

	err := c.RawRequest(ctx, endpointSetBootstrapDaemon, params, resp)
	if err != nil {
		return nil, fmt.Errorf("raw request: %w", err)
	}

	return resp, nil
}

// PopBlocks removes blocks from the tip of the chain.
//
// State-changing: it rewrites the local chain. By default, non-coinbase
// transactions from the popped blocks return to the mempool.
func (c *Client) PopBlocks(
	ctx context.Context, params PopBlocksRequestParameters,
) (*PopBlocksResult, error) {
	resp := &PopBlocksResult{}

	err := c.RawRequest(ctx, endpointPopBlocks, params, resp)
	if err != nil {
		return nil, fmt.Errorf("raw request: %w", err)
	}

	return resp, nil
}

// SetLogHashRate shows or hides hash-rate lines in the daemon log.
//
// State-changing: it changes the daemon's logging configuration. The
// status is "NOT MINING" when the daemon is not mining.
func (c *Client) SetLogHashRate(
	ctx context.Context, params SetLogHashRateRequestParameters,
) (*SetLogHashRateResult, error) {
	resp := &SetLogHashRateResult{}

	err := c.RawRequest(ctx, endpointSetLogHashRate, params, resp)
	if err != nil {
		return nil, fmt.Errorf("raw request: %w", err)
	}

	return resp, nil
}
