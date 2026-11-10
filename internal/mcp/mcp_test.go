package mcp

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// serve answers the requests, each a JSON line, with the responses.
func serve(t *testing.T, s *Server, requests ...string) []map[string]any {
	t.Helper()
	var out bytes.Buffer
	s.In = strings.NewReader(strings.Join(requests, "\n") + "\n")
	s.Out = &out
	if err := s.Serve(); err != nil {
		t.Fatalf("Serve: %v", err)
	}
	var replies []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("reply %q: %v", line, err)
		}
		replies = append(replies, m)
	}
	return replies
}

func testServer() *Server {
	return &Server{
		Name:    "test",
		Version: "1.0",
		Tools: []Tool{{
			Name:        "echo",
			Description: "echoes",
			Schema:      ObjectSchema(map[string]any{"text": Prop("string", "what to say")}, []string{"text"}),
			Run: func(args map[string]any) (string, error) {
				text := ArgString(args, "text", "")
				if text == "" {
					return "no text", errNoText
				}
				return text, nil
			},
		}},
		Resources: []Resource{{
			URI:      "test://hello",
			Name:     "hello",
			MimeType: "text/plain",
			Read:     func() (string, error) { return "world", nil },
		}},
	}
}

var errNoText = &RPCError{Code: CodeBadParams, Message: "no text"}

func TestInitialize(t *testing.T) {
	got := serve(t, testServer(), `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}`)
	info := got[0]["result"].(map[string]any)
	if info["protocolVersion"] != ProtocolVersion {
		t.Errorf("protocolVersion = %v, want %s", info["protocolVersion"], ProtocolVersion)
	}
	caps := info["capabilities"].(map[string]any)
	if caps["tools"] == nil || caps["resources"] == nil {
		t.Errorf("capabilities = %v, want tools and resources", caps)
	}
	server := info["serverInfo"].(map[string]any)
	if server["name"] != "test" || server["version"] != "1.0" {
		t.Errorf("serverInfo = %v", server)
	}
}

func TestToolsList(t *testing.T) {
	got := serve(t, testServer(), `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	tools := got[0]["result"].(map[string]any)["tools"].([]any)
	if len(tools) != 1 {
		t.Fatalf("tools = %v, want one", tools)
	}
	tool := tools[0].(map[string]any)
	if tool["name"] != "echo" || tool["inputSchema"] == nil {
		t.Errorf("tool = %v", tool)
	}
}

func TestCallTool(t *testing.T) {
	got := serve(t, testServer(),
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"text":"hi"}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"echo","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"nope"}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/nope"}`,
	)
	ok := got[0]["result"].(map[string]any)
	content := ok["content"].([]any)[0].(map[string]any)
	if content["text"] != "hi" || ok["isError"] != nil {
		t.Errorf("result = %v, want the text", ok)
	}
	// A tool that fails answers with isError, not a protocol error.
	failed := got[1]["result"].(map[string]any)
	if failed["isError"] != true {
		t.Errorf("result = %v, want isError", failed)
	}
	for _, i := range []int{2, 3} {
		if got[i]["error"] == nil {
			t.Errorf("reply %d = %v, want an error", i, got[i])
		}
	}
	if code := got[2]["error"].(map[string]any)["code"].(float64); int(code) != CodeBadParams {
		t.Errorf("code = %v, want %d", code, CodeBadParams)
	}
	if code := got[3]["error"].(map[string]any)["code"].(float64); int(code) != CodeNoMethod {
		t.Errorf("code = %v, want %d", code, CodeNoMethod)
	}
}

func TestResources(t *testing.T) {
	got := serve(t, testServer(),
		`{"jsonrpc":"2.0","id":1,"method":"resources/list"}`,
		`{"jsonrpc":"2.0","id":2,"method":"resources/read","params":{"uri":"test://hello"}}`,
		`{"jsonrpc":"2.0","id":3,"method":"resources/read","params":{"uri":"test://nope"}}`,
	)
	list := got[0]["result"].(map[string]any)["resources"].([]any)
	if len(list) != 1 || list[0].(map[string]any)["uri"] != "test://hello" {
		t.Errorf("resources = %v", list)
	}
	contents := got[1]["result"].(map[string]any)["contents"].([]any)[0].(map[string]any)
	if contents["text"] != "world" || contents["mimeType"] != "text/plain" {
		t.Errorf("contents = %v", contents)
	}
	if got[2]["error"] == nil {
		t.Errorf("reply = %v, want an error", got[2])
	}
}

func TestNotificationsGetNoAnswer(t *testing.T) {
	got := serve(t, testServer(),
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":1,"method":"ping"}`,
	)
	if len(got) != 1 {
		t.Fatalf("got %d replies, want one", len(got))
	}
	if got[0]["id"].(float64) != 1 {
		t.Errorf("reply = %v, want the pong", got[0])
	}
}

func TestNoStream(t *testing.T) {
	s := &Server{}
	if err := s.Serve(); err != ErrNoStream {
		t.Errorf("Serve() = %v, want ErrNoStream", err)
	}
}
