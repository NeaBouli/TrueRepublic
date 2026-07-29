// Package healthcheck implements dependency-free local liveness and readiness
// probes for a TrueRepublic node against its CometBFT RPC endpoint.
//
// Liveness proves only that the local RPC responds with a valid JSON-RPC
// result; synchronization state must never affect it. Readiness additionally
// requires valid status, a positive latest block height, and catching_up=false.
//
// All returned errors are stable and operator-safe: they never contain URLs,
// paths, response bodies, or credential material.
package healthcheck

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	// DefaultRPCURL is the default local CometBFT RPC base URL.
	DefaultRPCURL = "http://127.0.0.1:26657"
	// DefaultTimeout is the default per-probe timeout.
	DefaultTimeout = 3 * time.Second
	// MaxTimeout is the largest accepted per-probe timeout.
	MaxTimeout = 10 * time.Second

	// maxBodyBytes bounds every RPC response body read by a probe.
	maxBodyBytes = 64 << 10
)

// Stable, operator-safe probe errors. None of them embed URLs, paths,
// response bodies, or credentials.
var (
	errMalformedURL   = errors.New("healthcheck: rpc url is not a valid url")
	errScheme         = errors.New("healthcheck: rpc url must use plain http")
	errUserinfo       = errors.New("healthcheck: rpc url must not contain userinfo")
	errQueryFragment  = errors.New("healthcheck: rpc url must not contain a query or fragment")
	errPath           = errors.New("healthcheck: rpc url must not contain a path")
	errHost           = errors.New("healthcheck: rpc url host must be a literal loopback address")
	errTimeoutRange   = errors.New("healthcheck: timeout must be positive and at most 10s")
	errRequestFailed  = errors.New("healthcheck: probe request failed")
	errRequestTimeout = errors.New("healthcheck: probe request timed out")
	errBodyTooLarge   = errors.New("healthcheck: response body exceeds 64 KiB limit")
	errBodyRead       = errors.New("healthcheck: failed to read response body")
	errInvalidJSON    = errors.New("healthcheck: response is not valid json")
	errInvalidJSONRPC = errors.New("healthcheck: response is not json-rpc 2.0")
	errMissingResult  = errors.New("healthcheck: response is missing a result object")
	errMissingSync    = errors.New("healthcheck: status is missing sync info")
	errCatchingUpType = errors.New("healthcheck: status is missing a boolean catching up flag")
	errHeight         = errors.New("healthcheck: latest block height is not a positive integer")
	errCatchingUp     = errors.New("healthcheck: node is catching up")
)

// Live reports nil when the node's CometBFT RPC answers /health with a valid
// JSON-RPC-shaped response containing a result object. Synchronization state
// is deliberately ignored so syncing never becomes a restart condition.
func Live(ctx context.Context, rpcURL string, timeout time.Duration) error {
	body, err := fetch(ctx, rpcURL, "/health", timeout)
	if err != nil {
		return err
	}
	result, err := decodeResult(body)
	if err != nil {
		return err
	}
	if result == nil {
		return errMissingResult
	}
	return nil
}

// Ready reports nil when the node's CometBFT RPC answers /status with a valid
// status whose sync_info has a positive integer latest_block_height and
// catching_up=false.
func Ready(ctx context.Context, rpcURL string, timeout time.Duration) error {
	body, err := fetch(ctx, rpcURL, "/status", timeout)
	if err != nil {
		return err
	}
	result, err := decodeResult(body)
	if err != nil {
		return err
	}
	if result == nil {
		return errMissingResult
	}
	syncInfo, ok := result["sync_info"].(map[string]any)
	if !ok {
		return errMissingSync
	}
	catchingUp, ok := syncInfo["catching_up"].(bool)
	if !ok {
		return errCatchingUpType
	}
	if catchingUp {
		return errCatchingUp
	}
	if !positiveHeight(syncInfo["latest_block_height"]) {
		return errHeight
	}
	return nil
}

// fetch validates the base URL and timeout, issues a bounded GET against
// base+path, enforces HTTP 200, and returns at most maxBodyBytes of body.
func fetch(ctx context.Context, baseURL, path string, timeout time.Duration) ([]byte, error) {
	base, err := validateBaseURL(baseURL)
	if err != nil {
		return nil, err
	}
	if err := validateTimeout(timeout); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path, nil)
	if err != nil {
		return nil, errRequestFailed
	}
	// Health probes are local control-plane operations. Never route them through
	// an environment-configured proxy or follow a redirect away from loopback.
	dialer := &net.Dialer{Timeout: timeout}
	transport := &http.Transport{
		DialContext:       dialer.DialContext,
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		var netErr net.Error
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
			return nil, errRequestTimeout
		}
		return nil, errRequestFailed
	}
	if resp.StatusCode != http.StatusOK {
		if err := resp.Body.Close(); err != nil {
			return nil, errBodyRead
		}
		return nil, fmt.Errorf("healthcheck: unexpected http status %d", resp.StatusCode)
	}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes+1))
	closeErr := resp.Body.Close()
	if readErr != nil || closeErr != nil {
		return nil, errBodyRead
	}
	if len(body) > maxBodyBytes {
		return nil, errBodyTooLarge
	}
	return body, nil
}

// validateBaseURL enforces plain HTTP with a literal loopback host
// (127.0.0.0/8 or ::1) and rejects userinfo, query, fragment, and non-root
// paths. It returns the normalized "scheme://host" base, never the raw input.
func validateBaseURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", errMalformedURL
	}
	if u.Scheme != "http" {
		return "", errScheme
	}
	if u.User != nil {
		return "", errUserinfo
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return "", errQueryFragment
	}
	if u.Path != "" && u.Path != "/" {
		return "", errPath
	}
	ip := net.ParseIP(u.Hostname())
	if ip == nil || !ip.IsLoopback() {
		return "", errHost
	}
	return "http://" + u.Host, nil
}

func validateTimeout(timeout time.Duration) error {
	if timeout <= 0 || timeout > MaxTimeout {
		return errTimeoutRange
	}
	return nil
}

// decodeResult parses a JSON-RPC-shaped document and returns its result
// object, or nil when the document has no result object.
func decodeResult(body []byte) (map[string]any, error) {
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, errInvalidJSON
	}
	if version, ok := doc["jsonrpc"].(string); !ok || version != "2.0" {
		return nil, errInvalidJSONRPC
	}
	result, ok := doc["result"].(map[string]any)
	if !ok {
		return nil, nil
	}
	return result, nil
}

// positiveHeight reports whether v is a positive integer block height, either
// as a canonical digit string (the CometBFT form) or an integral JSON number.
func positiveHeight(v any) bool {
	switch h := v.(type) {
	case string:
		if h == "" {
			return false
		}
		for _, r := range h {
			if r < '0' || r > '9' {
				return false
			}
		}
		n, err := strconv.ParseInt(h, 10, 64)
		return err == nil && n > 0
	case float64:
		return h > 0 && h == float64(int64(h))
	default:
		return false
	}
}
