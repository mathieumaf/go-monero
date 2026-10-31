package zmq_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mathieumaf/go-monero/pkg/zmq"
)

func TestJSONFromFrame(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		input         []byte
		expectedJSON  []byte
		expectedTopic zmq.Topic
		err           string
	}{
		{
			name:  "nil",
			input: nil,
			err:   "malformed",
		},

		{
			name:  "empty",
			input: []byte{},
			err:   "malformed",
		},

		{
			name:  "unknown-topic",
			input: []byte(`foobar:[{"foo":"bar"}]`),
			err:   "unknown topic",
		},

		{
			name:          "proper w/ known-topic",
			input:         []byte(`json-minimal-txpool_add:[{"foo":"bar"}]`),
			expectedTopic: zmq.TopicMinimalTxPoolAdd,
			expectedJSON:  []byte(`[{"foo":"bar"}]`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			aTopic, aJSON, err := zmq.JSONFromFrame(tc.input)
			if tc.err != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedTopic, aTopic)
			assert.Equal(t, tc.expectedJSON, aJSON)
		})
	}
}

// Anonymized monerod 0.18 json-full-miner_data frame. Hashes stand in for
// a real chain tip and mempool; the shape, including the 0x difficulty and
// a coin emission that does not fit in a float64 mantissa, matches the
// publisher in src/rpc/zmq_pub.cpp.
func TestFullMinerDataFrame(t *testing.T) {
	t.Parallel()

	const payload = `` +
		`{"major_version":16,"height":2731375,` +
		`"prev_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",` +
		`"seed_hash":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",` +
		`"difficulty":"0x48afae42de","median_weight":300000,` +
		`"already_generated_coins":18186022843595960691,` +
		`"tx_backlog":[` +
		`{"id":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc","weight":1535,"fee":30700000},` +
		`{"id":"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd","weight":2214,"fee":44280000}` +
		`]}`

	topic, gson, err := zmq.JSONFromFrame([]byte("json-full-miner_data:" + payload))
	require.NoError(t, err)
	assert.Equal(t, zmq.TopicFullMinerData, topic)
	assert.JSONEq(t, payload, string(gson))

	var got zmq.FullMinerData
	require.NoError(t, json.Unmarshal(gson, &got))
	assert.Equal(t, zmq.FullMinerData{
		MajorVersion:          16,
		Height:                2731375,
		PrevID:                "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		SeedHash:              "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Difficulty:            "0x48afae42de",
		MedianWeight:          300000,
		AlreadyGeneratedCoins: 18186022843595960691,
		TxBacklog: []zmq.MinerTxBacklog{
			{
				ID:     "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
				Weight: 1535,
				Fee:    30700000,
			},
			{
				ID:     "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
				Weight: 2214,
				Fee:    44280000,
			},
		},
	}, got)
}
