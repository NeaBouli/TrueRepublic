package healthcheck

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const (
	validHealthBody = `{"jsonrpc":"2.0","id":-1,"result":{}}`
	validStatusBody = `{"jsonrpc":"2.0","id":-1,"result":{"sync_info":{"latest_block_height":"123","catching_up":false}}}`
)

// rpcServer starts an httptest server that serves body with the given status
// for every path. Its URL is a literal loopback http URL, so it passes base
// URL validation.
func rpcServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestLiveSuccess(t *testing.T) {
	srv := rpcServer(t, http.StatusOK, validHealthBody)
	if err := Live(context.Background(), srv.URL, DefaultTimeout); err != nil {
		t.Fatalf("Live returned unexpected error: %v", err)
	}
}

func TestLiveIgnoresSynchronization(t *testing.T) {
	// Liveness must not depend on sync state: a catching-up node with a
	// result object is still live.
	srv := rpcServer(t, http.StatusOK, `{"jsonrpc":"2.0","id":-1,"result":{"catching_up":true,"latest_block_height":"0"}}`)
	if err := Live(context.Background(), srv.URL, DefaultTimeout); err != nil {
		t.Fatalf("Live must ignore synchronization state, got: %v", err)
	}
}

func TestReadySuccess(t *testing.T) {
	srv := rpcServer(t, http.StatusOK, validStatusBody)
	if err := Ready(context.Background(), srv.URL, DefaultTimeout); err != nil {
		t.Fatalf("Ready returned unexpected error: %v", err)
	}
}

func TestReadySuccessNumericHeight(t *testing.T) {
	srv := rpcServer(t, http.StatusOK, `{"jsonrpc":"2.0","result":{"sync_info":{"latest_block_height":42,"catching_up":false}}}`)
	if err := Ready(context.Background(), srv.URL, DefaultTimeout); err != nil {
		t.Fatalf("Ready must accept an integral JSON number height, got: %v", err)
	}
}

func TestInvalidBaseURLs(t *testing.T) {
	srv := rpcServer(t, http.StatusOK, validHealthBody)
	loopback := srv.URL // e.g. http://127.0.0.1:PORT

	cases := map[string]struct {
		url string
		err error
	}{
		"https scheme":        {"https://127.0.0.1:26657", errScheme},
		"no scheme":           {"127.0.0.1:26657", errMalformedURL},
		"userinfo":            {strings.Replace(loopback, "http://", "http://user:pass@", 1), errUserinfo},
		"query":               {loopback + "?x=1", errQueryFragment},
		"fragment":            {loopback + "#frag", errQueryFragment},
		"non-root path":       {loopback + "/status", errPath},
		"hostname not ip":     {"http://localhost:26657", errHost},
		"non-loopback ip":     {"http://192.168.1.10:26657", errHost},
		"public ip":           {"http://8.8.8.8:26657", errHost},
		"unspecified ip":      {"http://0.0.0.0:26657", errHost},
		"ipv6 non-loopback":   {"http://[2001:db8::1]:26657", errHost},
		"empty url":           {"", errScheme},
		"control character":   {"http://127.0.0.1:26657/\x7f", errMalformedURL},
		"short notation ipv4": {"http://127.1:26657", errHost},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			for _, probe := range []struct {
				name string
				fn   func(context.Context, string, time.Duration) error
			}{{"live", Live}, {"ready", Ready}} {
				err := probe.fn(context.Background(), tc.url, DefaultTimeout)
				if !errors.Is(err, tc.err) {
					t.Fatalf("%s: expected %v, got %v", probe.name, tc.err, err)
				}
				assertOperatorSafe(t, err, tc.url)
			}
		})
	}
}

func TestValidBaseURLVariants(t *testing.T) {
	srv := rpcServer(t, http.StatusOK, validHealthBody)
	// A trailing slash on the root path is still a root path.
	if err := Live(context.Background(), srv.URL+"/", DefaultTimeout); err != nil {
		t.Fatalf("trailing root slash must be accepted, got: %v", err)
	}
	// Non-default loopback addresses in 127.0.0.0/8 are valid as URLs; the
	// request will fail to connect, but must fail as a request error, not a
	// validation error. Some platforms refuse, others time out.
	err := Live(context.Background(), "http://127.0.0.2:1", 200*time.Millisecond)
	if !errors.Is(err, errRequestFailed) && !errors.Is(err, errRequestTimeout) {
		t.Fatalf("127.0.0.2 must pass URL validation and fail at request time, got: %v", err)
	}
}

func TestTimeoutBounds(t *testing.T) {
	srv := rpcServer(t, http.StatusOK, validHealthBody)
	for _, d := range []time.Duration{0, -time.Second, MaxTimeout + time.Nanosecond, 11 * time.Second, time.Hour} {
		if err := Live(context.Background(), srv.URL, d); !errors.Is(err, errTimeoutRange) {
			t.Fatalf("timeout %v: expected %v, got %v", d, errTimeoutRange, err)
		}
		if err := Ready(context.Background(), srv.URL, d); !errors.Is(err, errTimeoutRange) {
			t.Fatalf("timeout %v: expected %v, got %v", d, errTimeoutRange, err)
		}
	}
	// Boundary values must be accepted.
	for _, d := range []time.Duration{time.Nanosecond, MaxTimeout} {
		srv := rpcServer(t, http.StatusOK, validHealthBody)
		if err := Live(context.Background(), srv.URL, d); err != nil && errors.Is(err, errTimeoutRange) {
			t.Fatalf("timeout %v must be accepted, got %v", d, err)
		}
	}
}

func TestUnreachable(t *testing.T) {
	// Bind and immediately release a port so nothing listens on it.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}
	err = Live(context.Background(), "http://"+addr, DefaultTimeout)
	if !errors.Is(err, errRequestFailed) {
		t.Fatalf("expected %v, got %v", errRequestFailed, err)
	}
	assertOperatorSafe(t, err, addr)
}

func TestRequestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	err := Live(context.Background(), srv.URL, 50*time.Millisecond)
	if !errors.Is(err, errRequestTimeout) {
		t.Fatalf("expected %v, got %v", errRequestTimeout, err)
	}
	assertOperatorSafe(t, err, srv.URL)
}

func TestNon200Status(t *testing.T) {
	for _, status := range []int{http.StatusInternalServerError, http.StatusNotFound, http.StatusUnauthorized} {
		srv := rpcServer(t, status, validHealthBody)
		err := Live(context.Background(), srv.URL, DefaultTimeout)
		want := fmt.Sprintf("healthcheck: unexpected http status %d", status)
		if err == nil || err.Error() != want {
			t.Fatalf("status %d: expected %q, got %v", status, want, err)
		}
		assertOperatorSafe(t, err, srv.URL)
	}
}

func TestRedirectIsNotFollowed(t *testing.T) {
	var redirected atomic.Bool
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		redirected.Store(true)
		_, _ = w.Write([]byte(validHealthBody))
	}))
	t.Cleanup(target.Close)
	source := httptest.NewServer(http.RedirectHandler(target.URL, http.StatusFound))
	t.Cleanup(source.Close)

	err := Live(context.Background(), source.URL, DefaultTimeout)
	if err == nil || err.Error() != "healthcheck: unexpected http status 302" {
		t.Fatalf("expected redirect rejection, got %v", err)
	}
	if redirected.Load() {
		t.Fatal("health probe followed a redirect")
	}
}

func TestEnvironmentProxyIsIgnored(t *testing.T) {
	var proxyUsed atomic.Bool
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		proxyUsed.Store(true)
		http.Error(w, "proxy must not receive local probes", http.StatusBadGateway)
	}))
	t.Cleanup(proxy.Close)
	t.Setenv("HTTP_PROXY", proxy.URL)
	t.Setenv("NO_PROXY", "")

	target := rpcServer(t, http.StatusOK, validHealthBody)
	if err := Live(context.Background(), target.URL, DefaultTimeout); err != nil {
		t.Fatalf("local probe failed with proxy configured: %v", err)
	}
	if proxyUsed.Load() {
		t.Fatal("local health probe used the environment proxy")
	}
}

func TestMalformedJSON(t *testing.T) {
	for _, body := range []string{"not json", `{"result":`, ``, `[1,2,3]`, `"result"`, `{"result":{}} trailing`} {
		srv := rpcServer(t, http.StatusOK, body)
		if err := Live(context.Background(), srv.URL, DefaultTimeout); !errors.Is(err, errInvalidJSON) {
			t.Fatalf("body %q: expected %v, got %v", body, errInvalidJSON, err)
		}
		if err := Ready(context.Background(), srv.URL, DefaultTimeout); !errors.Is(err, errInvalidJSON) {
			t.Fatalf("body %q: expected %v, got %v", body, errInvalidJSON, err)
		}
	}
}

func TestOversizedBody(t *testing.T) {
	srv := rpcServer(t, http.StatusOK, strings.Repeat("x", maxBodyBytes+1))
	if err := Live(context.Background(), srv.URL, DefaultTimeout); !errors.Is(err, errBodyTooLarge) {
		t.Fatalf("expected %v, got %v", errBodyTooLarge, err)
	}
	if err := Ready(context.Background(), srv.URL, DefaultTimeout); !errors.Is(err, errBodyTooLarge) {
		t.Fatalf("expected %v, got %v", errBodyTooLarge, err)
	}
}

func TestMissingResult(t *testing.T) {
	for _, body := range []string{
		`{"jsonrpc":"2.0","id":-1}`,
		`{"jsonrpc":"2.0","id":-1,"result":null}`,
		`{"jsonrpc":"2.0","id":-1,"result":5}`,
		`{"jsonrpc":"2.0","id":-1,"result":"ok"}`,
		`{"jsonrpc":"2.0","id":-1,"error":{"code":-32600,"message":"invalid"}}`,
	} {
		srv := rpcServer(t, http.StatusOK, body)
		if err := Live(context.Background(), srv.URL, DefaultTimeout); !errors.Is(err, errMissingResult) {
			t.Fatalf("body %q: expected %v, got %v", body, errMissingResult, err)
		}
		if err := Ready(context.Background(), srv.URL, DefaultTimeout); !errors.Is(err, errMissingResult) {
			t.Fatalf("body %q: expected %v, got %v", body, errMissingResult, err)
		}
	}
}

func TestInvalidJSONRPCVersion(t *testing.T) {
	for _, body := range []string{
		`{"result":{}}`,
		`{"jsonrpc":"1.0","result":{}}`,
		`{"jsonrpc":2,"result":{}}`,
	} {
		srv := rpcServer(t, http.StatusOK, body)
		if err := Live(context.Background(), srv.URL, DefaultTimeout); !errors.Is(err, errInvalidJSONRPC) {
			t.Fatalf("body %q: expected %v, got %v", body, errInvalidJSONRPC, err)
		}
	}
}

func TestReadyStatusFailures(t *testing.T) {
	cases := map[string]struct {
		body string
		err  error
	}{
		"missing sync_info":      {`{"jsonrpc":"2.0","result":{}}`, errMissingSync},
		"sync_info not object":   {`{"jsonrpc":"2.0","result":{"sync_info":"x"}}`, errMissingSync},
		"catching_up true":       {`{"jsonrpc":"2.0","result":{"sync_info":{"latest_block_height":"10","catching_up":true}}}`, errCatchingUp},
		"catching_up missing":    {`{"jsonrpc":"2.0","result":{"sync_info":{"latest_block_height":"10"}}}`, errCatchingUpType},
		"catching_up not bool":   {`{"jsonrpc":"2.0","result":{"sync_info":{"latest_block_height":"10","catching_up":"false"}}}`, errCatchingUpType},
		"zero height":            {`{"jsonrpc":"2.0","result":{"sync_info":{"latest_block_height":"0","catching_up":false}}}`, errHeight},
		"negative height":        {`{"jsonrpc":"2.0","result":{"sync_info":{"latest_block_height":"-5","catching_up":false}}}`, errHeight},
		"malformed height":       {`{"jsonrpc":"2.0","result":{"sync_info":{"latest_block_height":"abc","catching_up":false}}}`, errHeight},
		"empty height":           {`{"jsonrpc":"2.0","result":{"sync_info":{"latest_block_height":"","catching_up":false}}}`, errHeight},
		"missing height":         {`{"jsonrpc":"2.0","result":{"sync_info":{"catching_up":false}}}`, errHeight},
		"fractional height":      {`{"jsonrpc":"2.0","result":{"sync_info":{"latest_block_height":1.5,"catching_up":false}}}`, errHeight},
		"height wrong type":      {`{"jsonrpc":"2.0","result":{"sync_info":{"latest_block_height":true,"catching_up":false}}}`, errHeight},
		"height with plus sign":  {`{"jsonrpc":"2.0","result":{"sync_info":{"latest_block_height":"+7","catching_up":false}}}`, errHeight},
		"height overflowing int": {`{"jsonrpc":"2.0","result":{"sync_info":{"latest_block_height":"99999999999999999999999999","catching_up":false}}}`, errHeight},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			srv := rpcServer(t, http.StatusOK, tc.body)
			err := Ready(context.Background(), srv.URL, DefaultTimeout)
			if !errors.Is(err, tc.err) {
				t.Fatalf("expected %v, got %v", tc.err, err)
			}
			assertOperatorSafe(t, err, tc.body)
		})
	}
}

// assertOperatorSafe verifies an error message never leaks URLs, paths,
// response bodies, or credential-looking material.
func assertOperatorSafe(t *testing.T, err error, secrets ...string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error")
	}
	msg := err.Error()
	if !strings.HasPrefix(msg, "healthcheck: ") {
		t.Fatalf("error %q lacks the stable healthcheck prefix", msg)
	}
	for _, s := range secrets {
		if s != "" && strings.Contains(msg, s) {
			t.Fatalf("error %q leaks sensitive input %q", msg, s)
		}
	}
	for _, banned := range []string{"http://", "https://", "user:pass", "catching_up", "latest_block_height"} {
		if strings.Contains(msg, banned) {
			t.Fatalf("error %q contains banned fragment %q", msg, banned)
		}
	}
}
