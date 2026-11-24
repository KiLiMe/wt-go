// Package mcp is a small Model Context Protocol server: JSON-RPC 2.0 over a
// stream (stdin/stdout by default), with the tools and resources capabilities.
// It depends on nothing but the standard library, so the exe stays a single
// file, and it touches no file itself: the tools the caller builds do.
package mcp

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// Boundary: the transport and the dispatch -- one line in, one message out,
// and the routing of a method to its handler. The handlers are handlers.go, the
// shape of what a tool answers args.go.

// ProtocolVersion is the revision this server speaks. Clients ask for their
// own in initialize; we answer with ours, as the spec allows.
const ProtocolVersion = "2025-06-18"

// maxLine is the longest single message the server reads. A client that sends
// a bigger one is not speaking this protocol, and reading it anyway would let
// one request decide how much memory the process takes.
const maxLine = 1 << 20

// JSON-RPC error codes.
const (
	CodeParse     = -32700
	CodeInvalid   = -32600
	CodeNoMethod  = -32601
	CodeBadParams = -32602
	CodeInternal  = -32603
)

// RPCError is the error object of a failed request.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *RPCError) Error() string { return e.Message }

// Tool is one callable tool. Schema is the inputSchema object: a JSON Schema
// with type "object". Run returns the text the client shows; an error comes
// back as a tool error, not a protocol error, as MCP asks.
type Tool struct {
	Name        string
	Title       string
	Description string
	Schema      map[string]any
	Run         func(args map[string]any) (string, error)
}

// Resource is one readable resource, addressed by URI.
type Resource struct {
	URI         string
	Name        string
	Description string
	MimeType    string
	Read        func() (string, error)
}

// Server answers requests from in and writes responses to out.
type Server struct {
	Name         string
	Version      string
	Instructions string
	Tools        []Tool
	Resources    []Resource
	In           io.Reader
	Out          io.Writer
	// Logf receives protocol-level notes (bad input and the like). It must
	// never write to Out: stdout carries the protocol alone.
	Logf func(format string, a ...any)
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// ErrNoStream is what Serve answers without somewhere to read or write.
var ErrNoStream = errors.New("mcp: In and Out are both needed")

// Serve reads messages until the stream ends. It reads one line per message
// rather than a json.Decoder on the stream: a decoder that fails half way
// leaves the stream between messages and every later request is read
// wrongly, while a line is a boundary the framing already gives us.
func (s *Server) Serve() error {
	if s.In == nil || s.Out == nil {
		return ErrNoStream
	}
	sc := bufio.NewScanner(bufio.NewReaderSize(s.In, 64<<10))
	sc.Buffer(make([]byte, 0, 64<<10), maxLine)
	enc := json.NewEncoder(s.Out)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			// One bad line must not end the session: the next line is still
			// a message, and the client is told which one it got wrong.
			s.logf("bad message: %v", err)
			if err := enc.Encode(&response{
				JSONRPC: "2.0",
				Error:   &RPCError{Code: CodeParse, Message: "parse error: " + err.Error()},
			}); err != nil {
				return err
			}
			continue
		}
		// Notifications (no id) get no answer, valid or not: the protocol
		// has nowhere to put one.
		if len(req.ID) == 0 {
			if req.JSONRPC != "2.0" {
				s.logf("bad message: jsonrpc = %q", req.JSONRPC)
			}
			continue
		}
		if err := enc.Encode(s.handle(&req)); err != nil {
			return err
		}
	}
	if err := sc.Err(); err != nil && err != io.EOF {
		// A line over maxLine means the client's message was never read:
		// say so and end, since the rest of the stream is half of it.
		s.logf("read: %v", err)
		return err
	}
	return nil
}

func (s *Server) logf(format string, a ...any) {
	if s.Logf != nil {
		s.Logf(format, a...)
	}
}

func (s *Server) handle(req *request) *response {
	res := &response{JSONRPC: "2.0", ID: req.ID}
	// The spec names the version in every message; a request that spells it
	// another way is not one we can answer with confidence.
	if req.JSONRPC != "2.0" {
		res.Error = &RPCError{Code: CodeInvalid, Message: fmt.Sprintf("invalid request: jsonrpc = %q, want \"2.0\"", req.JSONRPC)}
		return res
	}
	result, rerr := s.dispatch(req)
	if rerr != nil {
		res.Error = rerr
		return res
	}
	res.Result = result
	return res
}

func (s *Server) dispatch(req *request) (any, *RPCError) {
	switch req.Method {
	case "initialize":
		return s.initialize(req.Params), nil
	case "ping":
		return map[string]any{}, nil
	case "tools/list":
		return s.listTools(), nil
	case "tools/call":
		return s.callTool(req.Params)
	case "resources/list":
		return s.listResources(), nil
	case "resources/read":
		return s.readResource(req.Params)
	case "shutdown":
		return map[string]any{}, nil
	case "notifications/initialized", "notifications/cancelled":
		return map[string]any{}, nil
	}
	return nil, &RPCError{Code: CodeNoMethod, Message: "unknown method: " + req.Method}
}
