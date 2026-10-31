package daemon

import "fmt"

// AddressType is the network family of a peer address. Values match
// monerod's epee::net_utils::address_type and are stable across releases.
type AddressType uint8

const (
	// AddressTypeInvalid is an unset or unrecognized address family.
	AddressTypeInvalid AddressType = 0
	// AddressTypeIPv4 is an IPv4 peer.
	AddressTypeIPv4 AddressType = 1
	// AddressTypeIPv6 is an IPv6 peer.
	AddressTypeIPv6 AddressType = 2
	// AddressTypeI2P is an I2P peer.
	AddressTypeI2P AddressType = 3
	// AddressTypeTor is a Tor onion peer.
	AddressTypeTor AddressType = 4
)

// String returns the monerod name of the address family.
func (t AddressType) String() string {
	switch t {
	case AddressTypeInvalid:
		return "invalid"
	case AddressTypeIPv4:
		return "ipv4"
	case AddressTypeIPv6:
		return "ipv6"
	case AddressTypeI2P:
		return "i2p"
	case AddressTypeTor:
		return "tor"
	default:
		return fmt.Sprintf("address_type(%d)", uint8(t))
	}
}
