package wallet

import (
	"context"
	"fmt"
)

const (
	methodAutoRefresh       = "auto_refresh"
	methodCheckReserveProof = "check_reserve_proof"
	methodCheckSpendProof   = "check_spend_proof"
	methodCheckTxKey        = "check_tx_key"
	methodCheckTxProof      = "check_tx_proof"
	methodCreateAddress     = "create_address"
	methodGetAccounts       = "get_accounts"
	methodGetAddress        = "get_address"
	methodGetAddressIndex   = "get_address_index"
	methodGetAttribute      = "get_attribute"
	methodGetBalance        = "get_balance"
	methodGetBulkPayments   = "get_bulk_payments"
	methodGetHeight         = "get_height"
	methodGetLanguages      = "get_languages"
	methodGetPayments       = "get_payments"
	methodGetTransferByTxid = "get_transfer_by_txid"
	methodGetTransfers      = "get_transfers"
	methodGetTxNotes        = "get_tx_notes"
	methodIncomingTransfers = "incoming_transfers"
	methodQueryKey          = "query_key"
	methodRefresh           = "refresh"
	methodValidateAddress   = "validate_address"
)

// GetAccounts returns all accounts for a wallet, optionally filtered by tag.
func (c *Client) GetAccounts(
	ctx context.Context, params GetAccountsRequestParameters,
) (*GetAccountsResult, error) {
	resp := &GetAccountsResult{}

	if err := c.JSONRPC(ctx, methodGetAccounts, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetAddress returns the wallet's addresses for an account.
func (c *Client) GetAddress(
	ctx context.Context, params GetAddressRequestParameters,
) (*GetAddressResult, error) {
	resp := &GetAddressResult{}

	if err := c.JSONRPC(ctx, methodGetAddress, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetBalance returns the wallet's balance for the given account / subaddresses.
func (c *Client) GetBalance(
	ctx context.Context, params GetBalanceRequestParameters,
) (*GetBalanceResult, error) {
	resp := &GetBalanceResult{}

	if err := c.JSONRPC(ctx, methodGetBalance, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// CreateAddress creates a new address (or several) for an account.
func (c *Client) CreateAddress(
	ctx context.Context, accountIndex uint, count uint, label string,
) (*CreateAddressResult, error) {
	resp := &CreateAddressResult{}

	params := map[string]interface{}{
		"account_index": accountIndex,
		"label":         label,
		"count":         count,
	}
	if err := c.JSONRPC(ctx, methodCreateAddress, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// AutoRefresh configures whether and how often the wallet auto-refreshes.
func (c *Client) AutoRefresh(
	ctx context.Context, enable bool, period int64,
) (*AutoRefreshResult, error) {
	resp := &AutoRefreshResult{}

	params := map[string]interface{}{
		"enable": enable,
		"period": period,
	}
	if err := c.JSONRPC(ctx, methodAutoRefresh, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// Refresh refreshes the wallet from the daemon starting at startHeight.
func (c *Client) Refresh(
	ctx context.Context, startHeight uint64,
) (*RefreshResult, error) {
	resp := &RefreshResult{}

	params := map[string]interface{}{
		"start_height": startHeight,
	}
	if err := c.JSONRPC(ctx, methodRefresh, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetHeight returns the wallet's current blockchain height.
func (c *Client) GetHeight(ctx context.Context) (*GetHeightResult, error) {
	resp := &GetHeightResult{}

	if err := c.JSONRPC(ctx, methodGetHeight, nil, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetTransfers returns transfers matching the given filters.
func (c *Client) GetTransfers(
	ctx context.Context, params GetTransfersRequestParameters,
) (*GetTransfersResult, error) {
	resp := &GetTransfersResult{}

	if err := c.JSONRPC(ctx, methodGetTransfers, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetTransferByTxid returns transfer details for a single transaction id.
func (c *Client) GetTransferByTxid(
	ctx context.Context, params GetTransferByTxidRequestParameters,
) (*GetTransferByTxidResult, error) {
	resp := &GetTransferByTxidResult{}

	if err := c.JSONRPC(ctx, methodGetTransferByTxid, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// IncomingTransfers returns owned outputs (incoming transfers / UTXOs).
//
// transfer_type must be one of: "all", "available", "unavailable".
func (c *Client) IncomingTransfers(
	ctx context.Context, params IncomingTransfersRequestParameters,
) (*IncomingTransfersResult, error) {
	resp := &IncomingTransfersResult{}

	if err := c.JSONRPC(ctx, methodIncomingTransfers, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetPayments returns incoming payments for a single payment id.
func (c *Client) GetPayments(
	ctx context.Context, params GetPaymentsRequestParameters,
) (*GetPaymentsResult, error) {
	resp := &GetPaymentsResult{}

	if err := c.JSONRPC(ctx, methodGetPayments, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetBulkPayments returns incoming payments for one or more payment ids,
// optionally starting from a minimum block height. Preferred over GetPayments
// when querying multiple ids.
func (c *Client) GetBulkPayments(
	ctx context.Context, params GetBulkPaymentsRequestParameters,
) (*GetBulkPaymentsResult, error) {
	resp := &GetBulkPaymentsResult{}

	if err := c.JSONRPC(ctx, methodGetBulkPayments, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// ValidateAddress checks whether an address (or OpenAlias) is valid.
func (c *Client) ValidateAddress(
	ctx context.Context, params ValidateAddressRequestParameters,
) (*ValidateAddressResult, error) {
	resp := &ValidateAddressResult{}

	if err := c.JSONRPC(ctx, methodValidateAddress, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetAddressIndex resolves a (sub)address to its account and address indices.
func (c *Client) GetAddressIndex(
	ctx context.Context, params GetAddressIndexRequestParameters,
) (*GetAddressIndexResult, error) {
	resp := &GetAddressIndexResult{}

	if err := c.JSONRPC(ctx, methodGetAddressIndex, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// QueryKey returns a wallet key. key_type must be "mnemonic", "view_key", or
// "spend_key". Unavailable in restricted RPC mode.
func (c *Client) QueryKey(
	ctx context.Context, params QueryKeyRequestParameters,
) (*QueryKeyResult, error) {
	resp := &QueryKeyResult{}

	if err := c.JSONRPC(ctx, methodQueryKey, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// CheckTxKey verifies a transaction using its secret key and destination
// address.
func (c *Client) CheckTxKey(
	ctx context.Context, params CheckTxKeyRequestParameters,
) (*CheckTxKeyResult, error) {
	resp := &CheckTxKeyResult{}

	if err := c.JSONRPC(ctx, methodCheckTxKey, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// CheckTxProof verifies a transaction proof signature.
func (c *Client) CheckTxProof(
	ctx context.Context, params CheckTxProofRequestParameters,
) (*CheckTxProofResult, error) {
	resp := &CheckTxProofResult{}

	if err := c.JSONRPC(ctx, methodCheckTxProof, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// CheckSpendProof verifies a spend proof signature.
func (c *Client) CheckSpendProof(
	ctx context.Context, params CheckSpendProofRequestParameters,
) (*CheckSpendProofResult, error) {
	resp := &CheckSpendProofResult{}

	if err := c.JSONRPC(ctx, methodCheckSpendProof, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// CheckReserveProof verifies a reserve proof signature.
func (c *Client) CheckReserveProof(
	ctx context.Context, params CheckReserveProofRequestParameters,
) (*CheckReserveProofResult, error) {
	resp := &CheckReserveProofResult{}

	if err := c.JSONRPC(ctx, methodCheckReserveProof, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetTxNotes returns notes attached to the given transaction ids.
func (c *Client) GetTxNotes(
	ctx context.Context, params GetTxNotesRequestParameters,
) (*GetTxNotesResult, error) {
	resp := &GetTxNotesResult{}

	if err := c.JSONRPC(ctx, methodGetTxNotes, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetAttribute returns a wallet attribute value by key.
func (c *Client) GetAttribute(
	ctx context.Context, params GetAttributeRequestParameters,
) (*GetAttributeResult, error) {
	resp := &GetAttributeResult{}

	if err := c.JSONRPC(ctx, methodGetAttribute, params, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}

// GetLanguages returns available wallet seed languages.
func (c *Client) GetLanguages(ctx context.Context) (*GetLanguagesResult, error) {
	resp := &GetLanguagesResult{}

	if err := c.JSONRPC(ctx, methodGetLanguages, nil, resp); err != nil {
		return nil, fmt.Errorf("jsonrpc: %w", err)
	}

	return resp, nil
}
