// Package rpc provides a client that's able to communicate with a `monerod`
// daemon via its RPC interfaces.
//
// # Errors
//
// Failures from [Client.JSONRPC] and [Client.RawRequest] stay human-readable
// and work with errors.Is and errors.As.
//
// A JSON-RPC error object is returned as an [*RPCError]. Its text keeps the
// historical form "rpc error: code=<n> message=<text>". errors.As recovers
// the *RPCError value, and errors.Is matches [ErrRPC].
//
// An HTTP status outside 2xx is returned as an [*HTTPStatusError]. StatusCode
// is the status, and Body is the response body truncated to [MaxHTTPErrorBody]
// bytes. Its text keeps the prefix "non-2xx status code: <status>" and appends
// the body when one was present. errors.As recovers the *HTTPStatusError, and
// errors.Is matches [ErrHTTPStatus].
//
// Transport failures are wrapped with %w. context.Canceled and
// context.DeadlineExceeded therefore match through errors.Is.
//
// Retries stay off unless the client is created with [WithRetry]. Without
// that option each call performs a single attempt, as before.
package rpc
