package daemon_test

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieumaf/go-monero/pkg/rpc"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

// Decoding tests load anonymized monerod 0.18.x responses from testdata.
// IPv4 addresses are RFC 5737 documentation ranges or loopback.
// Hashes, peer ids, and connection ids are synthetic.

//go:embed testdata/*.json
var fixtureFS embed.FS

const jsonRPCPath = "/json_rpc"

const (
	fixtureTipHash   = "7777777777777777777777777777777777777777777777777777777777777777"
	fixtureOldHash   = "1111111111111111111111111111111111111111111111111111111111111111"
	fixturePrevHash  = "2222222222222222222222222222222222222222222222222222222222222222"
	fixtureMinerHash = "3333333333333333333333333333333333333333333333333333333333333333"
	fixtureTxHashA   = "4444444444444444444444444444444444444444444444444444444444444444"
	fixtureTxHashB   = "5555555555555555555555555555555555555555555555555555555555555555"
)

// Fixtures contain the keys monerod 0.18.x returns for fields this package
// maps. strictDecode rejects any other key, which is how a refreshed capture
// reports drift.
func TestDecodeDaemonFixtures(t *testing.T) {
	for _, tc := range decodeCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			body, err := fixtureFS.ReadFile("testdata/" + tc.file)
			require.NoError(t, err)

			got, rec := callFixture(t, body, tc.call)
			assert.Equal(t, http.MethodGet, rec.method)
			assert.Equal(t, tc.path, rec.path)

			if tc.method != "" {
				var req rpc.RequestEnvelope
				require.NoError(t, json.Unmarshal(rec.body, &req))
				assert.Equal(t, tc.method, req.Method)
				assert.Equal(t, "2.0", req.JSONRPC)
			}

			strict := tc.newResult()
			require.NoError(t, strictDecode(body, tc.method != "", strict))
			assert.Equal(t, strict, got)

			tc.check(t, got)
		})
	}
}

// strictDecode reports an unmapped key. The client ignores that key and
// still returns the mapped fee fields.
func TestDisallowUnknownFieldsRejectsUnmappedKey(t *testing.T) {
	t.Parallel()

	body, err := fixtureFS.ReadFile("testdata/get_fee_estimate.json")
	require.NoError(t, err)

	mutated := withResultField(t, body, "not_a_monerod_field", true)

	got, _ := callFixture(t, mutated, func(ctx context.Context, c *daemon.Client) (any, error) {
		return c.GetFeeEstimate(ctx, 0)
	})
	fee := mustAs[daemon.GetFeeEstimateResult](t, got)
	assert.Equal(t, 7874, fee.Fee)
	assert.Equal(t, 10000, fee.QuantizationMask)

	err = strictDecode(mutated, true, &daemon.GetFeeEstimateResult{})
	require.Error(t, err)
	assert.ErrorContains(t, err, "not_a_monerod_field")
}

func TestEveryFixtureHasADecodeCase(t *testing.T) {
	t.Parallel()

	entries, err := fixtureFS.ReadDir("testdata")
	require.NoError(t, err)

	covered := map[string]bool{}
	for _, tc := range decodeCases() {
		covered[tc.file] = true
	}

	for _, entry := range entries {
		assert.Truef(t, covered[entry.Name()], "missing decode case for %s", entry.Name())
	}
}

func TestFixturesUseDocumentationAddresses(t *testing.T) {
	t.Parallel()

	entries, err := fixtureFS.ReadDir("testdata")
	require.NoError(t, err)

	for _, entry := range entries {
		body, err := fixtureFS.ReadFile("testdata/" + entry.Name())
		require.NoError(t, err)
		assert.NotContains(t, string(body), ".onion")
		assert.NotContains(t, string(body), ".i2p")

		var payload any
		require.NoError(t, json.Unmarshal(body, &payload))
		walkStrings(payload, func(s string) {
			ip, ok := addressIP(s)
			if !ok {
				return
			}
			assert.Truef(t, documentationIPv4(ip), "%s has non-documentation address %s", entry.Name(), s)
		})
	}
}

type daemonCall func(context.Context, *daemon.Client) (any, error)

type decodeCase struct {
	name      string
	file      string
	path      string
	method    string
	call      daemonCall
	newResult func() any
	check     func(*testing.T, any)
}

func decodeCases() []decodeCase {
	return []decodeCase{
		{
			name:   "get_info",
			file:   "get_info.json",
			path:   jsonRPCPath,
			method: "get_info",
			call: func(ctx context.Context, c *daemon.Client) (any, error) {
				return c.GetInfo(ctx)
			},
			newResult: func() any { return &daemon.GetInfoResult{} },
			check:     checkGetInfo,
		},
		{
			name:   "get_connections",
			file:   "get_connections.json",
			path:   jsonRPCPath,
			method: "get_connections",
			call: func(ctx context.Context, c *daemon.Client) (any, error) {
				return c.GetConnections(ctx)
			},
			newResult: func() any { return &daemon.GetConnectionsResult{} },
			check:     checkGetConnections,
		},
		{
			name: "get_peer_list",
			file: "get_peer_list.json",
			path: "/get_peer_list",
			call: func(ctx context.Context, c *daemon.Client) (any, error) {
				return c.GetPeerList(ctx)
			},
			newResult: func() any { return &daemon.GetPeerListResult{} },
			check:     checkGetPeerList,
		},
		{
			name: "get_net_stats",
			file: "get_net_stats.json",
			path: "/get_net_stats",
			call: func(ctx context.Context, c *daemon.Client) (any, error) {
				return c.GetNetStats(ctx)
			},
			newResult: func() any { return &daemon.GetNetStatsResult{} },
			check:     checkGetNetStats,
		},
		{
			name:   "get_fee_estimate",
			file:   "get_fee_estimate.json",
			path:   jsonRPCPath,
			method: "get_fee_estimate",
			call: func(ctx context.Context, c *daemon.Client) (any, error) {
				return c.GetFeeEstimate(ctx, 0)
			},
			newResult: func() any { return &daemon.GetFeeEstimateResult{} },
			check:     checkGetFeeEstimate,
		},
		{
			name: "get_transaction_pool_stats",
			file: "get_transaction_pool_stats.json",
			path: "/get_transaction_pool_stats",
			call: func(ctx context.Context, c *daemon.Client) (any, error) {
				return c.GetTransactionPoolStats(ctx)
			},
			newResult: func() any { return &daemon.GetTransactionPoolStatsResult{} },
			check:     checkGetTransactionPoolStats,
		},
		{
			name:   "get_last_block_header",
			file:   "get_last_block_header.json",
			path:   jsonRPCPath,
			method: "get_last_block_header",
			call: func(ctx context.Context, c *daemon.Client) (any, error) {
				return c.GetLastBlockHeader(ctx)
			},
			newResult: func() any { return &daemon.GetLastBlockHeaderResult{} },
			check:     checkGetLastBlockHeader,
		},
		{
			name:   "get_block",
			file:   "get_block.json",
			path:   jsonRPCPath,
			method: "get_block",
			call: func(ctx context.Context, c *daemon.Client) (any, error) {
				return c.GetBlock(ctx, daemon.GetBlockRequestParameters{Height: 3201465})
			},
			newResult: func() any { return &daemon.GetBlockResult{} },
			check:     checkGetBlock,
		},
		{
			name:   "hard_fork_info",
			file:   "hard_fork_info.json",
			path:   jsonRPCPath,
			method: "hard_fork_info",
			call: func(ctx context.Context, c *daemon.Client) (any, error) {
				return c.HardForkInfo(ctx)
			},
			newResult: func() any { return &daemon.HardForkInfoResult{} },
			check:     checkHardForkInfo,
		},
		{
			name:   "get_bans",
			file:   "get_bans.json",
			path:   jsonRPCPath,
			method: "get_bans",
			call: func(ctx context.Context, c *daemon.Client) (any, error) {
				return c.GetBans(ctx)
			},
			newResult: func() any { return &daemon.GetBansResult{} },
			check:     checkGetBans,
		},
		{
			name:   "sync_info",
			file:   "sync_info.json",
			path:   jsonRPCPath,
			method: "sync_info",
			call: func(ctx context.Context, c *daemon.Client) (any, error) {
				return c.SyncInfo(ctx)
			},
			newResult: func() any { return &daemon.SyncInfoResult{} },
			check:     checkSyncInfo,
		},
	}
}

func checkGetInfo(t *testing.T, got any) {
	t.Helper()

	info := mustAs[daemon.GetInfoResult](t, got)
	assert.Equal(t, "0.18.3.4-release", info.Version)
	assert.Equal(t, "mainnet", info.Nettype)
	assert.True(t, info.Mainnet)
	assert.False(t, info.Testnet)
	assert.False(t, info.Stagenet)
	assert.True(t, info.Synchronized)
	assert.Equal(t, uint64(3201506), info.Height)
	assert.Equal(t, uint64(3201506), info.HeightWithoutBootstrap)
	assert.Equal(t, uint64(300000000000), info.Difficulty)
	assert.Equal(t, int64(400000000000000000), info.CumulativeDifficulty)
	assert.Equal(t, "0x45d964b800", info.WideDifficulty)
	assert.Equal(t, "0x58d15e176280000", info.WideCumulativeDifficulty)
	assert.Equal(t, fixtureTipHash, info.TopBlockHash)
	assert.Equal(t, uint64(12), info.TxPoolSize)
	assert.Equal(t, uint64(14500021), info.TxCount)
	assert.Equal(t, uint(4), info.IncomingConnectionsCount)
	assert.Equal(t, uint(8), info.OutgoingConnectionsCount)
	assert.Equal(t, "OK", info.Status)
	assert.False(t, info.Untrusted)
}

func checkGetConnections(t *testing.T, got any) {
	t.Helper()

	res := mustAs[daemon.GetConnectionsResult](t, got)
	require.Len(t, res.Connections, 3)

	in := res.Connections[0]
	assert.Equal(t, "192.0.2.10:18080", in.Address)
	assert.Equal(t, "192.0.2.10", in.Host)
	assert.Equal(t, "192.0.2.10", in.IP)
	assert.True(t, in.Incoming)
	assert.False(t, in.Localhost)
	assert.False(t, in.LocalIP)
	assert.Equal(t, "d1d1d1d1d1d1d1d1", in.PeerID)
	assert.Equal(t, uint64(3201506), in.Height)
	assert.Equal(t, "normal", in.State)
	assert.Equal(t, "18080", in.Port)
	assert.Equal(t, uint64(1), in.SupportFlags)
	assert.Equal(t, "c1c1c1c1c1c1c1c1c1c1c1c1c1c1c1c1", in.ConnectionID)
	assert.Equal(t, uint64(4096), in.AvgDownload)
	assert.Equal(t, uint64(3600), in.LiveTime)

	out := res.Connections[1]
	assert.Equal(t, "198.51.100.20:18080", out.Address)
	assert.False(t, out.Incoming)
	assert.Equal(t, uint64(3201504), out.Height)
	assert.Equal(t, "d2d2d2d2d2d2d2d2", out.PeerID)
	assert.Equal(t, uint64(8192), out.AvgDownload)

	local := res.Connections[2]
	assert.Equal(t, "127.0.0.1:18080", local.Address)
	assert.True(t, local.Localhost)
	assert.True(t, local.LocalIP)
	assert.False(t, local.Incoming)
	assert.Equal(t, "OK", res.Status)
	assert.False(t, res.Untrusted)
}

func checkGetPeerList(t *testing.T, got any) {
	t.Helper()

	res := mustAs[daemon.GetPeerListResult](t, got)
	require.Len(t, res.WhiteList, 1)
	require.Len(t, res.GrayList, 1)

	white := res.WhiteList[0]
	assert.Equal(t, "192.0.2.10", white.Host)
	assert.Equal(t, uint64(0xA1A1A1A1A1A1A1A1), white.ID)
	assert.Equal(t, uint32(167903424), white.IP, "192.0.2.10")
	assert.Equal(t, uint16(18080), white.Port)
	assert.Equal(t, uint16(18081), white.RPCPort)
	assert.Equal(t, int64(1737990000), white.LastSeen)

	gray := res.GrayList[0]
	assert.Equal(t, "198.51.100.20", gray.Host)
	assert.Equal(t, uint64(0xB2B2B2B2B2B2B2B2), gray.ID)
	assert.Equal(t, uint32(342111174), gray.IP, "198.51.100.20")
	assert.Equal(t, uint16(18080), gray.Port)
	assert.Equal(t, uint32(387), gray.PruningSeed)
	assert.Equal(t, int64(1737000000), gray.LastSeen)
	assert.Equal(t, "OK", res.Status)
	assert.False(t, res.Untrusted)
}

func checkGetNetStats(t *testing.T, got any) {
	t.Helper()

	res := mustAs[daemon.GetNetStatsResult](t, got)
	assert.Equal(t, int64(1737000000), res.StartTime)
	assert.Equal(t, uint64(3743474809), res.TotalBytesIn)
	assert.Equal(t, uint64(5932012405), res.TotalBytesOut)
	assert.Equal(t, uint64(2130592), res.TotalPacketsIn)
	assert.Equal(t, uint64(1010674), res.TotalPacketsOut)
	assert.Equal(t, "OK", res.Status)
	assert.False(t, res.Untrusted)
}

func checkGetFeeEstimate(t *testing.T, got any) {
	t.Helper()

	res := mustAs[daemon.GetFeeEstimateResult](t, got)
	assert.Equal(t, 7874, res.Fee)
	assert.Equal(t, 10000, res.QuantizationMask)
	assert.Equal(t, uint64(0), res.Credits)
	assert.Empty(t, res.TopHash)
	assert.Equal(t, "OK", res.Status)
	assert.False(t, res.Untrusted)
}

func checkGetTransactionPoolStats(t *testing.T, got any) {
	t.Helper()

	res := mustAs[daemon.GetTransactionPoolStatsResult](t, got)
	stats := res.PoolStats
	assert.Equal(t, uint64(14325), stats.BytesMax)
	assert.Equal(t, uint64(1450), stats.BytesMin)
	assert.Equal(t, uint64(2214), stats.BytesMed)
	assert.Equal(t, uint64(28416), stats.BytesTotal)
	assert.Equal(t, uint64(532080000), stats.FeeTotal)
	assert.Equal(t, uint64(12), stats.TxsTotal)
	assert.Equal(t, int64(1737995000), stats.Oldest)
	assert.Equal(t, uint64(1), stats.Num10M)
	assert.Equal(t, uint64(297), stats.Histo98Pc)
	assert.Equal(t, uint64(0), stats.NumFailing)
	require.Len(t, stats.Histo, 3)
	assert.Equal(t, uint64(1600), stats.Histo[0].Bytes)
	assert.Equal(t, uint64(6), stats.Histo[0].Txs)
	assert.Equal(t, uint64(2), stats.Histo[2].Txs)
	assert.Equal(t, "OK", res.Status)
	assert.False(t, res.Untrusted)
}

func checkGetLastBlockHeader(t *testing.T, got any) {
	t.Helper()

	res := mustAs[daemon.GetLastBlockHeaderResult](t, got)
	header := res.BlockHeader
	assert.Equal(t, fixtureTipHash, header.Hash)
	assert.Equal(t, uint64(3201505), header.Height)
	assert.Equal(t, uint64(0), header.Depth)
	assert.Equal(t, uint(16), header.MajorVersion)
	assert.Equal(t, uint(16), header.MinorVersion)
	assert.Equal(t, uint64(300000000000), header.Difficulty)
	assert.Equal(t, "0x45d964b800", header.WideDifficulty)
	assert.Equal(t, uint64(618000000000), header.Reward)
	assert.Equal(t, uint64(188625235), header.Nonce)
	assert.Equal(t, uint(3), header.NumTxes)
	assert.Equal(t, uint64(5500), header.BlockWeight)
	assert.False(t, header.OrphanStatus)
	assert.Empty(t, header.PowHash)
	assert.Equal(t, fixtureMinerHash, header.MinerTxHash)
	assert.Equal(t, "OK", res.Status)
	assert.False(t, res.Untrusted)
}

func checkGetBlock(t *testing.T, got any) {
	t.Helper()

	block := mustAs[daemon.GetBlockResult](t, got)
	assert.Equal(t, strings.Repeat("aa55", 64), block.Blob)
	assert.Equal(t, fixtureOldHash, block.BlockHeader.Hash)
	assert.Equal(t, uint64(3201465), block.BlockHeader.Height)
	assert.Equal(t, uint64(40), block.BlockHeader.Depth)
	assert.Equal(t, uint64(128), block.BlockHeader.BlockSize)
	assert.Equal(t, fixtureMinerHash, block.MinerTxHash)
	assert.Equal(t, fixtureMinerHash, block.BlockHeader.MinerTxHash)
	assert.Equal(t, fixturePrevHash, block.BlockHeader.PrevHash)
	assert.Equal(t, uint64(249602367), block.BlockHeader.Nonce)
	assert.Equal(t, "0x45622f2400", block.BlockHeader.WideDifficulty)
	assert.Equal(t, "OK", block.Status)
	assert.False(t, block.Untrusted)

	inner, err := block.InnerJSON()
	require.NoError(t, err)
	assert.Equal(t, uint(16), inner.MajorVersion)
	assert.Equal(t, uint(16), inner.MinorVersion)
	assert.Equal(t, fixturePrevHash, inner.PrevID)
	assert.Equal(t, 2, inner.MinerTx.Version)
	assert.Equal(t, 3201525, inner.MinerTx.UnlockTime)
	require.Len(t, inner.MinerTx.Vin, 1)
	assert.Equal(t, 3201465, inner.MinerTx.Vin[0].Gen.Height)
	require.Len(t, inner.MinerTx.Vout, 1)
	assert.Equal(t, uint64(600061380000), inner.MinerOutputs())
	assert.Equal(t, []string{fixtureTxHashA, fixtureTxHashB}, inner.TxHashes)
}

func checkHardForkInfo(t *testing.T, got any) {
	t.Helper()

	res := mustAs[daemon.HardForkInfoResult](t, got)
	assert.Equal(t, 16, res.Version)
	assert.True(t, res.Enabled)
	assert.Equal(t, 2, res.State)
	assert.Equal(t, 10080, res.Votes)
	assert.Equal(t, 10080, res.Window)
	assert.Equal(t, 16, res.Voting)
	assert.Equal(t, 2689608, res.EarliestHeight)
	assert.Equal(t, 0, res.Threshold)
	assert.Equal(t, "OK", res.Status)
	assert.False(t, res.Untrusted)
}

func checkGetBans(t *testing.T, got any) {
	t.Helper()

	res := mustAs[daemon.GetBansResult](t, got)
	require.Len(t, res.Bans, 1)
	assert.Equal(t, "203.0.113.30", res.Bans[0].Host)
	assert.Equal(t, 510722251, res.Bans[0].IP, "203.0.113.30")
	assert.Equal(t, uint(3600), res.Bans[0].Seconds)
	assert.Equal(t, "OK", res.Status)
	assert.False(t, res.Untrusted)
}

func checkSyncInfo(t *testing.T, got any) {
	t.Helper()

	res := mustAs[daemon.SyncInfoResult](t, got)
	assert.Equal(t, uint64(3201506), res.Height)
	assert.Equal(t, uint64(3201500), res.TargetHeight)
	assert.Equal(t, "[]", res.Overview)
	assert.Equal(t, uint64(0), res.NextNeededPruningSeed)
	require.Len(t, res.Peers, 1)

	info := res.Peers[0].Info
	assert.Equal(t, "192.0.2.40:18080", info.Address)
	assert.Equal(t, uint64(1), info.AddressType)
	assert.Equal(t, "192.0.2.40", info.Host)
	assert.Equal(t, "192.0.2.40", info.IP)
	assert.Equal(t, "e1e1e1e1e1e1e1e1", info.PeerID)
	assert.Equal(t, uint64(387), info.PruningSeed)
	assert.Equal(t, uint64(18089), info.RPCPort)
	assert.Equal(t, uint64(0), info.RPCCreditsPerHash)
	assert.Equal(t, "normal", info.State)
	assert.False(t, info.Incoming)
	assert.Equal(t, 1, info.SupportFlags)
	assert.Equal(t, "OK", res.Status)
	assert.False(t, res.Untrusted)
}

type httpCapture struct {
	method string
	path   string
	body   []byte
}

func callFixture(t *testing.T, body []byte, call daemonCall) (any, httpCapture) {
	t.Helper()

	var rec httpCapture

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.method = r.Method
		rec.path = r.URL.Path

		var err error
		rec.body, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err = w.Write(body); err != nil {
			return
		}
	}))
	t.Cleanup(srv.Close)

	client, err := rpc.NewClient(srv.URL, rpc.WithHTTPClient(srv.Client()))
	require.NoError(t, err)

	got, err := call(context.Background(), daemon.NewClient(client))
	require.NoError(t, err)

	return got, rec
}

func strictDecode(body []byte, jsonRPC bool, dest any) error {
	payload := body
	if jsonRPC {
		var env struct {
			ID      string          `json:"id"`
			JSONRPC string          `json:"jsonrpc"`
			Result  json.RawMessage `json:"result"`
		}

		dec := json.NewDecoder(bytes.NewReader(body))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&env); err != nil {
			return err
		}
		if err := rejectTrailing(dec); err != nil {
			return err
		}
		if len(env.Result) == 0 {
			return errors.New("jsonrpc envelope has an empty result")
		}
		payload = env.Result
	}

	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dest); err != nil {
		return err
	}

	return rejectTrailing(dec)
}

func rejectTrailing(dec *json.Decoder) error {
	var extra any
	if err := dec.Decode(&extra); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}

		return err
	}

	return errors.New("trailing json after value")
}

func withResultField(t *testing.T, body []byte, key string, value any) []byte {
	t.Helper()

	var env map[string]any
	require.NoError(t, json.Unmarshal(body, &env))

	raw, ok := env["result"]
	require.True(t, ok)

	result, ok := raw.(map[string]any)
	require.True(t, ok)

	result[key] = value

	out, err := json.Marshal(env)
	require.NoError(t, err)

	return out
}

func mustAs[T any](t *testing.T, got any) *T {
	t.Helper()

	v, ok := got.(*T)
	require.Truef(t, ok, "got %T", got)

	return v
}

func walkStrings(v any, visit func(string)) {
	switch x := v.(type) {
	case map[string]any:
		for _, child := range x {
			walkStrings(child, visit)
		}
	case []any:
		for _, child := range x {
			walkStrings(child, visit)
		}
	case string:
		visit(x)
	}
}

func addressIP(s string) (net.IP, bool) {
	if ip := net.ParseIP(s); ip != nil {
		return ip, true
	}

	host, _, err := net.SplitHostPort(s)
	if err != nil {
		return nil, false
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return nil, false
	}

	return ip, true
}

var documentationNets = mustCIDRs(
	"127.0.0.0/8",
	"192.0.2.0/24",
	"198.51.100.0/24",
	"203.0.113.0/24",
)

func mustCIDRs(cidrs ...string) []*net.IPNet {
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(err)
		}

		nets = append(nets, network)
	}

	return nets
}

func documentationIPv4(ip net.IP) bool {
	v4 := ip.To4()
	if v4 == nil {
		return false
	}

	for _, network := range documentationNets {
		if network.Contains(v4) {
			return true
		}
	}

	return false
}
