package rpcbench

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type rpcClient struct {
	httpClient *http.Client
	method     string
}

func newRPCClient(method string) *rpcClient {
	return &rpcClient{
		httpClient: http.DefaultClient,
		method:     method,
	}
}

func (c *rpcClient) call(ctx context.Context, url string) Outcome {
	body, err := json.Marshal(rpcRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  c.method,
		Params:  []any{},
	})
	if err != nil {
		return OutcomeError
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return OutcomeError
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return OutcomeTimeout
		}
		return OutcomeError
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return OutcomeError
	}

	var rpcResp rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return OutcomeError
	}

	if rpcResp.JSONRPC != "2.0" {
		return OutcomeError
	}
	if rpcResp.ID != 1 {
		return OutcomeError
	}
	if rpcResp.Error != nil {
		return OutcomeError
	}
	if rpcResp.Result == nil {
		return OutcomeError
	}

	return OutcomeSuccess
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
