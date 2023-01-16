package traefik_jsonrpc_filter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Config the plugin configuration.
type Config struct {
	Allowlist           []string `json:"allowlist,omitempty"`
	BatchedRequestLimit int      `json:"batchedRequestLimit,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Allowlist:           make([]string, 0),
		BatchedRequestLimit: 1,
	}
}

// Demo a Demo plugin.
type JSONRPCFilter struct {
	next                http.Handler
	allowlist           []string
	batchedRequestLimit int
	name                string
}

type JSONRPCRequest struct {
	Method string `json:"method"`
}

func stringInSlice(target string, list []string) bool {
	for _, b := range list {
		if b == target {
			return true
		}
	}
	return false
}

func (jf *JSONRPCFilter) isSingleRequestBlocked(req JSONRPCRequest) bool {
	return !stringInSlice(req.Method, jf.allowlist)
}

func (jf *JSONRPCFilter) isBatchRequestBlocked(reqs []JSONRPCRequest) bool {
	if len(reqs) > jf.batchedRequestLimit {
		return true
	}

	for _, req := range reqs {
		if jf.isSingleRequestBlocked(req) {
			return true
		}
	}

	return false
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.Allowlist) == 0 {
		return nil, fmt.Errorf("allowlist cannot be empty")
	}

	return &JSONRPCFilter{
		allowlist:           config.Allowlist,
		batchedRequestLimit: config.BatchedRequestLimit,
		next:                next,
		name:                name,
	}, nil
}

func (jf *JSONRPCFilter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	parsed_request := JSONRPCRequest{}
	err = json.Unmarshal(body, &parsed_request)

	if err == nil {
		blocked := jf.isSingleRequestBlocked(parsed_request)

		if blocked {
			http.Error(rw, "JSON-RPC method blocked", http.StatusForbidden)
			return
		}

		jf.next.ServeHTTP(rw, req)
		return
	}

	batched_requests := make([]JSONRPCRequest, 0)
	err = json.Unmarshal(body, &batched_requests)

	if err == nil {
		blocked := jf.isBatchRequestBlocked(batched_requests)

		if blocked {
			http.Error(rw, "JSON-RPC methods blocked", http.StatusForbidden)
			return
		}

		jf.next.ServeHTTP(rw, req)
		return
	}

	jf.next.ServeHTTP(rw, req)
}
