package handlers

import (
	"context"
	"net"
	"strings"

	"github.com/miekg/dns"
)

type DoTResolver struct {
	upstream string
	client   *dns.Client
}

func NewDoTResolver(upstream string) *DoTResolver {
	c := &dns.Client{
		Net: "tcp-tls",
	}

	// Resolve upstream to IP
	ips, err := net.LookupIP(upstream)
	if err != nil {
		panic(err)
	}
	if len(ips) == 0 {
		panic("No IP found")
	}

	// Use the first IPv4 address
	for _, ip := range ips {
		if ip.To4() != nil {
			upstream = ip.String()
			break
		}
	}

	if !strings.Contains(upstream, ":") {
		upstream += ":853"
	}

	return &DoTResolver{
		upstream: upstream,
		client:   c,
	}
}

func (r *DoTResolver) Exchange(ctx context.Context, msg *dns.Msg) (*dns.Msg, error) {
	client := r.client
	ret_msg, _, err := client.ExchangeContext(ctx, msg, r.upstream)

	return ret_msg, err
}
