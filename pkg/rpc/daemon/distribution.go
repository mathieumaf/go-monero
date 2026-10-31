package daemon

import (
	"encoding/json"
	"fmt"
)

// DistributionPayload is the raw JSON of one output distribution. Monerod
// sends a JSON array of counts when binary is false, and a string of packed
// little-endian uint64 counts when binary is true and compress is false.
type DistributionPayload []byte

// MarshalJSON returns the original JSON value.
func (p DistributionPayload) MarshalJSON() ([]byte, error) {
	if len(p) == 0 {
		return []byte("null"), nil
	}

	return append([]byte(nil), p...), nil
}

// UnmarshalJSON stores the raw JSON value.
func (p *DistributionPayload) UnmarshalJSON(data []byte) error {
	*p = append([]byte(nil), data...)
	return nil
}

// Counts returns the per-block output counts.
//
// A compressed payload is reported as an error. Request compress=false, which
// is the zero value of GetOutputDistributionRequestParameters, to receive a
// form Counts can read.
func (d OutputDistribution) Counts() ([]uint64, error) {
	if len(d.Distribution) == 0 || string(d.Distribution) == "null" {
		if d.CompressedData != "" {
			return nil, fmt.Errorf("compressed output distribution is not decoded")
		}

		return []uint64{}, nil
	}

	switch d.Distribution[0] {
	case '[':
		var counts []uint64
		if err := json.Unmarshal(d.Distribution, &counts); err != nil {
			return nil, fmt.Errorf("distribution: %w", err)
		}

		return counts, nil
	case '"':
		raw, err := decodeJSONStringBytes(json.RawMessage(d.Distribution))
		if err != nil {
			return nil, fmt.Errorf("distribution blob: %w", err)
		}

		counts, err := parseLittleEndianUint64s(raw)
		if err != nil {
			return nil, fmt.Errorf("distribution blob: %w", err)
		}

		return counts, nil
	default:
		return nil, fmt.Errorf("distribution: unexpected JSON value")
	}
}
