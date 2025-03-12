package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/miekg/dns"
)

type DoHResolver struct {
	upstream string
	client   *http.Client
}

func NewDoHResolver(upstream string) *DoHResolver {
	host := regexp.MustCompile(`^https://|/dns-query$`).ReplaceAllString(upstream, "")
	if ip := net.ParseIP(host); ip != nil && ip.To4() == nil {
		host = fmt.Sprintf("[%s]", ip)
	}

	upstream = fmt.Sprintf("https://%s/dns-query", host)

	c := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 5 * time.Second,
			MaxIdleConnsPerHost: 100,
			MaxIdleConns:        100,
		},
	}

	return &DoHResolver{
		upstream: upstream,
		client:   c,
	}
}

func (r *DoHResolver) Exchange(ctx context.Context, msg *dns.Msg) (*dns.Msg, error) {
	client := r.client
	pack, err := msg.Pack()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s?dns=%s", r.upstream, base64.RawStdEncoding.EncodeToString(pack))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Accept", "application/dns-message")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("doh status error")
	}

	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}

	resultMsg := new(dns.Msg)
	err = resultMsg.Unpack(buf.Bytes())
	if err != nil {
		return nil, err
	}

	if resultMsg.Rcode != dns.RcodeSuccess {
		return nil, errors.New("doh rcode wasn't successful")
	}

	return resultMsg, nil
}
