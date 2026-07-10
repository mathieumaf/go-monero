package wallet

// SubaddressIndex identifies a wallet subaddress (account = major, address = minor).
type SubaddressIndex struct {
	// Major is the account index.
	Major uint32 `json:"major"`

	// Minor is the address index within the account.
	Minor uint32 `json:"minor"`
}

// TransferDestination is a single destination of an outgoing transfer.
type TransferDestination struct {
	// Amount is the amount sent to Address, in atomic units.
	Amount uint64 `json:"amount"`

	// Address is the destination public address (base58).
	Address string `json:"address"`
}

// Transfer is a wallet transfer entry returned by get_transfers / get_transfer_by_txid.
type Transfer struct {
	// Txid is the transaction id.
	Txid string `json:"txid"`

	// PaymentID is the payment id for this transfer (may be all zeros).
	PaymentID string `json:"payment_id"`

	// Height is the height of the first block that confirmed this transfer
	// (0 if unconfirmed).
	Height uint64 `json:"height"`

	// Timestamp is the POSIX timestamp of the confirming block (or submission
	// time if not yet mined).
	Timestamp uint64 `json:"timestamp"`

	// Amount is the total amount of this transfer in atomic units.
	Amount uint64 `json:"amount"`

	// Amounts lists individual output amounts when multiple outputs were
	// received in the same transaction.
	Amounts []uint64 `json:"amounts"`

	// Fee is the transaction fee in atomic units.
	Fee uint64 `json:"fee"`

	// Note is a user-defined note attached to this transfer.
	Note string `json:"note"`

	// Destinations lists destinations for outgoing transfers. Only present
	// when this wallet cache constructed the transaction.
	Destinations []TransferDestination `json:"destinations"`

	// Type is one of: "in", "out", "pending", "failed", "pool".
	Type string `json:"type"`

	// UnlockTime is the number of blocks until the transfer is safely
	// spendable (0 = unlocked by height rule).
	UnlockTime uint64 `json:"unlock_time"`

	// Locked is true if the transfer is not yet spendable.
	Locked bool `json:"locked"`

	// SubaddrIndex is the primary subaddress index for this transfer.
	SubaddrIndex SubaddressIndex `json:"subaddr_index"`

	// SubaddrIndices lists all subaddress indices involved.
	SubaddrIndices []SubaddressIndex `json:"subaddr_indices"`

	// Address is the (sub)address related to this transfer (base58).
	Address string `json:"address"`

	// DoubleSpendSeen is true if the key image(s) have been seen before.
	DoubleSpendSeen bool `json:"double_spend_seen"`

	// Confirmations is the number of blocks mined since the confirming
	// block (or until inclusion if still unconfirmed).
	Confirmations uint64 `json:"confirmations"`

	// SuggestedConfirmationsThreshold is a wallet-suggested confirmation
	// count for amounts relative to block rewards.
	SuggestedConfirmationsThreshold uint64 `json:"suggested_confirmations_threshold"`
}

// Payment is an incoming payment matched by payment id.
type Payment struct {
	// PaymentID is the payment id that matched.
	PaymentID string `json:"payment_id"`

	// TxHash is the transaction hash (tx id).
	TxHash string `json:"tx_hash"`

	// Amount is the payment amount in atomic units.
	Amount uint64 `json:"amount"`

	// BlockHeight is the height of the block that first confirmed this payment.
	BlockHeight uint64 `json:"block_height"`

	// UnlockTime is the block height until which this payment is locked.
	UnlockTime uint64 `json:"unlock_time"`

	// Locked is true if the output is not yet spendable.
	Locked bool `json:"locked"`

	// SubaddrIndex is the receiving subaddress index.
	SubaddrIndex SubaddressIndex `json:"subaddr_index"`

	// Address is the receiving address (base58).
	Address string `json:"address"`
}

// IncomingTransfer is a single owned output (UTXO-like) returned by
// incoming_transfers.
type IncomingTransfer struct {
	// Amount is the output amount in atomic units.
	Amount uint64 `json:"amount"`

	// Spent is true if the output has been spent.
	Spent bool `json:"spent"`

	// GlobalIndex is the output's global index in the blockchain.
	GlobalIndex uint64 `json:"global_index"`

	// TxHash is the hash of the transaction that created this output.
	TxHash string `json:"tx_hash"`

	// SubaddrIndex is the subaddress that owns this output.
	SubaddrIndex SubaddressIndex `json:"subaddr_index"`

	// KeyImage is the key image of this output (hex).
	KeyImage string `json:"key_image"`

	// Pubkey is the owned output public key (hex).
	Pubkey string `json:"pubkey"`

	// BlockHeight is the height of the block that included this output.
	BlockHeight uint64 `json:"block_height"`

	// Frozen is true if the output has been frozen via the freeze RPC.
	Frozen bool `json:"frozen"`

	// Unlocked is true if the output is currently unlocked/spendable.
	Unlocked bool `json:"unlocked"`
}

// GetAccountsRequestParameters are the parameters for get_accounts.
type GetAccountsRequestParameters struct {
	// Tag optionally filters accounts by tag.
	Tag string `json:"tag,omitempty"`

	// StrictBalances, when true, only considers the blockchain for balances
	// (ignoring recent unmined actions).
	StrictBalances bool `json:"strict_balances,omitempty"`
}

// GetAccountsResult is the result of get_accounts.
type GetAccountsResult struct {
	SubaddressAccounts []struct {
		AccountIndex    uint   `json:"account_index"`
		Balance         uint64 `json:"balance"`
		BaseAddress     string `json:"base_address"`
		Label           string `json:"label"`
		Tag             string `json:"tag"`
		UnlockedBalance uint64 `json:"unlocked_balance"`
	} `json:"subaddress_accounts"`

	TotalBalance         uint64 `json:"total_balance"`
	TotalUnlockedBalance uint64 `json:"total_unlocked_balance"`
}

// GetAddressRequestParameters are the parameters for get_address.
type GetAddressRequestParameters struct {
	AccountIndex   uint   `json:"account_index"`
	AddressIndices []uint `json:"address_indices,omitempty"`
}

// GetAddressResult is the result of get_address.
type GetAddressResult struct {
	Address   string `json:"address"`
	Addresses []struct {
		Address      string `json:"address"`
		AddressIndex uint   `json:"address_index"`
		Label        string `json:"label"`
		Used         bool   `json:"used"`
	} `json:"addresses"`
}

// GetBalanceRequestParameters are the parameters for get_balance.
type GetBalanceRequestParameters struct {
	AccountIndex   uint   `json:"account_index"`
	AddressIndices []uint `json:"address_indices,omitempty"`
	AllAccounts    bool   `json:"all_accounts,omitempty"`
	Strict         bool   `json:"strict,omitempty"`
}

// GetBalanceResult is the result of get_balance.
type GetBalanceResult struct {
	// Balance is the total balance of the current monero-wallet-rpc session
	// in atomic units.
	Balance uint64 `json:"balance"`

	// BlocksToUnlock is how many blocks are necessary before all funds are
	// unlocked.
	BlocksToUnlock uint `json:"blocks_to_unlock"`

	// MultisigImportNeeded is true if importing multisig data is needed to
	// return a correct balance.
	MultisigImportNeeded bool `json:"multisig_import_needed"`

	// PerSubaddress is balance information for each subaddress in the
	// requested account(s).
	PerSubaddress []SubAddress `json:"per_subaddress"`

	// TimeToUnlock is the time in seconds before the balance is safe to spend.
	TimeToUnlock uint64 `json:"time_to_unlock"`

	// UnlockedBalance is the unlocked (spendable) balance in atomic units.
	UnlockedBalance uint64 `json:"unlocked_balance"`
}

// SubAddress holds balance information for a single subaddress.
type SubAddress struct {
	// AccountIndex is the index of the account.
	AccountIndex uint `json:"account_index"`

	// Address is the address at this index (base58 representation of the
	// public keys).
	Address string `json:"address"`

	// AddressIndex is the index of the subaddress in the account.
	AddressIndex uint `json:"address_index"`

	// Balance is the balance for the subaddress in atomic units.
	Balance uint64 `json:"balance"`

	// BlocksToUnlock is how many blocks until the subaddress balance is fully
	// unlocked.
	BlocksToUnlock uint `json:"blocks_to_unlock"`

	// Label is the user-defined label for the subaddress.
	Label string `json:"label"`

	// NumUnspentOutputs is the number of unspent outputs for the subaddress.
	NumUnspentOutputs uint `json:"num_unspent_outputs"`

	// TimeToUnlock is the time in seconds before the subaddress balance is
	// safe to spend.
	TimeToUnlock uint64 `json:"time_to_unlock"`

	// UnlockedBalance is the unlocked balance for the subaddress in atomic
	// units.
	UnlockedBalance uint64 `json:"unlocked_balance"`
}

// CreateAddressResult is the result of create_address.
type CreateAddressResult struct {
	Address        string   `json:"address"`
	AddressIndex   uint     `json:"address_index"`
	AddressIndices []uint   `json:"address_indices"`
	Addresses      []string `json:"addresses"`
}

// RefreshResult is the result of refresh.
type RefreshResult struct {
	BlocksFetched uint64 `json:"blocks_fetched"`
	ReceivedMoney bool   `json:"received_money"`
}

// AutoRefreshResult is the result of auto_refresh (empty object).
type AutoRefreshResult struct{}

// GetHeightResult is the result of get_height.
type GetHeightResult struct {
	// Height is the wallet's current blockchain height.
	Height uint64 `json:"height"`
}

// GetTransfersRequestParameters are the parameters for get_transfers.
//
// Boolean type filters (In, Out, Pending, Failed, Pool) select which transfer
// categories to return. At least one should typically be true.
type GetTransfersRequestParameters struct {
	// In includes incoming transfers.
	In bool `json:"in,omitempty"`

	// Out includes outgoing transfers.
	Out bool `json:"out,omitempty"`

	// Pending includes pending (unconfirmed) transfers.
	Pending bool `json:"pending,omitempty"`

	// Failed includes failed transfers.
	Failed bool `json:"failed,omitempty"`

	// Pool includes transfers still in the tx pool.
	Pool bool `json:"pool,omitempty"`

	// FilterByHeight restricts results to [MinHeight, MaxHeight].
	FilterByHeight bool `json:"filter_by_height,omitempty"`

	// MinHeight is the inclusive lower height bound when FilterByHeight is set.
	MinHeight uint64 `json:"min_height,omitempty"`

	// MaxHeight is the inclusive upper height bound when FilterByHeight is set.
	// monerod defaults this to a very large value when omitted.
	MaxHeight uint64 `json:"max_height,omitempty"`

	// AccountIndex selects the account to query (ignored when AllAccounts).
	AccountIndex uint32 `json:"account_index,omitempty"`

	// SubaddrIndices optionally restricts to specific subaddress indices.
	SubaddrIndices []uint32 `json:"subaddr_indices,omitempty"`

	// AllAccounts queries every account when true.
	AllAccounts bool `json:"all_accounts,omitempty"`
}

// GetTransfersResult is the result of get_transfers. Only categories requested
// in the parameters are populated.
type GetTransfersResult struct {
	In      []Transfer `json:"in"`
	Out     []Transfer `json:"out"`
	Pending []Transfer `json:"pending"`
	Failed  []Transfer `json:"failed"`
	Pool    []Transfer `json:"pool"`
}

// GetTransferByTxidRequestParameters are the parameters for get_transfer_by_txid.
type GetTransferByTxidRequestParameters struct {
	// Txid is the transaction id to look up.
	Txid string `json:"txid"`

	// AccountIndex optionally scopes the lookup to an account (default 0).
	AccountIndex uint32 `json:"account_index,omitempty"`
}

// GetTransferByTxidResult is the result of get_transfer_by_txid.
type GetTransferByTxidResult struct {
	// Transfer is the primary transfer entry for the txid.
	Transfer Transfer `json:"transfer"`

	// Transfers lists all transfer entries when multiple outputs were
	// received in the same transaction.
	Transfers []Transfer `json:"transfers"`
}

// IncomingTransfersRequestParameters are the parameters for incoming_transfers.
type IncomingTransfersRequestParameters struct {
	// TransferType filters outputs: "all", "available", or "unavailable".
	TransferType string `json:"transfer_type"`

	// AccountIndex selects the account.
	AccountIndex uint32 `json:"account_index,omitempty"`

	// SubaddrIndices optionally restricts to specific subaddress indices.
	SubaddrIndices []uint32 `json:"subaddr_indices,omitempty"`
}

// IncomingTransfersResult is the result of incoming_transfers.
type IncomingTransfersResult struct {
	Transfers []IncomingTransfer `json:"transfers"`
}

// GetPaymentsRequestParameters are the parameters for get_payments.
type GetPaymentsRequestParameters struct {
	// PaymentID is a 16-character hex payment id.
	PaymentID string `json:"payment_id"`
}

// GetPaymentsResult is the result of get_payments.
type GetPaymentsResult struct {
	Payments []Payment `json:"payments"`
}

// GetBulkPaymentsRequestParameters are the parameters for get_bulk_payments.
type GetBulkPaymentsRequestParameters struct {
	// PaymentIDs is the list of 16-character hex payment ids.
	PaymentIDs []string `json:"payment_ids"`

	// MinBlockHeight is the height at which to start looking for payments.
	MinBlockHeight uint64 `json:"min_block_height,omitempty"`
}

// GetBulkPaymentsResult is the result of get_bulk_payments.
type GetBulkPaymentsResult struct {
	Payments []Payment `json:"payments"`
}

// ValidateAddressRequestParameters are the parameters for validate_address.
type ValidateAddressRequestParameters struct {
	// Address is the address (or OpenAlias) to validate.
	Address string `json:"address"`

	// AnyNetType accepts addresses from any network type when true.
	AnyNetType bool `json:"any_net_type,omitempty"`

	// AllowOpenalias resolves OpenAlias addresses when true.
	AllowOpenalias bool `json:"allow_openalias,omitempty"`
}

// ValidateAddressResult is the result of validate_address.
type ValidateAddressResult struct {
	// Valid is true if the address is valid.
	Valid bool `json:"valid"`

	// Integrated is true if the address is an integrated address.
	Integrated bool `json:"integrated"`

	// Subaddress is true if the address is a subaddress.
	Subaddress bool `json:"subaddress"`

	// Nettype is the network type: "mainnet", "testnet", "stagenet".
	Nettype string `json:"nettype"`

	// OpenaliasAddress is the resolved OpenAlias address when applicable.
	OpenaliasAddress string `json:"openalias_address"`
}

// GetAddressIndexRequestParameters are the parameters for get_address_index.
type GetAddressIndexRequestParameters struct {
	// Address is the (sub)address to look up.
	Address string `json:"address"`
}

// GetAddressIndexResult is the result of get_address_index.
type GetAddressIndexResult struct {
	// Index is the account (major) and address (minor) index.
	Index SubaddressIndex `json:"index"`
}

// QueryKeyRequestParameters are the parameters for query_key.
//
// KeyType must be one of: "mnemonic", "view_key", "spend_key".
// Restricted RPC mode denies this method.
type QueryKeyRequestParameters struct {
	// KeyType selects which key to return: "mnemonic", "view_key", or
	// "spend_key".
	KeyType string `json:"key_type"`
}

// QueryKeyResult is the result of query_key.
type QueryKeyResult struct {
	// Key is the requested key material (mnemonic words or hex secret key).
	Key string `json:"key"`
}

// CheckTxKeyRequestParameters are the parameters for check_tx_key.
type CheckTxKeyRequestParameters struct {
	// Txid is the transaction id.
	Txid string `json:"txid"`

	// TxKey is the transaction secret key.
	TxKey string `json:"tx_key"`

	// Address is the destination public address of the transaction.
	Address string `json:"address"`
}

// CheckTxKeyResult is the result of check_tx_key.
type CheckTxKeyResult struct {
	// Confirmations is the number of blocks mined after the one with the
	// transaction.
	Confirmations uint64 `json:"confirmations"`

	// InPool is true if the transaction is still in the pool.
	InPool bool `json:"in_pool"`

	// Received is the amount received by Address in atomic units.
	Received uint64 `json:"received"`
}

// CheckTxProofRequestParameters are the parameters for check_tx_proof.
type CheckTxProofRequestParameters struct {
	// Txid is the transaction id.
	Txid string `json:"txid"`

	// Address is the destination public address of the transaction.
	Address string `json:"address"`

	// Message is the optional message used when creating the proof.
	Message string `json:"message,omitempty"`

	// Signature is the transaction proof signature to verify.
	Signature string `json:"signature"`
}

// CheckTxProofResult is the result of check_tx_proof.
type CheckTxProofResult struct {
	// Confirmations is the number of blocks mined after the one with the
	// transaction.
	Confirmations uint64 `json:"confirmations"`

	// Good is true if the proof is valid.
	Good bool `json:"good"`

	// InPool is true if the transaction is still in the pool.
	InPool bool `json:"in_pool"`

	// Received is the amount received by Address in atomic units.
	Received uint64 `json:"received"`
}

// CheckSpendProofRequestParameters are the parameters for check_spend_proof.
type CheckSpendProofRequestParameters struct {
	// Txid is the transaction id.
	Txid string `json:"txid"`

	// Message is the optional message used when creating the proof.
	Message string `json:"message,omitempty"`

	// Signature is the spend proof signature to verify.
	Signature string `json:"signature"`
}

// CheckSpendProofResult is the result of check_spend_proof.
type CheckSpendProofResult struct {
	// Good is true if the proof is valid.
	Good bool `json:"good"`
}

// CheckReserveProofRequestParameters are the parameters for check_reserve_proof.
type CheckReserveProofRequestParameters struct {
	// Address is the public address of the wallet that produced the proof.
	Address string `json:"address"`

	// Message is the optional message used when creating the proof.
	Message string `json:"message,omitempty"`

	// Signature is the reserve proof signature to verify.
	Signature string `json:"signature"`
}

// CheckReserveProofResult is the result of check_reserve_proof.
type CheckReserveProofResult struct {
	// Good is true if the proof is valid.
	Good bool `json:"good"`

	// Spent is the amount of the proven reserve that has been spent, in
	// atomic units.
	Spent uint64 `json:"spent"`

	// Total is the total amount proven in atomic units.
	Total uint64 `json:"total"`
}

// GetTxNotesRequestParameters are the parameters for get_tx_notes.
type GetTxNotesRequestParameters struct {
	// Txids is the list of transaction ids to fetch notes for.
	Txids []string `json:"txids"`
}

// GetTxNotesResult is the result of get_tx_notes.
type GetTxNotesResult struct {
	// Notes are the notes for each requested txid (same order).
	Notes []string `json:"notes"`
}

// GetAttributeRequestParameters are the parameters for get_attribute.
type GetAttributeRequestParameters struct {
	// Key is the attribute name.
	Key string `json:"key"`
}

// GetAttributeResult is the result of get_attribute.
type GetAttributeResult struct {
	// Value is the attribute value.
	Value string `json:"value"`
}

// GetLanguagesResult is the result of get_languages.
type GetLanguagesResult struct {
	// Languages is the list of available seed languages (English names).
	Languages []string `json:"languages"`

	// LanguagesLocal is the list of seed languages in their native names.
	LanguagesLocal []string `json:"languages_local"`
}
