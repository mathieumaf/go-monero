package daemon

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
)

// ParseTxpoolBacklog decodes the packed get_txpool_backlog blob. Monerod
// writes one record after another, each three little-endian uint64 values:
// weight, fee, and seconds spent in the pool.
func ParseTxpoolBacklog(blob []byte) ([]TxBacklogEntry, error) {
	words, err := parseLittleEndianUint64s(blob)
	if err != nil {
		return nil, fmt.Errorf("txpool backlog: %w", err)
	}

	const fieldsPerEntry = 3
	if len(words)%fieldsPerEntry != 0 {
		return nil, fmt.Errorf(
			"txpool backlog: %d fields is not a multiple of %d",
			len(words), fieldsPerEntry,
		)
	}

	entries := make([]TxBacklogEntry, len(words)/fieldsPerEntry)
	for i := range entries {
		base := i * fieldsPerEntry
		entries[i] = TxBacklogEntry{
			Weight:     words[base],
			Fee:        words[base+1],
			TimeInPool: words[base+2],
		}
	}

	return entries, nil
}

func parseLittleEndianUint64s(blob []byte) ([]uint64, error) {
	const word = 8
	if len(blob)%word != 0 {
		return nil, fmt.Errorf("pod blob: size %d is not a multiple of %d", len(blob), word)
	}

	out := make([]uint64, len(blob)/word)
	for i := range out {
		out[i] = binary.LittleEndian.Uint64(blob[i*word : (i+1)*word])
	}

	return out, nil
}

// UnmarshalJSON decodes the daemon payload and the packed backlog blob.
//
// The backlog string is read as bytes. Bytes 0x80-0xff stay as they appear
// in the JSON text, and \u00XX escapes below 0x20 become that byte. encoding/json
// rejects a raw NUL inside a string, which is how monerod 0.18 writes a
// non-empty backlog, so a response body has to escape those bytes before
// this decoder runs.
func (r *GetTxpoolBacklogResult) UnmarshalJSON(data []byte) error {
	var wire struct {
		Backlog   json.RawMessage `json:"backlog"`
		Status    string          `json:"status"`
		Untrusted bool            `json:"untrusted"`
		Credits   uint64          `json:"credits"`
		TopHash   string          `json:"top_hash"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return fmt.Errorf("unmarshal txpool backlog: %w", err)
	}

	blob, err := decodeJSONStringBytes(wire.Backlog)
	if err != nil {
		return err
	}

	entries, err := ParseTxpoolBacklog(blob)
	if err != nil {
		return err
	}

	r.Backlog = string(blob)
	r.Entries = entries
	r.Status = wire.Status
	r.Untrusted = wire.Untrusted
	r.Credits = wire.Credits
	r.TopHash = wire.TopHash

	return nil
}

// decodeJSONStringBytes reads a JSON string as a byte sequence. Raw bytes are
// copied unchanged. \u escapes follow JSON and are encoded as UTF-8, so
// \u0000 through \u007f are single bytes.
func decodeJSONStringBytes(raw json.RawMessage) ([]byte, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, nil
	}
	if raw[0] != '"' {
		return nil, fmt.Errorf("json string: expected a string")
	}

	out := make([]byte, 0, len(raw))
	for i := 1; i < len(raw); i++ {
		if raw[i] == '"' {
			return out, nil
		}
		if raw[i] != '\\' {
			out = append(out, raw[i])
			continue
		}

		decoded, next, err := decodeJSONEscape(raw, i)
		if err != nil {
			return nil, err
		}
		out = append(out, decoded...)
		i = next
	}

	return nil, fmt.Errorf("json string: unterminated string")
}

func decodeJSONEscape(raw []byte, slash int) ([]byte, int, error) {
	if slash+1 >= len(raw) {
		return nil, 0, fmt.Errorf("json string: truncated escape")
	}

	if escaped, ok := jsonStringEscapes[raw[slash+1]]; ok {
		return []byte{escaped}, slash + 1, nil
	}
	if raw[slash+1] != 'u' {
		return nil, 0, fmt.Errorf("json string: unknown escape")
	}
	if slash+5 >= len(raw) {
		return nil, 0, fmt.Errorf("json string: truncated unicode escape")
	}

	value, err := strconv.ParseUint(string(raw[slash+2:slash+6]), 16, 32)
	if err != nil {
		return nil, 0, fmt.Errorf("json string: unicode escape: %w", err)
	}

	return []byte(string(rune(value))), slash + 5, nil
}

// jsonStringEscapes matches the single-character escapes monerod emits, plus
// the escapes required by JSON.
var jsonStringEscapes = map[byte]byte{
	'"':  '"',
	'\\': '\\',
	'/':  '/',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
	'v':  '\v',
}
