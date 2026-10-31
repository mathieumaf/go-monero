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

type capturedRPC struct {
	Method string
	Params json.RawMessage
}

func serveDaemon(t *testing.T, reply []byte) (*daemon.Client, *capturedRPC) {
	t.Helper()

	call := &capturedRPC{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var env struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err = json.Unmarshal(body, &env); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		call.Method = env.Method
		call.Params = append(json.RawMessage(nil), env.Params...)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(reply)
	}))
	t.Cleanup(srv.Close)

	rpcClient, err := rpc.NewClient(srv.URL, rpc.WithHTTPClient(srv.Client()))
	require.NoError(t, err)

	return daemon.NewClient(rpcClient), call
}

func rpcEnvelope(t *testing.T, result any) []byte {
	t.Helper()

	payload, err := json.Marshal(map[string]any{
		"id":      "0",
		"jsonrpc": "2.0",
		"result":  result,
	})
	require.NoError(t, err)

	return payload
}

func TestAddressTypeString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		value daemon.AddressType
		want  string
	}{
		{daemon.AddressTypeInvalid, "invalid"},
		{daemon.AddressTypeIPv4, "ipv4"},
		{daemon.AddressTypeIPv6, "ipv6"},
		{daemon.AddressTypeI2P, "i2p"},
		{daemon.AddressTypeTor, "tor"},
		{daemon.AddressType(9), "address_type(9)"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, tc.value.String())
	}
}

func TestGetFeeEstimate(t *testing.T) {
	t.Parallel()

	// Anonymized monerod 0.18 response: four fee tiers, slow to fastest.
	reply := rpcEnvelope(t, json.RawMessage(`{
		"credits": 0,
		"fee": 7874,
		"fees": [20000, 80000, 320000, 4000000],
		"quantization_mask": 10000,
		"status": "OK",
		"top_hash": "",
		"untrusted": false
	}`))

	client, call := serveDaemon(t, reply)
	got, err := client.GetFeeEstimate(context.Background(), 3)
	require.NoError(t, err)

	assert.Equal(t, "get_fee_estimate", call.Method)
	assert.JSONEq(t, `{"grace_blocks":3}`, string(call.Params))
	assert.Equal(t, 7874, got.Fee)
	assert.Equal(t, []uint64{20000, 80000, 320000, 4000000}, got.Fees)
	assert.Equal(t, uint64(20000), got.Fees[daemon.FeePrioritySlow])
	assert.Equal(t, uint64(80000), got.Fees[daemon.FeePriorityNormal])
	assert.Equal(t, uint64(320000), got.Fees[daemon.FeePriorityFast])
	assert.Equal(t, uint64(4000000), got.Fees[daemon.FeePriorityFastest])
	assert.Equal(t, 10000, got.QuantizationMask)
	assert.Equal(t, "OK", got.Status)
	assert.False(t, got.Untrusted)
}

func TestGetFeeEstimateLegacyResponse(t *testing.T) {
	t.Parallel()

	reply := rpcEnvelope(t, json.RawMessage(`{
		"fee": 7874,
		"quantization_mask": 10000,
		"status": "OK",
		"untrusted": false
	}`))

	client, _ := serveDaemon(t, reply)
	got, err := client.GetFeeEstimate(context.Background(), 0)
	require.NoError(t, err)
	assert.Equal(t, 7874, got.Fee)
	assert.Nil(t, got.Fees)
}

func TestGetConnections(t *testing.T) {
	t.Parallel()

	// Anonymized monerod 0.18 get_connections payload. Addresses use
	// documentation ranges and placeholder onion / I2P hosts.
	reply := rpcEnvelope(t, json.RawMessage(`{
		"connections": [
			{
				"address": "203.0.113.10:18080",
				"address_type": 1,
				"avg_download": 1200,
				"avg_upload": 40,
				"connection_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				"current_download": 100,
				"current_upload": 20,
				"height": 3321000,
				"host": "203.0.113.10",
				"incoming": false,
				"ip": "203.0.113.10",
				"live_time": 3600,
				"local_ip": false,
				"localhost": false,
				"peer_id": "0123456789abcdef",
				"port": "18080",
				"pruning_seed": 387,
				"recv_count": 4096,
				"recv_idle_time": 2,
				"rpc_credits_per_hash": 0,
				"rpc_port": 18081,
				"send_count": 2048,
				"send_idle_time": 3,
				"state": "normal",
				"support_flags": 1
			},
			{
				"address": "[2001:db8::10]:18080",
				"address_type": 2,
				"avg_download": 0,
				"avg_upload": 0,
				"connection_id": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				"current_download": 0,
				"current_upload": 0,
				"height": 3321001,
				"host": "2001:db8::10",
				"incoming": true,
				"ip": "2001:db8::10",
				"live_time": 12,
				"local_ip": false,
				"localhost": false,
				"peer_id": "fedcba9876543210",
				"port": "18080",
				"pruning_seed": 0,
				"recv_count": 100,
				"recv_idle_time": 1,
				"rpc_credits_per_hash": 8,
				"rpc_port": 18089,
				"send_count": 50,
				"send_idle_time": 1,
				"state": "normal",
				"support_flags": 1
			},
			{
				"address": "exampleexampleexampleexample.b32.i2p",
				"address_type": 3,
				"avg_download": 10,
				"avg_upload": 10,
				"connection_id": "cccccccccccccccccccccccccccccccc",
				"current_download": 1,
				"current_upload": 1,
				"height": 3320990,
				"host": "exampleexampleexampleexample.b32.i2p",
				"incoming": false,
				"ip": "exampleexampleexampleexample.b32.i2p",
				"live_time": 800,
				"local_ip": false,
				"localhost": false,
				"peer_id": "aabbccddeeff0011",
				"port": "18080",
				"pruning_seed": 0,
				"recv_count": 10,
				"recv_idle_time": 4,
				"rpc_credits_per_hash": 0,
				"rpc_port": 0,
				"send_count": 10,
				"send_idle_time": 4,
				"state": "normal",
				"support_flags": 1
			},
			{
				"address": "exampleonionaddress1234567890abcd.onion:18080",
				"address_type": 4,
				"avg_download": 5,
				"avg_upload": 5,
				"connection_id": "dddddddddddddddddddddddddddddddd",
				"current_download": 1,
				"current_upload": 1,
				"height": 3320500,
				"host": "exampleonionaddress1234567890abcd.onion",
				"incoming": true,
				"ip": "exampleonionaddress1234567890abcd.onion",
				"live_time": 90,
				"local_ip": false,
				"localhost": false,
				"peer_id": "0011223344556677",
				"port": "18080",
				"pruning_seed": 512,
				"recv_count": 32,
				"recv_idle_time": 6,
				"rpc_credits_per_hash": 0,
				"rpc_port": 0,
				"send_count": 32,
				"send_idle_time": 6,
				"state": "synchronizing",
				"support_flags": 0
			}
		],
		"status": "OK",
		"untrusted": false
	}`))

	client, call := serveDaemon(t, reply)
	got, err := client.GetConnections(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "get_connections", call.Method)
	assert.Empty(t, call.Params)
	require.Len(t, got.Connections, 4)
	assert.Equal(t, "OK", got.Status)

	ipv4 := got.Connections[0]
	assert.Equal(t, daemon.AddressTypeIPv4, ipv4.AddressType)
	assert.Equal(t, "ipv4", ipv4.AddressType.String())
	assert.Equal(t, uint64(18081), ipv4.RPCPort)
	assert.Equal(t, uint64(0), ipv4.RPCCreditsPerHash)
	assert.Equal(t, uint64(387), ipv4.PruningSeed)
	assert.Equal(t, "203.0.113.10:18080", ipv4.Address)
	assert.Equal(t, uint64(3321000), ipv4.Height)
	assert.False(t, ipv4.Incoming)

	assert.Equal(t, daemon.AddressTypeIPv6, got.Connections[1].AddressType)
	assert.Equal(t, uint64(18089), got.Connections[1].RPCPort)
	assert.Equal(t, uint64(8), got.Connections[1].RPCCreditsPerHash)
	assert.True(t, got.Connections[1].Incoming)

	assert.Equal(t, daemon.AddressTypeI2P, got.Connections[2].AddressType)
	assert.Equal(t, "i2p", got.Connections[2].AddressType.String())

	assert.Equal(t, daemon.AddressTypeTor, got.Connections[3].AddressType)
	assert.Equal(t, "tor", got.Connections[3].AddressType.String())
	assert.Equal(t, uint64(512), got.Connections[3].PruningSeed)
	assert.Equal(t, "synchronizing", got.Connections[3].State)
}

func TestGetConnectionsLegacyResponse(t *testing.T) {
	t.Parallel()

	reply := rpcEnvelope(t, json.RawMessage(`{
		"connections": [{
			"address": "203.0.113.20:18080",
			"avg_download": 1,
			"avg_upload": 1,
			"connection_id": "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
			"current_download": 0,
			"current_upload": 0,
			"height": 10,
			"host": "203.0.113.20",
			"incoming": false,
			"ip": "203.0.113.20",
			"live_time": 5,
			"local_ip": false,
			"localhost": false,
			"peer_id": "ff",
			"port": "18080",
			"recv_count": 1,
			"recv_idle_time": 1,
			"send_count": 1,
			"send_idle_time": 1,
			"state": "normal",
			"support_flags": 1
		}],
		"status": "OK",
		"untrusted": false
	}`))

	client, _ := serveDaemon(t, reply)
	got, err := client.GetConnections(context.Background())
	require.NoError(t, err)
	require.Len(t, got.Connections, 1)
	assert.Equal(t, daemon.AddressTypeInvalid, got.Connections[0].AddressType)
	assert.Equal(t, uint64(0), got.Connections[0].RPCPort)
	assert.Equal(t, uint64(0), got.Connections[0].PruningSeed)
	assert.Equal(t, uint64(1), got.Connections[0].SupportFlags)
}

func TestSyncInfoUsesConnectionInfo(t *testing.T) {
	t.Parallel()

	reply := rpcEnvelope(t, json.RawMessage(`{
		"credits": 0,
		"height": 3321000,
		"next_needed_pruning_seed": 0,
		"overview": "",
		"peers": [{
			"info": {
				"address": "exampleonionaddress1234567890abcd.onion:18080",
				"address_type": 4,
				"avg_download": 1,
				"avg_upload": 2,
				"connection_id": "ffffffffffffffffffffffffffffffff",
				"current_download": 0,
				"current_upload": 0,
				"height": 3321000,
				"host": "exampleonionaddress1234567890abcd.onion",
				"incoming": false,
				"ip": "exampleonionaddress1234567890abcd.onion",
				"live_time": 15,
				"local_ip": false,
				"localhost": false,
				"peer_id": "abc",
				"port": "18080",
				"pruning_seed": 9,
				"recv_count": 3,
				"recv_idle_time": 1,
				"rpc_credits_per_hash": 4,
				"rpc_port": 18081,
				"send_count": 4,
				"send_idle_time": 1,
				"state": "normal",
				"support_flags": 1
			}
		}],
		"status": "OK",
		"target_height": 3321000,
		"top_hash": "",
		"untrusted": false
	}`))

	client, call := serveDaemon(t, reply)
	got, err := client.SyncInfo(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "sync_info", call.Method)
	require.Len(t, got.Peers, 1)
	info := got.Peers[0].Info
	assert.Equal(t, daemon.AddressTypeTor, info.AddressType)
	assert.Equal(t, uint64(18081), info.RPCPort)
	assert.Equal(t, uint64(4), info.RPCCreditsPerHash)
	assert.Equal(t, uint64(9), info.PruningSeed)
	assert.Equal(t, uint64(1), info.SupportFlags)
}
