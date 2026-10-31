// Package zmq subscribes to monerod's ZMQ PUB interface (--zmq-pub).
//
// ZMQ RPC (--zmq-rpc-bind) is a ZMQ_REP socket. Each request is one
// message holding a JSON-RPC object; monerod passes that string to
// RpcHandler.handle and writes the JSON-RPC response back. It is a
// second transport for the methods already exposed at HTTP /json_rpc,
// not another PUB topic, and it is intentionally not implemented here.
package zmq
