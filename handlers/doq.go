package handlers

import (
	"context"
	"strings"

	"github.com/miekg/dns"
	"github.com/tantalor93/doq-go/doq"
)

type DoQResolver struct {
	upstream string
	client   *doq.Client
}

func NewDoQResolver(upstream string) *DoQResolver {
	// if there isn't a port specified, default to 853
	if !strings.Contains(upstream, ":") {
		upstream += ":853"
	}

	addr := upstream

	client := doq.NewClient(addr)

	return &DoQResolver{
		upstream: addr,
		client:   client,
	}
}

func (r *DoQResolver) Exchange(ctx context.Context, msg *dns.Msg) (*dns.Msg, error) {
	client := r.client
	return client.Send(ctx, msg)
}
