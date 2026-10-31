package daemon_test

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

// jsonBinaryString encodes blob the way a strict JSON producer can carry
// monerod's packed records: bytes below 0x20 become \u00XX and every other
// byte is copied raw, including 0x80-0xff.
func jsonBinaryString(blob []byte) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, c := range blob {
		switch c {
		case '"', '\\':
			b.WriteByte('\\')
			b.WriteByte(c)
		default:
			if c < 0x20 {
				fmt.Fprintf(&b, `\u%04x`, c)
			} else {
				b.WriteByte(c)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}

func packUint64s(words ...uint64) []byte {
	buf := make([]byte, 8*len(words))
	for i, word := range words {
		binary.LittleEndian.PutUint64(buf[i*8:], word)
	}
	return buf
}

func TestParseTxpoolBacklog(t *testing.T) {
	t.Parallel()

	blob := packUint64s(2000, 100000, 30, 1535, 30700000, 12)
	got, err := daemon.ParseTxpoolBacklog(blob)
	require.NoError(t, err)
	assert.Equal(t, []daemon.TxBacklogEntry{
		{Weight: 2000, Fee: 100000, TimeInPool: 30},
		{Weight: 1535, Fee: 30700000, TimeInPool: 12},
	}, got)

	empty, err := daemon.ParseTxpoolBacklog(nil)
	require.NoError(t, err)
	assert.Empty(t, empty)

	_, err = daemon.ParseTxpoolBacklog([]byte{1, 2, 3})
	require.Error(t, err)

	_, err = daemon.ParseTxpoolBacklog(packUint64s(1))
	require.Error(t, err)
}

func TestGetTxpoolBacklog(t *testing.T) {
	t.Parallel()

	blob := packUint64s(2000, 100000, 30)
	reply := rpcEnvelope(t, json.RawMessage(fmt.Sprintf(
		`{"backlog":%s,"status":"OK","untrusted":false,"credits":4,"top_hash":"aa"}`,
		jsonBinaryString(blob),
	)))

	client, call := serveDaemon(t, reply)
	got, err := client.GetTxpoolBacklog(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "get_txpool_backlog", call.Method)
	assert.Empty(t, call.Params)
	assert.Equal(t, "OK", got.Status)
	assert.False(t, got.Untrusted)
	assert.Equal(t, uint64(4), got.Credits)
	assert.Equal(t, "aa", got.TopHash)
	require.Len(t, got.Entries, 1)
	assert.Equal(t, uint64(2000), got.Entries[0].Weight)
	assert.Equal(t, uint64(100000), got.Entries[0].Fee)
	assert.Equal(t, uint64(30), got.Entries[0].TimeInPool)
}

func TestGetTxpoolBacklogEmptyAndInvalid(t *testing.T) {
	t.Parallel()

	emptyReply := rpcEnvelope(t, json.RawMessage(`{"status":"OK","untrusted":false}`))
	client, _ := serveDaemon(t, emptyReply)
	got, err := client.GetTxpoolBacklog(context.Background())
	require.NoError(t, err)
	assert.Empty(t, got.Entries)
	assert.Equal(t, "OK", got.Status)

	badReply := rpcEnvelope(t, json.RawMessage(`{"backlog":"abcd","status":"OK"}`))
	client, _ = serveDaemon(t, badReply)
	_, err = client.GetTxpoolBacklog(context.Background())
	require.Error(t, err)
}

func TestGetTxpoolBacklogRawNUL(t *testing.T) {
	t.Parallel()

	// monerod 0.18 writes the packed records with raw NUL bytes inside the
	// JSON string. encoding/json rejects that document.
	blob := packUint64s(2000, 100000, 30)
	body := []byte(`{"id":"0","jsonrpc":"2.0","result":{"backlog":"`)
	body = append(body, blob...)
	body = append(body, []byte(`","status":"OK","untrusted":false}}`)...)

	client, _ := serveDaemon(t, body)
	_, err := client.GetTxpoolBacklog(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode")
}

func TestGetMinerData(t *testing.T) {
	t.Parallel()

	reply := rpcEnvelope(t, json.RawMessage(`{
		"already_generated_coins": 18186022843595960691,
		"difficulty": "0x48afae42de",
		"height": 2731375,
		"major_version": 16,
		"median_weight": 300000,
		"prev_id": "1111111111111111111111111111111111111111111111111111111111111111",
		"seed_hash": "2222222222222222222222222222222222222222222222222222222222222222",
		"status": "OK",
		"tx_backlog": [{
			"fee": 30700000,
			"id": "3333333333333333333333333333333333333333333333333333333333333333",
			"weight": 1535
		}],
		"untrusted": false
	}`))

	client, call := serveDaemon(t, reply)
	got, err := client.GetMinerData(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "get_miner_data", call.Method)
	assert.Empty(t, call.Params)
	assert.Equal(t, uint64(16), got.MajorVersion)
	assert.Equal(t, uint64(2731375), got.Height)
	assert.Equal(t, "0x48afae42de", got.Difficulty)
	assert.Equal(t, uint64(300000), got.MedianWeight)
	assert.Equal(t, uint64(18186022843595960691), got.AlreadyGeneratedCoins)
	require.Len(t, got.TxBacklog, 1)
	assert.Equal(t, uint64(1535), got.TxBacklog[0].Weight)
	assert.Equal(t, uint64(30700000), got.TxBacklog[0].Fee)
	assert.Equal(t, "3333333333333333333333333333333333333333333333333333333333333333", got.TxBacklog[0].ID)
	assert.Equal(t, "OK", got.Status)
}

func TestBanned(t *testing.T) {
	t.Parallel()

	reply := rpcEnvelope(t, json.RawMessage(`{
		"banned": true,
		"seconds": 690413,
		"status": "OK"
	}`))

	client, call := serveDaemon(t, reply)
	got, err := client.Banned(context.Background(), daemon.BannedRequestParameters{
		Address: "203.0.113.50",
	})
	require.NoError(t, err)

	assert.Equal(t, "banned", call.Method)
	assert.JSONEq(t, `{"address":"203.0.113.50"}`, string(call.Params))
	assert.True(t, got.Banned)
	assert.Equal(t, uint64(690413), got.Seconds)
	assert.Equal(t, "OK", got.Status)
}

func TestFlushTxpool(t *testing.T) {
	t.Parallel()

	reply := rpcEnvelope(t, json.RawMessage(`{"status":"OK","untrusted":false}`))
	client, call := serveDaemon(t, reply)

	got, err := client.FlushTxpool(context.Background(), daemon.FlushTxpoolRequestParameters{
		Txids: []string{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	})
	require.NoError(t, err)
	assert.Equal(t, "flush_txpool", call.Method)
	assert.JSONEq(t,
		`{"txids":["aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"]}`,
		string(call.Params),
	)
	assert.Equal(t, "OK", got.Status)

	_, err = client.FlushTxpool(context.Background(), daemon.FlushTxpoolRequestParameters{})
	require.NoError(t, err)
	assert.JSONEq(t, `{"txids":[]}`, string(call.Params))
}

func TestGetOutputHistogram(t *testing.T) {
	t.Parallel()

	reply := rpcEnvelope(t, json.RawMessage(`{
		"credits": 0,
		"histogram": [{
			"amount": 20000000000,
			"recent_instances": 0,
			"total_instances": 381477,
			"unlocked_instances": 0
		}],
		"status": "OK",
		"top_hash": "",
		"untrusted": false
	}`))

	client, call := serveDaemon(t, reply)
	got, err := client.GetOutputHistogram(context.Background(), daemon.GetOutputHistogramRequestParameters{
		Amounts: []uint64{20000000000},
	})
	require.NoError(t, err)

	assert.Equal(t, "get_output_histogram", call.Method)
	assert.JSONEq(t, `{"amounts":[20000000000]}`, string(call.Params))
	require.Len(t, got.Histogram, 1)
	assert.Equal(t, uint64(20000000000), got.Histogram[0].Amount)
	assert.Equal(t, uint64(381477), got.Histogram[0].TotalInstances)
	assert.Equal(t, uint64(0), got.Histogram[0].UnlockedInstances)
	assert.Equal(t, uint64(0), got.Histogram[0].RecentInstances)
	assert.Equal(t, "OK", got.Status)
}

func TestGetOutputDistribution(t *testing.T) {
	t.Parallel()

	reply := rpcEnvelope(t, json.RawMessage(`{
		"distributions": [{
			"amount": 0,
			"base": 10,
			"binary": false,
			"compress": false,
			"distribution": [1, 2, 3, 4],
			"start_height": 1462078
		}],
		"status": "OK",
		"untrusted": false
	}`))

	client, call := serveDaemon(t, reply)
	got, err := client.GetOutputDistribution(context.Background(), daemon.GetOutputDistributionRequestParameters{
		Amounts:    []uint64{0},
		FromHeight: 1462078,
	})
	require.NoError(t, err)

	assert.Equal(t, "get_output_distribution", call.Method)
	assert.JSONEq(t, `{
		"amounts": [0],
		"cumulative": false,
		"from_height": 1462078,
		"to_height": 0,
		"binary": false,
		"compress": false
	}`, string(call.Params))
	require.Len(t, got.Distributions, 1)
	assert.Equal(t, uint64(0), got.Distributions[0].Amount)
	assert.Equal(t, uint64(10), got.Distributions[0].Base)
	assert.Equal(t, uint64(1462078), got.Distributions[0].StartHeight)
	counts, err := got.Distributions[0].Counts()
	require.NoError(t, err)
	assert.Equal(t, []uint64{1, 2, 3, 4}, counts)
	assert.Equal(t, "OK", got.Status)
}

func TestOutputDistributionBlobAndCompressed(t *testing.T) {
	t.Parallel()

	var blob daemon.OutputDistribution
	require.NoError(t, json.Unmarshal([]byte(fmt.Sprintf(
		`{"distribution":%s,"binary":true}`,
		jsonBinaryString(packUint64s(0x180, 5)),
	)), &blob))
	counts, err := blob.Counts()
	require.NoError(t, err)
	assert.Equal(t, []uint64{0x180, 5}, counts)

	compressed := daemon.OutputDistribution{Compress: true, CompressedData: "00"}
	_, err = compressed.Counts()
	require.Error(t, err)
}

func TestSubmitBlock(t *testing.T) {
	t.Parallel()

	reply := rpcEnvelope(t, json.RawMessage(`{
		"block_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"status": "OK",
		"untrusted": false
	}`))

	client, call := serveDaemon(t, reply)
	got, err := client.SubmitBlock(context.Background(), daemon.SubmitBlockRequest{"0c0c"})
	require.NoError(t, err)

	assert.Equal(t, "submit_block", call.Method)
	assert.JSONEq(t, `["0c0c"]`, string(call.Params))
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", got.BlockID)
	assert.Equal(t, "OK", got.Status)
}

func TestPruneBlockchain(t *testing.T) {
	t.Parallel()

	reply := rpcEnvelope(t, json.RawMessage(`{
		"pruned": true,
		"pruning_seed": 387,
		"status": "OK",
		"untrusted": false
	}`))

	client, call := serveDaemon(t, reply)
	got, err := client.PruneBlockchain(context.Background(), daemon.PruneBlockchainRequestParameters{
		Check: true,
	})
	require.NoError(t, err)

	assert.Equal(t, "prune_blockchain", call.Method)
	assert.JSONEq(t, `{"check":true}`, string(call.Params))
	assert.True(t, got.Pruned)
	assert.Equal(t, uint64(387), got.PruningSeed)
	assert.Equal(t, "OK", got.Status)
}

func TestCalcPow(t *testing.T) {
	t.Parallel()

	const hash = "d0402d6834e26fb94a9ce38c6424d27d2069896a9b8b1ce685d79936bca6e0a8"
	reply := rpcEnvelope(t, hash)

	client, call := serveDaemon(t, reply)
	got, err := client.CalcPow(context.Background(), daemon.CalcPowRequestParameters{
		MajorVersion: 14,
		Height:       2286447,
		BlockBlob:    "0c0c",
		SeedHash:     "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	})
	require.NoError(t, err)

	assert.Equal(t, "calc_pow", call.Method)
	assert.JSONEq(t, `{
		"major_version": 14,
		"height": 2286447,
		"block_blob": "0c0c",
		"seed_hash": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	}`, string(call.Params))
	assert.Equal(t, hash, got)
}

func TestFlushCache(t *testing.T) {
	t.Parallel()

	reply := rpcEnvelope(t, json.RawMessage(`{"status":"OK","untrusted":false}`))
	client, call := serveDaemon(t, reply)
	got, err := client.FlushCache(context.Background(), daemon.FlushCacheRequestParameters{
		BadBlocks: true,
	})
	require.NoError(t, err)

	assert.Equal(t, "flush_cache", call.Method)
	assert.JSONEq(t, `{"bad_blocks":true}`, string(call.Params))
	assert.Equal(t, "OK", got.Status)
	assert.False(t, got.Untrusted)
}

func TestAddAuxPow(t *testing.T) {
	t.Parallel()

	reply := rpcEnvelope(t, json.RawMessage(`{
		"aux_pow": [{
			"hash": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			"id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		}],
		"blockhashing_blob": "cc",
		"blocktemplate_blob": "dd",
		"merkle_root": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"merkle_tree_depth": 0,
		"status": "OK",
		"untrusted": false
	}`))

	client, call := serveDaemon(t, reply)
	got, err := client.AddAuxPow(context.Background(), daemon.AddAuxPowRequestParameters{
		BlocktemplateBlob: "ee",
		AuxPow: []daemon.AuxPow{{
			ID:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Hash: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		}},
	})
	require.NoError(t, err)

	assert.Equal(t, "add_aux_pow", call.Method)
	assert.JSONEq(t, `{
		"blocktemplate_blob": "ee",
		"aux_pow": [{
			"id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"hash": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		}]
	}`, string(call.Params))
	assert.Equal(t, "cc", got.BlockhashingBlob)
	assert.Equal(t, "dd", got.BlocktemplateBlob)
	assert.Equal(t, uint64(0), got.MerkleTreeDepth)
	require.Len(t, got.AuxPow, 1)
	assert.Equal(t, "OK", got.Status)
}
