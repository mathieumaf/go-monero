package rpc_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mathieumaf/go-monero/pkg/rpc"
)

func TestRPCErrorStringAndSentinel(t *testing.T) {
	t.Parallel()

	err := &rpc.RPCError{Code: -1, Message: "foo"}
	require.Equal(t, "rpc error: code=-1 message=foo", err.Error())
	require.ErrorIs(t, err, rpc.ErrRPC)
	require.NotErrorIs(t, err, rpc.ErrHTTPStatus)
}

func TestHTTPStatusErrorStringAndSentinel(t *testing.T) {
	t.Parallel()

	empty := &rpc.HTTPStatusError{StatusCode: http.StatusInternalServerError}
	require.Equal(t, "non-2xx status code: 500", empty.Error())
	require.ErrorIs(t, empty, rpc.ErrHTTPStatus)

	withBody := &rpc.HTTPStatusError{
		StatusCode: http.StatusForbidden,
		Body:       "restricted",
	}
	require.Equal(t, "non-2xx status code: 403: restricted", withBody.Error())

	truncated := &rpc.HTTPStatusError{
		StatusCode: http.StatusBadGateway,
		Body:       "nope",
		Truncated:  true,
	}
	require.Equal(t, "non-2xx status code: 502: nope...", truncated.Error())
	require.NotErrorIs(t, truncated, rpc.ErrRPC)
}

func TestJSONRPCReturnsRPCError(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"id":"id","jsonrpc":"2.0","error":{"code":-1,"message":"foo"}}`)
	}))

	err := client.JSONRPC(context.Background(), "rpc-method", nil, nil)
	require.EqualError(t, err, "rpc error: code=-1 message=foo")
	require.ErrorIs(t, err, rpc.ErrRPC)

	var rpcErr *rpc.RPCError
	require.ErrorAs(t, err, &rpcErr)
	require.Equal(t, -1, rpcErr.Code)
	require.Equal(t, "foo", rpcErr.Message)

	var statusErr *rpc.HTTPStatusError
	require.False(t, errors.As(err, &statusErr))
}

func TestJSONRPCErrorWithZeroCode(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"error":{"code":0,"message":"busy"}}`)
	}))

	err := client.JSONRPC(context.Background(), "rpc-method", nil, nil)
	var rpcErr *rpc.RPCError
	require.ErrorAs(t, err, &rpcErr)
	require.Equal(t, 0, rpcErr.Code)
	require.Equal(t, "busy", rpcErr.Message)
	require.EqualError(t, err, "rpc error: code=0 message=busy")
}

func TestJSONRPCHTTPStatusKeepsTruncatedBody(t *testing.T) {
	t.Parallel()

	const extra = "END"
	payload := strings.Repeat("x", rpc.MaxHTTPErrorBody) + extra

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, payload)
	}))

	err := client.JSONRPC(context.Background(), "rpc-method", nil, nil)
	require.ErrorIs(t, err, rpc.ErrHTTPStatus)
	require.Contains(t, err.Error(), "submit request: non-2xx status code: 403: ")
	require.NotContains(t, err.Error(), extra)

	var statusErr *rpc.HTTPStatusError
	require.ErrorAs(t, err, &statusErr)
	require.Equal(t, http.StatusForbidden, statusErr.StatusCode)
	require.True(t, statusErr.Truncated)
	require.Len(t, statusErr.Body, rpc.MaxHTTPErrorBody)
	require.Equal(t, strings.Repeat("x", rpc.MaxHTTPErrorBody), statusErr.Body)

	var rpcErr *rpc.RPCError
	require.False(t, errors.As(err, &rpcErr))
}

func TestJSONRPCHTTPStatusEmptyBodyKeepsHistoricalString(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	err := client.JSONRPC(context.Background(), "rpc-method", nil, nil)
	require.EqualError(t, err, "submit request: non-2xx status code: 500")

	var statusErr *rpc.HTTPStatusError
	require.ErrorAs(t, err, &statusErr)
	require.Equal(t, http.StatusInternalServerError, statusErr.StatusCode)
	require.Empty(t, statusErr.Body)
	require.False(t, statusErr.Truncated)
}

func TestJSONRPCHTTPStatusShortBodyIsNotTruncated(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, "restricted rpc")
	}))

	err := client.JSONRPC(context.Background(), "rpc-method", nil, nil)
	require.EqualError(t, err, "submit request: non-2xx status code: 401: restricted rpc")

	var statusErr *rpc.HTTPStatusError
	require.ErrorAs(t, err, &statusErr)
	require.Equal(t, "restricted rpc", statusErr.Body)
	require.False(t, statusErr.Truncated)
}

func TestRawRequestHTTPStatusError(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "missing", http.StatusNotFound)
	}))

	err := client.RawRequest(context.Background(), "/get_height", nil, nil)
	var statusErr *rpc.HTTPStatusError
	require.ErrorAs(t, err, &statusErr)
	require.Equal(t, http.StatusNotFound, statusErr.StatusCode)
	require.Contains(t, statusErr.Body, "missing")
	require.ErrorIs(t, err, rpc.ErrHTTPStatus)
}

func TestHTTPStatusErrorWrapsBodyReadFailure(t *testing.T) {
	t.Parallel()

	httpClient := &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Status:     "502 Bad Gateway",
			Header:     make(http.Header),
			Body:       io.NopCloser(errReader{err: io.ErrUnexpectedEOF}),
			Request:    r,
		}, nil
	})}

	client, err := rpc.NewClient("http://example.test", rpc.WithHTTPClient(httpClient))
	require.NoError(t, err)

	err = client.RawRequest(context.Background(), "/get_info", nil, nil)
	require.ErrorIs(t, err, rpc.ErrHTTPStatus)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
	require.Contains(t, err.Error(), "non-2xx status code: 502")
	require.Contains(t, err.Error(), "read body: unexpected EOF")

	var statusErr *rpc.HTTPStatusError
	require.ErrorAs(t, err, &statusErr)
	require.Equal(t, http.StatusBadGateway, statusErr.StatusCode)
	require.Empty(t, statusErr.Body)
}

func TestContextCanceledIsWrapped(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.JSONRPC(ctx, "method", nil, nil)
	require.ErrorIs(t, err, context.Canceled)
	require.Contains(t, err.Error(), "do:")
}

func TestContextDeadlineExceededIsWrapped(t *testing.T) {
	t.Parallel()

	handler, started, release := stallRequest(t)
	client := newTestClient(t, handler)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- client.JSONRPC(ctx, "method", nil, nil)
	}()

	var err error
	select {
	case <-started:
		err = waitErr(t, errCh)
	case err = <-errCh:
	}

	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Contains(t, err.Error(), "do:")
	release()
}

func TestContextCanceledInFlightIsWrapped(t *testing.T) {
	t.Parallel()

	handler, started, release := stallRequest(t)
	client := newTestClient(t, handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- client.JSONRPC(ctx, "method", nil, nil)
	}()

	select {
	case <-started:
		cancel()
	case <-time.After(5 * time.Second):
		t.Fatal("request did not start")
	}

	err := waitErr(t, errCh)
	require.ErrorIs(t, err, context.Canceled)
	require.Contains(t, err.Error(), "do:")
	release()
}

func TestWithRetryDisabledByDefault(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, "unavailable")
	}))

	err := client.JSONRPC(context.Background(), "get_info", nil, nil)
	require.Equal(t, int32(1), calls.Load())
	require.EqualError(t, err, "submit request: non-2xx status code: 503: unavailable")
}

func TestWithRetryRetriesHTTPStatusThenSucceeds(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = io.WriteString(w, "unavailable")
			return
		}
		fmt.Fprint(w, `{"id":"0","jsonrpc":"2.0","result":{"ok":"yes"}}`)
	}), rpc.WithRetry(func(attempt int, err error) (time.Duration, bool) {
		var statusErr *rpc.HTTPStatusError
		if attempt == 0 && errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusServiceUnavailable {
			return 0, true
		}
		return 0, false
	}))

	result := map[string]string{}
	err := client.JSONRPC(context.Background(), "get_info", nil, &result)
	require.NoError(t, err)
	require.Equal(t, map[string]string{"ok": "yes"}, result)
	require.Equal(t, int32(2), calls.Load())
}

func TestWithRetryPolicyCanStop(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	var saw *rpc.RPCError
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		fmt.Fprint(w, `{"error":{"code":-32601,"message":"Method not found"}}`)
	}), rpc.WithRetry(func(_ int, err error) (time.Duration, bool) {
		require.ErrorAs(t, err, &saw)
		return 0, false
	}))

	err := client.JSONRPC(context.Background(), "nope", nil, nil)
	require.Equal(t, int32(1), calls.Load())
	require.ErrorIs(t, err, rpc.ErrRPC)
	require.Equal(t, -32601, saw.Code)
	require.Equal(t, "Method not found", saw.Message)
}

func TestWithRetryDoesNotRetryContextError(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	handler, started, release := stallRequest(t)
	counting := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		handler.ServeHTTP(w, r)
	})
	client := newTestClient(t, counting, rpc.WithRetry(func(int, error) (time.Duration, bool) {
		return time.Hour, true
	}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- client.JSONRPC(ctx, "method", nil, nil)
	}()

	select {
	case <-started:
		cancel()
	case <-time.After(5 * time.Second):
		t.Fatal("request did not start")
	}

	err := waitErr(t, errCh)
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, int32(1), calls.Load())
	release()
}

func TestWithRetryStopsWhenContextEndsDuringWait(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, "proxy")
	}), rpc.WithRetry(func(int, error) (time.Duration, bool) {
		cancel()
		return time.Hour, true
	}))

	err := client.JSONRPC(ctx, "get_info", nil, nil)
	require.ErrorIs(t, err, context.Canceled)
	require.ErrorIs(t, err, rpc.ErrHTTPStatus)
	require.Contains(t, err.Error(), "retry canceled:")
	require.Equal(t, int32(1), calls.Load())

	var statusErr *rpc.HTTPStatusError
	require.ErrorAs(t, err, &statusErr)
	require.Equal(t, "proxy", statusErr.Body)
}

func TestWithRetryNotCalledOnSuccess(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"result":{"ok":true}}`)
	}), rpc.WithRetry(func(int, error) (time.Duration, bool) {
		t.Fatal("policy should not run after success")
		return 0, false
	}))

	require.NoError(t, client.JSONRPC(context.Background(), "get_info", nil, &map[string]bool{}))
}

func TestHTTPStatusBodyAtExactLimitIsNotTruncated(t *testing.T) {
	t.Parallel()

	body := strings.Repeat("y", rpc.MaxHTTPErrorBody)
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, body)
	}))

	err := client.RawRequest(context.Background(), "/get_info", nil, nil)
	var statusErr *rpc.HTTPStatusError
	require.ErrorAs(t, err, &statusErr)
	require.False(t, statusErr.Truncated)
	require.Equal(t, body, statusErr.Body)
	require.NotContains(t, statusErr.Error(), "...")
}

// stallRequest blocks inside the handler until release is called.
// HTTP/1 does not cancel the server request context just because the client
// abandoned the call, so the test releases the handler after it has observed
// the client error. Otherwise httptest.Server.Close waits forever.
func stallRequest(t *testing.T) (http.Handler, <-chan struct{}, func()) {
	t.Helper()

	started := make(chan struct{})
	releaseCh := make(chan struct{})

	var once sync.Once
	release := func() {
		once.Do(func() { close(releaseCh) })
	}
	t.Cleanup(release)

	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		close(started)
		<-releaseCh
	})

	return handler, started, release
}

func waitErr(t *testing.T, errCh <-chan error) error {
	t.Helper()

	select {
	case err := <-errCh:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("request did not return")
		return nil
	}
}

func newTestClient(t *testing.T, handler http.Handler, opts ...rpc.ClientOption) *rpc.Client {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	opts = append([]rpc.ClientOption{rpc.WithHTTPClient(srv.Client())}, opts...)
	client, err := rpc.NewClient(srv.URL, opts...)
	require.NoError(t, err)

	return client
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type errReader struct {
	err error
}

func (e errReader) Read([]byte) (int, error) {
	return 0, e.err
}
