package brain

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

const helperEnv = "NEA_AI_FAKE_MCP"

// TestHelperProcess is invoked as a subprocess by the other tests in this file.
// When NEA_AI_FAKE_MCP is set, it speaks the MCP JSON-RPC subset that
// PingMCPCommand expects. Otherwise it returns immediately so the test binary
// behaves normally for the rest of the suite.
func TestHelperProcess(t *testing.T) {
	role := os.Getenv(helperEnv)
	if role == "" {
		return
	}
	defer os.Exit(0)

	reader := bufio.NewReader(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	initReq, err := readRequest(reader)
	if err != nil {
		return
	}
	_ = encoder.Encode(map[string]any{
		"jsonrpc": "2.0",
		"id":      initReq["id"],
		"result":  map[string]any{"protocolVersion": "2024-11-05"},
	})

	toolsReq, err := readRequest(reader)
	if err != nil {
		return
	}

	switch role {
	case "happy":
		_ = encoder.Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      toolsReq["id"],
			"result": map[string]any{
				"tools": []map[string]any{
					{"name": "search"},
					{"name": "store"},
				},
			},
		})
	case "no_tools":
		_ = encoder.Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      toolsReq["id"],
			"result":  map[string]any{"tools": []map[string]any{}},
		})
	case "rpc_error":
		_ = encoder.Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      toolsReq["id"],
			"error":   map[string]any{"code": -32601, "message": "method not found"},
		})
	}
}

func readRequest(reader *bufio.Reader) (map[string]any, error) {
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	var request map[string]any
	if err := json.Unmarshal(line, &request); err != nil {
		return nil, err
	}
	return request, nil
}

func TestPingMCPCommandReturnsToolsOnSuccess(t *testing.T) {
	t.Setenv(helperEnv, "happy")

	result := PingMCPCommand(os.Args[0], "-test.run=TestHelperProcess", "--")
	if !result.OK {
		t.Fatalf("expected OK, got error %q", result.ErrorText)
	}
	if len(result.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %v", result.Tools)
	}
	want := map[string]bool{"search": true, "store": true}
	for _, name := range result.Tools {
		if !want[name] {
			t.Errorf("unexpected tool %q", name)
		}
	}
}

func TestPingMCPCommandFailsWhenNoToolsReturned(t *testing.T) {
	t.Setenv(helperEnv, "no_tools")

	result := PingMCPCommand(os.Args[0], "-test.run=TestHelperProcess", "--")
	if result.OK {
		t.Fatal("expected failure when tools list is empty")
	}
	if !strings.Contains(result.ErrorText, "tools/list") {
		t.Errorf("error %q should mention tools/list", result.ErrorText)
	}
}

func TestPingMCPCommandFailsOnRPCError(t *testing.T) {
	t.Setenv(helperEnv, "rpc_error")

	result := PingMCPCommand(os.Args[0], "-test.run=TestHelperProcess", "--")
	if result.OK {
		t.Fatal("expected failure on RPC error")
	}
	if !strings.Contains(result.ErrorText, "method not found") {
		t.Errorf("error %q should propagate RPC message", result.ErrorText)
	}
}

func TestPingMCPCommandFailsWhenBinaryMissing(t *testing.T) {
	result := PingMCPCommand("/no/such/binary-nea-ai-test")
	if result.OK {
		t.Fatal("expected failure for missing binary")
	}
}
