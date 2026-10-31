package rpc

import (
	"errors"
	"fmt"
	"io"
)

// MaxHTTPErrorBody is the maximum number of response bytes kept on an
// HTTPStatusError. Longer bodies are truncated to this size.
const MaxHTTPErrorBody = 512

// ErrRPC matches every RPCError via errors.Is.
var ErrRPC = errors.New("rpc error")

// ErrHTTPStatus matches every HTTPStatusError via errors.Is.
var ErrHTTPStatus = errors.New("non-2xx status code")

// RPCError is an application-level JSON-RPC failure returned by
// [Client.JSONRPC] when the response envelope contains an error object.
//
// Error text keeps the historical form "rpc error: code=<n> message=<text>".
// errors.As recovers *RPCError, and errors.Is matches [ErrRPC].
type RPCError struct { //nolint:revive // rpc.Error is ambiguous; the API name is RPCError
	// Code is the numeric JSON-RPC error code reported by the server.
	Code int `json:"code"`

	// Message is the server-supplied explanation.
	Message string `json:"message"`
}

// Error returns the historical RPC error text.
func (e *RPCError) Error() string {
	return fmt.Sprintf("rpc error: code=%d message=%s", e.Code, e.Message)
}

// Unwrap returns [ErrRPC] so errors.Is can identify an RPC failure without
// inspecting the code or message.
func (e *RPCError) Unwrap() error {
	return ErrRPC
}

// HTTPStatusError is returned when a daemon or wallet endpoint responds with
// an HTTP status outside the 2xx range.
//
// Body holds at most [MaxHTTPErrorBody] bytes of the response, which is where
// a restricted RPC or a proxy usually explains the failure. Truncated is set
// when the response was longer than Body. Error text keeps the historical
// prefix "non-2xx status code: <status>". errors.As recovers *HTTPStatusError,
// and errors.Is matches [ErrHTTPStatus]. A failure while reading the body is
// wrapped and is also visible through errors.Is.
type HTTPStatusError struct {
	// StatusCode is the HTTP status code, outside the 2xx range.
	StatusCode int

	// Body is the response body, truncated to MaxHTTPErrorBody bytes.
	Body string

	// Truncated reports whether Body is shorter than the response.
	Truncated bool

	// readErr is set when the response body could not be read.
	readErr error

	// cause is the wrapped sentinel, plus readErr when the read failed.
	cause error
}

// Error returns the historical non-2xx text, appending a truncated body when
// one was present.
func (e *HTTPStatusError) Error() string {
	msg := fmt.Sprintf("non-2xx status code: %d", e.StatusCode)
	if e.Body != "" {
		msg += ": " + e.Body
		if e.Truncated {
			msg += "..."
		}
	}

	if e.readErr != nil {
		msg += ": read body: " + e.readErr.Error()
	}

	return msg
}

// Unwrap returns [ErrHTTPStatus], along with the body read failure when
// there was one, so errors.Is works for both.
func (e *HTTPStatusError) Unwrap() error {
	if e.cause != nil {
		return e.cause
	}

	return ErrHTTPStatus
}

// newHTTPStatusError reads at most MaxHTTPErrorBody bytes from body and
// builds the error returned for a non-2xx response.
func newHTTPStatusError(statusCode int, body io.Reader) *HTTPStatusError {
	statusErr := &HTTPStatusError{StatusCode: statusCode}
	if body == nil {
		return statusErr
	}

	text, truncated, err := readTruncatedBody(body, MaxHTTPErrorBody)
	statusErr.Body = text
	statusErr.Truncated = truncated
	if err != nil {
		statusErr.readErr = err
		statusErr.cause = errors.Join(ErrHTTPStatus, err)
	}

	return statusErr
}

// readTruncatedBody returns at most limit bytes. truncated is true when the
// reader held more than limit bytes.
func readTruncatedBody(r io.Reader, limit int) (string, bool, error) {
	if limit < 0 {
		limit = 0
	}

	buf, err := io.ReadAll(io.LimitReader(r, int64(limit)+1))
	if err != nil {
		return string(buf), false, err
	}

	if len(buf) > limit {
		return string(buf[:limit]), true, nil
	}

	return string(buf), false, nil
}
