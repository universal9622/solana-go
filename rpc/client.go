// Copyright 2020 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rpc

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	"github.com/klauspost/compress/gzhttp"
	"go.uber.org/ratelimit"
)

var ErrNotFound = errors.New("not found")
var ErrNotConfirmed = errors.New("not confirmed")

type Client struct {
	rpcURL    string
	rpcClient JSONRPCClient
	headers   http.Header
}

type JSONRPCClient interface {
	CallForInto(ctx context.Context, out interface{}, method string, params []interface{}) error
}

// New creates a new Solana JSON RPC client.
func New(rpcEndpoint string) *Client {
	opts := &jsonrpc.RPCClientOpts{
		HTTPClient: newHTTP(),
	}
	rpcClient := jsonrpc.NewClientWithOpts(rpcEndpoint, opts)
	return NewWithCustomRPCClient(rpcClient)
}

// NewWithCustomRPCClient creates a new Solana RPC client
// with the provided RPC client.
func NewWithCustomRPCClient(rpcClient JSONRPCClient) *Client {
	return &Client{
		rpcClient: rpcClient,
	}
}

var _ JSONRPCClient = &clientWithRateLimiting{}

type clientWithRateLimiting struct {
	rpcClient   jsonrpc.RPCClient
	rateLimiter ratelimit.Limiter
}

func (wr *clientWithRateLimiting) CallForInto(ctx context.Context, out interface{}, method string, params []interface{}) error {
	wr.rateLimiter.Take()
	return wr.rpcClient.CallForInto(ctx, &out, method, params)
}

// NewWithRateLimit creates a new rate-limitted Solana RPC client.
func NewWithRateLimit(
	rpcEndpoint string,
	rps int, // requests per second
) JSONRPCClient {
	opts := &jsonrpc.RPCClientOpts{
		HTTPClient: newHTTP(),
	}

	rpcClient := jsonrpc.NewClientWithOpts(rpcEndpoint, opts)

	return &clientWithRateLimiting{
		rpcClient:   rpcClient,
		rateLimiter: ratelimit.New(rps),
	}
}

func (c *Client) SetHeader(k, v string) {
	if c.headers == nil {
		c.headers = http.Header{}
	}
	c.headers.Set(k, v)
}

var (
	defaultMaxIdleConnsPerHost = 9
	defaultTimeout             = 5 * time.Minute
	defaultKeepAlive           = 180 * time.Second
)

func newHTTPTransport() *http.Transport {
	return &http.Transport{
		IdleConnTimeout:     defaultTimeout,
		MaxConnsPerHost:     defaultMaxIdleConnsPerHost,
		MaxIdleConnsPerHost: defaultMaxIdleConnsPerHost,
		Proxy:               http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   defaultTimeout,
			KeepAlive: defaultKeepAlive,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2: true,
		// MaxIdleConns:          100,
		TLSHandshakeTimeout: 10 * time.Second,
		// ExpectContinueTimeout: 1 * time.Second,
	}
}

// newHTTP returns a new Client from the provided config.
// Client is safe for concurrent use by multiple goroutines.
func newHTTP() *http.Client {
	tr := newHTTPTransport()

	return &http.Client{
		Timeout:   defaultTimeout,
		Transport: gzhttp.Transport(tr),
	}
}

// type wrapped struct {
// 	beforeHooks []JSONRPCClient
// 	cl          JSONRPCClient
// 	afterHooks  []JSONRPCClient
// }

// func (wr *wrapped) CallForInto(ctx context.Context, out interface{}, method string, params []interface{}) error {
// 	for _, ware := range wr.beforeHooks {
// 		err := ware.CallForInto(ctx, &out, method, params)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	err := wr.cl.CallForInto(ctx, &out, method, params)
// 	if err != nil {
// 		return err
// 	}

// 	for _, ware := range wr.afterHooks {
// 		err := ware.CallForInto(ctx, &out, method, params)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (cl *Client) WithBeforeHook(beforeHook JSONRPCClient) *Client {
// 	/* code */
// 	return nil
// }

// func (cl *Client) WithAfterHook(afterHook JSONRPCClient) *Client {
// 	/* code */
// 	return nil
// }

// // Expose this to make it easily expandable; or maybe don't ???
// func (cl *Client) CallForInto(ctx context.Context, out interface{}, method string, params []interface{}) error {
// 	return cl.rpcClient.CallForInto(ctx, &out, method, params)
// }
