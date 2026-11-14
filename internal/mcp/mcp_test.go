package mcp

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
	"testing"
)

// nan and inf are the JSON numbers json.Unmarshal cannot produce: a client
// that sends 1e999 gets +Inf, so the values reach ArgInt all the same.
func nan() float64 { return math.NaN() }

func inf(sign int) float64 { return math.Inf(sign) }

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

// A line that is not JSON is a parse error with no id, and the lines after it
// are still read: one bad message must not cost the client the session.
func TestParseError(t *testing.T) {
	got := serve(t, testServer(),
		`{oops`,
		`{"jsonrpc":"2.0","id":2,"method":"ping"}`,
	)
	if len(got) != 2 {
		t.Fatalf("got %d replies, want two", len(got))
	}
	errObj := got[0]["error"].(map[string]any)
	if code := int(errObj["code"].(float64)); code != CodeParse {
		t.Errorf("code = %d, want %d", code, CodeParse)
	}
	if got[0]["id"] != nil {
		t.Errorf("reply = %v, want no id on a parse error", got[0])
	}
	if got[1]["id"].(float64) != 2 || got[1]["error"] != nil {
		t.Errorf("reply = %v, want the pong after the bad line", got[1])
	}
}

// The spec names the version in every request; another spelling is an
// invalid request rather than an unknown method.
func TestInvalidJSONRPC(t *testing.T) {
	got := serve(t, testServer(),
		`{"jsonrpc":"1.0","id":1,"method":"ping"}`,
		`{"id":2,"method":"ping"}`,
	)
	for i := range got {
		if got[i]["error"] == nil {
			t.Fatalf("reply = %v, want an error", got[i])
		}
		if code := int(got[i]["error"].(map[string]any)["code"].(float64)); code != CodeInvalid {
			t.Errorf("code = %d, want %d", code, CodeInvalid)
		}
		if got[i]["id"] == nil {
			t.Errorf("reply = %v, want the id echoed", got[i])
		}
	}
}

// A message longer than maxLine is refused and ends the session: the rest of
// the line is not a message boundary, so the stream cannot be picked up again.
func TestOversizedMessage(t *testing.T) {
	big := `{"jsonrpc":"2.0","id":1,"method":"ping","x":"` + strings.Repeat("a", maxLine) + `"}`
	var out bytes.Buffer
	s := testServer()
	s.In = strings.NewReader(big + "\n")
	s.Out = &out
	if err := s.Serve(); err == nil {
		t.Error("Serve() = nil, want the read error")
	}
}

func TestArgInt(t *testing.T) {
	args := func(v any) map[string]any { return map[string]any{"n": v} }
	cases := []struct {
		name string
		args map[string]any
		want int
	}{
		{"missing", map[string]any{}, 7},
		{"wrong type", args("3"), 7},
		{"inside", args(5.0), 5},
		{"below min", args(-3.0), 1},
		{"above max", args(1e300), 10},
		{"negative overflow", args(-1e300), 1},
		// int64 wraps these to a number inside the range if they are cast
		// before the clamp.
		{"just past int64", args(9.3e18), 10},
		{"nan", args(nan()), 7},
		{"+inf", args(inf(1)), 7},
		{"-inf", args(inf(-1)), 7},
	}
	for _, c := range cases {
		if got := ArgInt(c.args, "n", 7, 1, 10); got != c.want {
			t.Errorf("%s: ArgInt() = %d, want %d", c.name, got, c.want)
		}
	}
}
