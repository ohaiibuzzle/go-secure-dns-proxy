package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/miekg/dns"
	"github.com/ohaiibuzzle/go-secure-dns-proxy/handlers"
)

type Resolver interface {
	Exchange(ctx context.Context, msg *dns.Msg) (*dns.Msg, error)
}
type RuntimeParams struct {
	upstream  string
	bootstrap string
	rsv       Resolver
}

var args RuntimeParams

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	// log.Default().Printf("--> %s\n", r.Question[0].Name)
	m, err := args.rsv.Exchange(context.Background(), r)
	if err != nil {
		log.Default().Printf("Error: %s\n", err)
		dns.HandleFailed(w, r)
		return
	}
	w.WriteMsg(m)
	// log.Default().Printf("<-- %s\n", m.Answer[0].String())
}

func main() {
	// Parse args
	type_flag := flag.String("type", "doh", "Type of resolver to use")
	upstream_flag := flag.String("upstream", "dns0.eu", "Upstream resolver to use")
	port_flag := flag.Int("port", 53, "Port to listen on")
	pidfile_flag := flag.String("pidfile", "", "File to write pid to")

	flag.Parse()

	args.upstream = *upstream_flag
	switch *type_flag {
	case "doh":
		args.rsv = handlers.NewDoHResolver(args.upstream)
	case "dot":
		args.rsv = handlers.NewDoTResolver(args.upstream)
	case "doq":
		args.rsv = handlers.NewDoQResolver(args.upstream)
	default:
		panic("Unknown type")
	}

	fmt.Printf("Using %s resolver with upstream %s\n", *type_flag, args.upstream)

	server := &dns.Server{
		Addr: fmt.Sprintf(":%d", *port_flag),
		Net:  "udp",
	}

	if *pidfile_flag != "" {
		f, err := os.Create(*pidfile_flag)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		fmt.Fprintf(f, "%d\n", os.Getpid())
	}

	dns.HandleFunc(".", handleRequest)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	fmt.Println("Listening on port", *port_flag)

	select {}
}
