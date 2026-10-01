package main

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/pires/go-proxyproto"
)

// listen binds a TCP listener on addr. With proxyProtocol enabled, peers in
// trustedProxies may send a PROXY protocol (v1 or v2) header, whose source
// address then becomes the connection's remote address. Headers from any other
// peer are rejected so clients cannot spoof their address.
func listen(addr string, proxyProtocol bool, trustedProxies []netip.Prefix) (net.Listener, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	if !proxyProtocol {
		return listener, nil
	}
	return &proxyproto.Listener{
		Listener:   listener,
		ConnPolicy: proxyProtocolPolicy(trustedProxies),
	}, nil
}

func proxyProtocolPolicy(trustedProxies []netip.Prefix) proxyproto.ConnPolicyFunc {
	return func(options proxyproto.ConnPolicyOptions) (proxyproto.Policy, error) {
		upstream, ok := options.Upstream.(*net.TCPAddr)
		if !ok {
			return proxyproto.REJECT, fmt.Errorf("%w: unexpected address %v", proxyproto.ErrInvalidUpstream, options.Upstream)
		}
		addr := upstream.AddrPort().Addr().Unmap()
		for _, prefix := range trustedProxies {
			if prefix.Contains(addr) {
				return proxyproto.USE, nil
			}
		}
		return proxyproto.REJECT, nil
	}
}
