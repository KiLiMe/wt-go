package mcp

import (
	"encoding/json"
	"sort"
)

// Boundary: the methods this server answers -- initialize and the tools and
// resources it publishes. A handler decodes its arguments with args.go and
// answers with a value the transport sends.

func (s *Server) initialize(params json.RawMessage) any {
	var in struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	_ = json.Unmarshal(params, &in)
	caps := map[string]any{
		"tools": map[string]any{"listChanged": false},
	}
	if len(s.Resources) > 0 {
		caps["resources"] = map[string]any{"subscribe": false, "listChanged": false}
	}
	out := map[string]any{
		"protocolVersion": ProtocolVersion,
		"capabilities":    caps,
		"serverInfo": map[string]any{
			"name":    s.Name,
			"version": s.Version,
		},
	}
	if s.Instructions != "" {
		out["instructions"] = s.Instructions
	}
	return out
}

func (s *Server) listTools() any {
	tools := make([]map[string]any, 0, len(s.Tools))
	for _, t := range s.Tools {
		m := map[string]any{
			"name":        t.Name,
			"description": t.Description,
			"inputSchema": t.Schema,
		}
		if t.Title != "" {
			m["title"] = t.Title
		}
		tools = append(tools, m)
	}
	sort.Slice(tools, func(i, j int) bool {
		return tools[i]["name"].(string) < tools[j]["name"].(string)
	})
	return map[string]any{"tools": tools}
}

func (s *Server) callTool(params json.RawMessage) (any, *RPCError) {
	var in struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(params, &in); err != nil {
		return nil, &RPCError{Code: CodeBadParams, Message: "tools/call: " + err.Error()}
	}
	for _, t := range s.Tools {
		if t.Name != in.Name {
			continue
		}
		text, err := t.Run(in.Arguments)
		if err != nil {
			return map[string]any{
				"content": []map[string]any{{"type": "text", "text": err.Error()}},
				"isError": true,
			}, nil
		}
		return map[string]any{
			"content": []map[string]any{{"type": "text", "text": text}},
		}, nil
	}
	return nil, &RPCError{Code: CodeBadParams, Message: "unknown tool: " + in.Name}
}

func (s *Server) listResources() any {
	out := make([]map[string]any, 0, len(s.Resources))
	for _, r := range s.Resources {
		out = append(out, map[string]any{
			"uri":         r.URI,
			"name":        r.Name,
			"description": r.Description,
			"mimeType":    r.MimeType,
		})
	}
	return map[string]any{"resources": out}
}

func (s *Server) readResource(params json.RawMessage) (any, *RPCError) {
	var in struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(params, &in); err != nil {
		return nil, &RPCError{Code: CodeBadParams, Message: "resources/read: " + err.Error()}
	}
	for _, r := range s.Resources {
		if r.URI != in.URI {
			continue
		}
		text, err := r.Read()
		if err != nil {
			return nil, &RPCError{Code: CodeInternal, Message: err.Error()}
		}
		return map[string]any{
			"contents": []map[string]any{{
				"uri":      r.URI,
				"mimeType": r.MimeType,
				"text":     text,
			}},
		}, nil
	}
	return nil, &RPCError{Code: CodeBadParams, Message: "unknown resource: " + in.URI}
}
