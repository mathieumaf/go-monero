package daemon_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieumaf/go-monero/pkg/rpc"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

// Fixtures below follow monerod 0.18 JSON "other" responses. Hashes and
// URIs are synthetic stand-ins for the samples in the monerod RPC docs.

const (
	altBlocksFixture = `{
  "blks_hashes": [
    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
  ],
  "credits": 0,
  "status": "OK",
  "top_hash": "",
  "untrusted": false
}`

	poolHashesFixture = `{
  "credits": 0,
  "status": "OK",
  "top_hash": "",
  "tx_hashes": [
    "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
    "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
  ],
  "untrusted": false
}`

	keyImageFixture = `{
  "credits": 0,
  "spent_status": [0, 1, 2],
  "status": "OK",
  "top_hash": "",
  "untrusted": false
}`

	inPeersFixture = `{
  "in_peers": 3232235535,
  "status": "OK",
  "untrusted": false
}`

	outPeersFixture = `{
  "out_peers": 32,
  "status": "OK",
  "untrusted": false
}`

	updateFixture = `{
  "auto_uri": "https://downloads.example/cli/monero-linux-x64-v0.18.4.0.tar.bz2",
  "hash": "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
  "path": "/var/tmp/monero-update.tar.bz2",
  "status": "OK",
  "untrusted": false,
  "update": true,
  "user_uri": "https://downloads.example/cli/monero-linux-x64-v0.18.4.0.tar.bz2",
  "version": "0.18.4.0"
}`

	sendRawFixture = `{
  "double_spend": true,
  "fee_too_low": false,
  "invalid_input": false,
  "invalid_output": false,
  "low_mixin": false,
  "nonzero_unlock_time": false,
  "not_relayed": true,
  "overspend": false,
  "reason": "Not relayed",
  "sanity_check_failed": false,
  "status": "Failed",
  "too_big": false,
  "too_few_outputs": false,
  "tx_extra_too_big": false,
  "untrusted": false
}`

	popBlocksFixture = `{
  "height": 76482,
  "status": "OK",
  "untrusted": false
}`

	statusOnlyFixture = `{
  "status": "OK"
}`

	notMiningFixture = `{
  "status": "NOT MINING",
  "untrusted": false
}`
)

type capturedRequest struct {
	method  string
	path    string
	body    []byte
	readErr error
}

func newDaemonServer(
	t *testing.T, status int, response string,
) (*daemon.Client, *capturedRequest) {
	t.Helper()

	got := &capturedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method = r.Method
		got.path = r.URL.Path
		got.body, got.readErr = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, response)
	}))
	t.Cleanup(srv.Close)

	rpcClient, err := rpc.NewClient(srv.URL, rpc.WithHTTPClient(srv.Client()))
	require.NoError(t, err)

	return daemon.NewClient(rpcClient), got
}

func assertRequest(t *testing.T, got *capturedRequest, path string, params any) {
	t.Helper()

	require.NoError(t, got.readErr)
	assert.Equal(t, http.MethodGet, got.method)
	assert.Equal(t, path, got.path)

	if params == nil {
		assert.Empty(t, got.body)
		return
	}

	expected, err := json.Marshal(params)
	require.NoError(t, err)
	assert.JSONEq(t, string(expected), string(got.body))
}

func TestGetAltBlocksHashes(t *testing.T) {
	t.Parallel()

	client, got := newDaemonServer(t, http.StatusOK, altBlocksFixture)
	resp, err := client.GetAltBlocksHashes(context.Background())
	require.NoError(t, err)
	assertRequest(t, got, "/get_alt_blocks_hashes", nil)

	assert.Equal(t, "OK", resp.Status)
	assert.False(t, resp.Untrusted)
	assert.Equal(t, []string{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}, resp.BlksHashes)
}

func TestGetTransactionPoolHashes(t *testing.T) {
	t.Parallel()

	client, got := newDaemonServer(t, http.StatusOK, poolHashesFixture)
	resp, err := client.GetTransactionPoolHashes(context.Background())
	require.NoError(t, err)
	assertRequest(t, got, "/get_transaction_pool_hashes", nil)

	assert.Equal(t, "OK", resp.Status)
	assert.Equal(t, uint64(0), resp.Credits)
	assert.Empty(t, resp.TopHash)
	assert.Equal(t, []string{
		"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
	}, resp.TxHashes)
}

func TestIsKeyImageSpent(t *testing.T) {
	t.Parallel()

	params := daemon.IsKeyImageSpentRequestParameters{
		KeyImages: []string{
			"1111111111111111111111111111111111111111111111111111111111111111",
			"2222222222222222222222222222222222222222222222222222222222222222",
			"3333333333333333333333333333333333333333333333333333333333333333",
		},
	}
	client, got := newDaemonServer(t, http.StatusOK, keyImageFixture)
	resp, err := client.IsKeyImageSpent(context.Background(), params)
	require.NoError(t, err)
	assertRequest(t, got, "/is_key_image_spent", params)

	assert.Equal(t, []daemon.KeyImageSpentStatus{
		daemon.KeyImageUnspent,
		daemon.KeyImageSpentInBlockchain,
		daemon.KeyImageSpentInPool,
	}, resp.SpentStatus)
}

func TestInPeersReadDoesNotApplyLimit(t *testing.T) {
	t.Parallel()

	params := daemon.InPeersRequestParameters{Set: false, InPeers: 0}
	client, got := newDaemonServer(t, http.StatusOK, inPeersFixture)
	resp, err := client.InPeers(context.Background(), params)
	require.NoError(t, err)
	assertRequest(t, got, "/in_peers", params)

	assert.Equal(t, uint32(3232235535), resp.InPeers)
	assert.Equal(t, "OK", resp.Status)
	assert.False(t, resp.Untrusted)
}

func TestOutPeersSetEncodesLimit(t *testing.T) {
	t.Parallel()

	params := daemon.OutPeersRequestParameters{Set: true, OutPeers: 32}
	client, got := newDaemonServer(t, http.StatusOK, outPeersFixture)
	resp, err := client.OutPeers(context.Background(), params)
	require.NoError(t, err)
	assertRequest(t, got, "/out_peers", params)

	assert.Equal(t, uint32(32), resp.OutPeers)
	assert.JSONEq(t, `{"set":true,"out_peers":32}`, string(got.body))
}

func TestUpdateCheck(t *testing.T) {
	t.Parallel()

	params := daemon.UpdateRequestParameters{Command: daemon.UpdateCommandCheck}
	client, got := newDaemonServer(t, http.StatusOK, updateFixture)
	resp, err := client.Update(context.Background(), params)
	require.NoError(t, err)
	assertRequest(t, got, "/update", params)

	assert.True(t, resp.Update)
	assert.Equal(t, "0.18.4.0", resp.Version)
	assert.Equal(t, "https://downloads.example/cli/monero-linux-x64-v0.18.4.0.tar.bz2", resp.UserURI)
	assert.Equal(t, resp.UserURI, resp.AutoURI)
	assert.Equal(t, "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", resp.Hash)
	assert.Equal(t, "/var/tmp/monero-update.tar.bz2", resp.Path)
	assert.Equal(t, "OK", resp.Status)
}

func TestSendRawTransactionOmitsDefaultSanityCheck(t *testing.T) {
	t.Parallel()

	params := daemon.SendRawTransactionRequestParameters{TxAsHex: "de6a"}
	client, got := newDaemonServer(t, http.StatusOK, sendRawFixture)
	resp, err := client.SendRawTransaction(context.Background(), params)
	require.NoError(t, err)
	assertRequest(t, got, "/send_raw_transaction", params)

	assert.NotContains(t, string(got.body), "do_sanity_checks")
	assert.NotContains(t, string(got.body), "do_not_relay")
	assert.Equal(t, "Failed", resp.Status)
	assert.Equal(t, "Not relayed", resp.Reason)
	assert.True(t, resp.DoubleSpend)
	assert.True(t, resp.NotRelayed)
	assert.False(t, resp.FeeTooLow)
	assert.False(t, resp.SanityCheckFailed)
	assert.False(t, resp.Untrusted)
}

func TestSendRawTransactionCanDisableSanityChecks(t *testing.T) {
	t.Parallel()

	skip := false
	params := daemon.SendRawTransactionRequestParameters{
		TxAsHex:        "de6a",
		DoNotRelay:     true,
		DoSanityChecks: &skip,
	}
	client, got := newDaemonServer(t, http.StatusOK, statusOnlyFixture)
	_, err := client.SendRawTransaction(context.Background(), params)
	require.NoError(t, err)
	assertRequest(t, got, "/send_raw_transaction", params)
	assert.JSONEq(t, `{
		"tx_as_hex": "de6a",
		"do_not_relay": true,
		"do_sanity_checks": false
	}`, string(got.body))
}

func TestSaveBC(t *testing.T) {
	t.Parallel()

	client, got := newDaemonServer(t, http.StatusOK, statusOnlyFixture)
	resp, err := client.SaveBC(context.Background())
	require.NoError(t, err)
	assertRequest(t, got, "/save_bc", nil)
	assert.Equal(t, "OK", resp.Status)
	assert.False(t, resp.Untrusted)
}

func TestStopDaemon(t *testing.T) {
	t.Parallel()

	client, got := newDaemonServer(t, http.StatusOK, statusOnlyFixture)
	resp, err := client.StopDaemon(context.Background())
	require.NoError(t, err)
	assertRequest(t, got, "/stop_daemon", nil)
	assert.Equal(t, "OK", resp.Status)
}

func TestSetBootstrapDaemon(t *testing.T) {
	t.Parallel()

	params := daemon.SetBootstrapDaemonRequestParameters{
		Address:  "http://bootstrap.example:18081",
		Username: "rpc",
		Password: "secret",
		Proxy:    "",
	}
	client, got := newDaemonServer(t, http.StatusOK, statusOnlyFixture)
	resp, err := client.SetBootstrapDaemon(context.Background(), params)
	require.NoError(t, err)
	assertRequest(t, got, "/set_bootstrap_daemon", params)
	assert.Equal(t, "OK", resp.Status)
	assert.JSONEq(t, `{
		"address": "http://bootstrap.example:18081",
		"username": "rpc",
		"password": "secret",
		"proxy": ""
	}`, string(got.body))
}

func TestPopBlocksOmitsKeepTxsByDefault(t *testing.T) {
	t.Parallel()

	params := daemon.PopBlocksRequestParameters{Nblocks: 6}
	client, got := newDaemonServer(t, http.StatusOK, popBlocksFixture)
	resp, err := client.PopBlocks(context.Background(), params)
	require.NoError(t, err)
	assertRequest(t, got, "/pop_blocks", params)

	assert.NotContains(t, string(got.body), "keep_txs")
	assert.Equal(t, uint64(76482), resp.Height)
	assert.Equal(t, "OK", resp.Status)
}

func TestPopBlocksCanDropTransactions(t *testing.T) {
	t.Parallel()

	keep := false
	params := daemon.PopBlocksRequestParameters{Nblocks: 6, KeepTxs: &keep}
	client, got := newDaemonServer(t, http.StatusOK, popBlocksFixture)
	_, err := client.PopBlocks(context.Background(), params)
	require.NoError(t, err)
	assert.JSONEq(t, `{"nblocks":6,"keep_txs":false}`, string(got.body))
}

func TestSetLogHashRate(t *testing.T) {
	t.Parallel()

	params := daemon.SetLogHashRateRequestParameters{Visible: false}
	client, got := newDaemonServer(t, http.StatusOK, notMiningFixture)
	resp, err := client.SetLogHashRate(context.Background(), params)
	require.NoError(t, err)
	assertRequest(t, got, "/set_log_hash_rate", params)
	assert.JSONEq(t, `{"visible":false}`, string(got.body))
	assert.Equal(t, "NOT MINING", resp.Status)
	assert.False(t, resp.Untrusted)
}

func TestOtherEndpointHTTPError(t *testing.T) {
	t.Parallel()

	client, _ := newDaemonServer(t, http.StatusInternalServerError, `{"status":"BUSY"}`)
	_, err := client.GetAltBlocksHashes(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "raw request")
	assert.Contains(t, err.Error(), "non-2xx")
}
