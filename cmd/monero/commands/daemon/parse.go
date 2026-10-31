package daemon

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

func parseUint64List(values []string) ([]uint64, error) {
	out := make([]uint64, 0, len(values))
	for _, value := range values {
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("amount %q: %w", value, err)
		}
		out = append(out, parsed)
	}

	return out, nil
}

func parseAuxPow(values []string) ([]daemon.AuxPow, error) {
	out := make([]daemon.AuxPow, 0, len(values))
	for _, value := range values {
		id, hash, ok := strings.Cut(value, ":")
		if !ok || id == "" || hash == "" {
			return nil, fmt.Errorf("aux pow %q must be id:hash", value)
		}
		out = append(out, daemon.AuxPow{ID: id, Hash: hash})
	}

	return out, nil
}
