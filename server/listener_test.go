package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"testing"

	"github.com/pires/go-proxyproto"
)

func TestListenProxyProtocol(t *testing.T) {
	loopback := []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")}
	source := &net.TCPAddr{IP: net.ParseIP("203.0.113.1"), Port: 1000}

	tests := []struct {
		name           string
		proxyProtocol  bool
		trustedProxies []netip.Prefix
		headerVersion  byte
		wantRemoteIP   string
		wantErr        bool
	}{
		{
			name:         "disabled",
			wantRemoteIP: "127.0.0.1",
		},
		{
			name:          "disabled ignores header",
			headerVersion: 1,
			wantErr:       true,
		},
		{
			name:           "trusted v1 header",
			proxyProtocol:  true,
			trustedProxies: loopback,
			headerVersion:  1,
			wantRemoteIP:   "203.0.113.1",
		},
		{
			name:           "trusted v2 header",
			proxyProtocol:  true,
			trustedProxies: loopback,
			headerVersion:  2,
			wantRemoteIP:   "203.0.113.1",
		},
		{
			name:           "trusted without header",
			proxyProtocol:  true,
			trustedProxies: loopback,
			wantRemoteIP:   "127.0.0.1",
		},
		{
			name:           "untrusted header",
			proxyProtocol:  true,
			trustedProxies: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
			headerVersion:  1,
			wantErr:        true,
		},
		{
			name:           "untrusted without header",
			proxyProtocol:  true,
			trustedProxies: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
			wantRemoteIP:   "127.0.0.1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			listener, err := listen("127.0.0.1:0", test.proxyProtocol, test.trustedProxies)
			if err != nil {
				t.Fatalf("listen() error = %v", err)
			}
			server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, r.RemoteAddr)
			})}
			go server.Serve(listener)
			t.Cleanup(func() { server.Close() })

			conn, err := net.Dial("tcp", listener.Addr().String())
			if err != nil {
				t.Fatalf("Dial() error = %v", err)
			}
			defer conn.Close()

			if test.headerVersion != 0 {
				header := proxyproto.HeaderProxyFromAddrs(test.headerVersion, source, conn.RemoteAddr())
				if _, err := header.WriteTo(conn); err != nil {
					t.Fatalf("WriteTo() error = %v", err)
				}
			}
			fmt.Fprint(conn, "GET / HTTP/1.1\r\nHost: test\r\nConnection: close\r\n\r\n")

			response, err := http.ReadResponse(bufio.NewReader(conn), nil)
			if test.wantErr {
				if err == nil {
					response.Body.Close()
					if response.StatusCode == http.StatusOK {
						t.Fatalf("status = %d, want a failed request", response.StatusCode)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("ReadResponse() error = %v", err)
			}
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			host, _, err := net.SplitHostPort(string(body))
			if err != nil {
				t.Fatalf("SplitHostPort(%q) error = %v", body, err)
			}
			if host != test.wantRemoteIP {
				t.Errorf("RemoteAddr host = %q, want %q", host, test.wantRemoteIP)
			}
		})
	}
}
