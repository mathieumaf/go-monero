package wallet_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieumaf/go-monero/pkg/rpc/wallet"
)

func TestTransferJSONRoundTrip(t *testing.T) {
	const raw = `{
		"address": "53zii2WaqQwZU4oUsCUcrHgaSv2CrUGCSFJLdQnkLPyH7ZLPYHjtoHhi14dqjF6jywNRknYLwbate2eGv8TuZcS7GuR7wMY",
		"amount": 100000000000,
		"amounts": [100000000000],
		"confirmations": 19,
		"double_spend_seen": false,
		"fee": 53840000,
		"height": 1140109,
		"locked": false,
		"note": "",
		"payment_id": "0000000000000000",
		"subaddr_index": {"major": 0, "minor": 0},
		"subaddr_indices": [{"major": 0, "minor": 0}],
		"suggested_confirmations_threshold": 1,
		"timestamp": 1658360753,
		"txid": "765f7124d01bd2eb2d4e7e59aa44a28c24339a41e4009f463955b087017b0ca3",
		"type": "in",
		"unlock_time": 0
	}`

	var tr wallet.Transfer
	require.NoError(t, json.Unmarshal([]byte(raw), &tr))

	assert.Equal(t, "in", tr.Type)
	assert.Equal(t, uint64(100000000000), tr.Amount)
	assert.Equal(t, uint64(19), tr.Confirmations)
	assert.Equal(t, uint32(0), tr.SubaddrIndex.Major)
	assert.Equal(t, uint32(0), tr.SubaddrIndex.Minor)
	assert.Equal(t,
		"765f7124d01bd2eb2d4e7e59aa44a28c24339a41e4009f463955b087017b0ca3",
		tr.Txid,
	)
}

func TestValidateAddressResultJSON(t *testing.T) {
	const raw = `{
		"valid": true,
		"integrated": false,
		"subaddress": true,
		"nettype": "mainnet",
		"openalias_address": ""
	}`

	var res wallet.ValidateAddressResult
	require.NoError(t, json.Unmarshal([]byte(raw), &res))
	assert.True(t, res.Valid)
	assert.True(t, res.Subaddress)
	assert.Equal(t, "mainnet", res.Nettype)
}

func TestGetTransfersResultJSON(t *testing.T) {
	const raw = `{
		"in": [{"txid": "aa", "type": "in", "amount": 1}],
		"out": [],
		"pending": [],
		"failed": [],
		"pool": [{"txid": "bb", "type": "pool", "amount": 2}]
	}`

	var res wallet.GetTransfersResult
	require.NoError(t, json.Unmarshal([]byte(raw), &res))
	require.Len(t, res.In, 1)
	require.Len(t, res.Pool, 1)
	assert.Equal(t, "aa", res.In[0].Txid)
	assert.Equal(t, uint64(2), res.Pool[0].Amount)
}

func TestIncomingTransferJSON(t *testing.T) {
	const raw = `{
		"amount": 1000,
		"spent": false,
		"global_index": 42,
		"tx_hash": "deadbeef",
		"subaddr_index": {"major": 1, "minor": 2},
		"key_image": "ki",
		"pubkey": "pk",
		"block_height": 99,
		"frozen": false,
		"unlocked": true
	}`

	var tr wallet.IncomingTransfer
	require.NoError(t, json.Unmarshal([]byte(raw), &tr))
	assert.Equal(t, uint64(1000), tr.Amount)
	assert.Equal(t, uint32(1), tr.SubaddrIndex.Major)
	assert.Equal(t, uint32(2), tr.SubaddrIndex.Minor)
	assert.True(t, tr.Unlocked)
}

func TestCheckReserveProofResultJSON(t *testing.T) {
	const raw = `{"good": true, "spent": 0, "total": 100000000000}`

	var res wallet.CheckReserveProofResult
	require.NoError(t, json.Unmarshal([]byte(raw), &res))
	assert.True(t, res.Good)
	assert.Equal(t, uint64(100000000000), res.Total)
}
