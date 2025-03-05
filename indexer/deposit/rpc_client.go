package deposit

import (
	"net"
	"net/http"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
)

func NewHTTPTransport(
	timeout time.Duration,
	maxIdleConnsPerHost int,
	keepAlive time.Duration,
) *http.Transport {
	return &http.Transport{
		IdleConnTimeout:     timeout,
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
		Proxy:               http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: keepAlive,
		}).Dial,
	}
}

// NewHTTP returns a new Client from the provided config.
func NewHTTP(
	timeout time.Duration,
	maxIdleConnsPerHost int,
	keepAlive time.Duration,
) *http.Client {
	tr := NewHTTPTransport(
		timeout,
		maxIdleConnsPerHost,
		keepAlive,
	)

	return &http.Client{
		Timeout:   timeout,
		Transport: tr,
	}
}

// NewRPC creates a new Solana JSON RPC client.
func NewRPC(rpcEndpoint string) *rpc.Client {
	var (
		defaultMaxIdleConnsPerHost = 10
		defaultTimeout             = 25 * time.Second
		defaultKeepAlive           = 180 * time.Second
	)
	opts := &jsonrpc.RPCClientOpts{
		HTTPClient: NewHTTP(
			defaultTimeout,
			defaultMaxIdleConnsPerHost,
			defaultKeepAlive,
		),
	}
	rpcClient := jsonrpc.NewClientWithOpts(rpcEndpoint, opts)
	return rpc.NewWithCustomRPCClient(rpcClient)
}
