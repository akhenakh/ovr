package action

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAction_IPv4(t *testing.T) {
	r := NewRegistry()

	tests := []struct{ in, want string }{
		{"192.168.1.1", "3232235777"},
		{"3232235777", "192.168.1.1"},
		{"0xC0A80101", "192.168.1.1"},
		{"0b11000000101010000000000100000001", "192.168.1.1"},
		{"0", "0.0.0.0"},
		{"255.255.255.255", "4294967295"},
	}
	for _, tc := range tests {
		got, err := r.TextAction("ipv4", []byte(tc.in))
		require.NoError(t, err, tc.in)
		require.Equal(t, tc.want, string(got), tc.in)
	}

	_, err := r.TextAction("ipv4", []byte("banana"))
	require.Error(t, err)
	_, err = r.TextAction("ipv4", []byte("4294967296"))
	require.Error(t, err)
}

func TestAction_Subnet(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		in                  string
		wantNetwork         string
		wantBroadcast       any
		wantHosts           string
		wantFirst, wantLast string
	}{
		{"192.168.1.10/24", "192.168.1.0", "192.168.1.255", "254", "192.168.1.1", "192.168.1.254"},
		{"10.0.0.5/31", "10.0.0.4", "10.0.0.5", "2", "10.0.0.4", "10.0.0.5"},
		{"10.0.0.5", "10.0.0.5", "10.0.0.5", "1", "10.0.0.5", "10.0.0.5"},
	}
	for _, tc := range tests {
		got, err := r.DictAction("subnet", []byte(tc.in))
		require.NoError(t, err, tc.in)
		require.Equal(t, tc.wantNetwork, got["network"], tc.in)
		require.Equal(t, tc.wantBroadcast, got["broadcast"], tc.in)
		require.Equal(t, tc.wantHosts, got["hosts"], tc.in)
		require.Equal(t, tc.wantFirst, got["first"], tc.in)
		require.Equal(t, tc.wantLast, got["last"], tc.in)
	}

	got, err := r.DictAction("subnet", []byte("2001:db8::/32"))
	require.NoError(t, err)
	require.Equal(t, "2001:db8::", got["network"])
	require.Equal(t, 96, got["host_bits"])
	require.NotContains(t, got, "broadcast")

	_, err = r.DictAction("subnet", []byte("nope"))
	require.Error(t, err)
}

func TestAction_IPv6(t *testing.T) {
	r := NewRegistry()

	got, err := r.DictAction("ipv6", []byte("2001:db8::1"))
	require.NoError(t, err)
	require.Equal(t, "2001:db8::1", got["compressed"])
	require.Equal(t, "2001:0db8:0000:0000:0000:0000:0000:0001", got["expanded"])
	require.Contains(t, got["reverse"], ".ip6.arpa")
	require.NotContains(t, got, "prefix")

	got, err = r.DictAction("ipv6", []byte("fd00::/48"))
	require.NoError(t, err)
	require.Equal(t, 48, got["prefix_len"])
	require.Equal(t, 80, got["host_bits"])
	require.Equal(t, "fd00::", got["network"])

	_, err = r.DictAction("ipv6", []byte("not-an-ip"))
	require.Error(t, err)
}
