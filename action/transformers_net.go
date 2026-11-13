package action

import (
	"fmt"
	"math/big"
	"net"
	"net/netip"
	"strconv"
	"strings"
)

var ipv4Action = New(Definition[[]byte, []byte]{
	Doc:          "Convert an IPv4 address between dotted, decimal, 0x hex and 0b binary forms, e.g. 192.168.1.1 to 3232235777",
	Names:        []string{"ipv4"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: TextFormat,
	Func: func(a Action, in []byte) ([]byte, error) {
		s := strings.TrimSpace(string(in))
		addr, dotted, err := parseIPv4(s)
		if err != nil {
			return nil, err
		}
		if dotted {
			return []byte(strconv.FormatUint(uint64(addr), 10)), nil
		}
		return []byte(netip.AddrFrom4([4]byte{
			byte(addr >> 24), byte(addr >> 16), byte(addr >> 8), byte(addr),
		}).String()), nil
	},
})

// parseIPv4 accepts dotted, decimal, 0x and 0b forms; dotted reports whether
// the input was the dotted form, so the caller can pick the opposite output.
func parseIPv4(s string) (addr uint32, dotted bool, err error) {
	if ip, err := netip.ParseAddr(s); err == nil && ip.Is4() {
		b := ip.As4()
		return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]), true, nil
	}

	base := 10
	lower := strings.ToLower(s)
	switch {
	case strings.HasPrefix(lower, "0x"):
		base, s = 16, s[2:]
	case strings.HasPrefix(lower, "0b"):
		base, s = 2, s[2:]
	}
	v, err := strconv.ParseUint(s, base, 32)
	if err != nil {
		return 0, false, fmt.Errorf("%q is not an IPv4 address or a 32-bit number", s)
	}
	return uint32(v), false, nil
}

var subnetAction = New(Definition[[]byte, map[string]any]{
	Doc:          "Network details for a CIDR or an address: network, netmask, wildcard, broadcast, host range and host count",
	Names:        []string{"subnet"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: DictFormat,
	Func: func(a Action, in []byte) (map[string]any, error) {
		s := strings.TrimSpace(string(in))
		if !strings.Contains(s, "/") {
			addr, err := netip.ParseAddr(s)
			if err != nil {
				return nil, fmt.Errorf("%q is not an address or a CIDR", s)
			}
			bits := 32
			if addr.Is6() && !addr.Is4() {
				bits = 128
			}
			s = fmt.Sprintf("%s/%d", addr, bits)
		}
		prefix, err := netip.ParsePrefix(s)
		if err != nil {
			return nil, fmt.Errorf("%q is not a valid CIDR: %w", s, err)
		}
		return subnetInfo(prefix), nil
	},
})

// subnetInfo builds the report dict for a prefix.
func subnetInfo(prefix netip.Prefix) map[string]any {
	prefix = prefix.Masked()
	addr := prefix.Addr()
	bits := 128
	if addr.Is4() {
		bits = 32
	}
	hostBits := bits - prefix.Bits()
	network := prefix.Addr()
	last := lastAddr(prefix)

	hosts := new(big.Int).Lsh(big.NewInt(1), uint(hostBits))
	firstHost, lastHost := network, last
	if hostBits >= 2 {
		firstHost = network.Next()
		lastHost = last.Prev()
		hosts.Sub(hosts, big.NewInt(2))
	}

	info := map[string]any{
		"cidr":    prefix.String(),
		"version": addr.BitLen(),
		"network": network.String(),
		"netmask": netmask(prefix),
		"first":   firstHost.String(),
		"last":    lastHost.String(),
		"hosts":   hosts.String(),
		"private": addr.IsPrivate(),
	}
	if hostBits > 0 {
		info["host_bits"] = hostBits
	}
	if addr.Is4() {
		info["broadcast"] = last.String()
		info["wildcard"] = wildcard(prefix)
	}
	return info
}

// lastAddr returns the highest address inside the prefix.
func lastAddr(prefix netip.Prefix) netip.Addr {
	addr := prefix.Masked().Addr()
	b := addr.As16()
	if addr.Is4() {
		b4 := addr.As4()
		for i := range b4 {
			if prefix.Bits() <= i*8 {
				b4[i] = 0xff
			}
		}
		// fill the bits after the prefix inside the last partial byte
		if rem := prefix.Bits() % 8; rem != 0 {
			b4[prefix.Bits()/8] |= 0xff >> rem
		}
		return netip.AddrFrom4(b4)
	}
	for i := range b {
		if prefix.Bits() <= i*8 {
			b[i] = 0xff
		}
	}
	if rem := prefix.Bits() % 8; rem != 0 {
		b[prefix.Bits()/8] |= 0xff >> rem
	}
	return netip.AddrFrom16(b)
}

// netmask renders the prefix length as an address mask.
func netmask(prefix netip.Prefix) string {
	if prefix.Addr().Is4() {
		return netip.AddrFrom4([4]byte(net.CIDRMask(prefix.Bits(), 32))).String()
	}
	return netip.AddrFrom16([16]byte(net.CIDRMask(prefix.Bits(), 128))).String()
}

// wildcard renders the host bits as an address, the inverse of the netmask.
func wildcard(prefix netip.Prefix) string {
	mask := net.CIDRMask(prefix.Bits(), 32)
	for i := range mask {
		mask[i] = ^mask[i]
	}
	return netip.AddrFrom4([4]byte(mask)).String()
}

var ipv6Action = New(Definition[[]byte, map[string]any]{
	Doc:          "IPv6 details for an address or a prefix: compressed, expanded, reverse DNS, prefix length and host bits",
	Names:        []string{"ipv6"},
	Type:         TransformAction,
	InputFormat:  TextFormat,
	OutputFormat: DictFormat,
	Func: func(a Action, in []byte) (map[string]any, error) {
		s := strings.TrimSpace(string(in))
		addr, prefix, err := parseAddrOrPrefix(s)
		if err != nil {
			return nil, err
		}
		info := map[string]any{
			"compressed": addr.String(),
			"expanded":   expandedAddr(addr),
			"loopback":   addr.IsLoopback(),
			"link_local": addr.IsLinkLocalUnicast(),
			"private":    addr.IsPrivate(),
		}
		if rev, err := reverseName(addr); err == nil {
			info["reverse"] = rev
		}
		if prefix != nil {
			info["prefix"] = prefix.String()
			info["prefix_len"] = prefix.Bits()
			info["network"] = prefix.Masked().Addr().String()
			info["host_bits"] = prefix.Addr().BitLen() - prefix.Bits()
		}
		return info, nil
	},
})

// parseAddrOrPrefix parses either form, returning a prefix only when given.
func parseAddrOrPrefix(s string) (netip.Addr, *netip.Prefix, error) {
	if strings.Contains(s, "/") {
		p, err := netip.ParsePrefix(s)
		if err != nil {
			return netip.Addr{}, nil, fmt.Errorf("%q is not a valid address or prefix: %w", s, err)
		}
		return p.Addr(), &p, nil
	}
	a, err := netip.ParseAddr(s)
	if err != nil {
		return netip.Addr{}, nil, fmt.Errorf("%q is not a valid address: %w", s, err)
	}
	return a, nil, nil
}

// expandedAddr renders the full uncompressed form (::f -> 0000:...:000f).
func expandedAddr(a netip.Addr) string {
	if a.Is4() {
		return a.String()
	}
	b := a.As16()
	parts := make([]string, 8)
	for i := range parts {
		parts[i] = fmt.Sprintf("%02x%02x", b[i*2], b[i*2+1])
	}
	return strings.Join(parts, ":")
}

// reverseName builds the in-addr.arpa or ip6.arpa name for the address.
// It never does a DNS lookup: actions must stay pure and offline.
func reverseName(a netip.Addr) (string, error) {
	if !a.IsValid() {
		return "", fmt.Errorf("invalid address")
	}
	if a.Is4() {
		b := a.As4()
		return fmt.Sprintf("%d.%d.%d.%d.in-addr.arpa", b[3], b[2], b[1], b[0]), nil
	}
	var sb strings.Builder
	b := a.As16()
	for i := len(b) - 1; i >= 0; i-- {
		fmt.Fprintf(&sb, "%x.%x.", b[i]&0x0f, b[i]>>4)
	}
	sb.WriteString("ip6.arpa")
	return sb.String(), nil
}
