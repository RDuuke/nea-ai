package brain

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"
)

type MCPPingResult struct {
	OK        bool     `json:"ok"`
	Tools     []string `json:"tools,omitempty"`
	ErrorText string   `json:"error,omitempty"`
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type toolsList struct {
	Tools []struct {
		Name string `json:"name"`
	} `json:"tools"`
}

func PingMCP(workDir string) MCPPingResult {
	brainPath, err := ResolveNeaBrain(workDir)
	if err != nil {
		return MCPPingResult{OK: false, ErrorText: err.Error()}
	}
	return PingMCPCommand(brainPath, "mcp")
}

func PingMCPCommand(command string, args ...string) MCPPingResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return MCPPingResult{OK: false, ErrorText: err.Error()}
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return MCPPingResult{OK: false, ErrorText: err.Error()}
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return MCPPingResult{OK: false, ErrorText: err.Error()}
	}

	if err := cmd.Start(); err != nil {
		return MCPPingResult{OK: false, ErrorText: err.Error()}
	}

	errCh := make(chan string, 1)
	go func() {
		data, _ := io.ReadAll(stderr)
		errCh <- string(data)
	}()

	encoder := json.NewEncoder(stdin)
	reader := bufio.NewReader(stdout)

	if err := encoder.Encode(rpcRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "nea-ai",
				"version": "doctor",
			},
		},
	}); err != nil {
		return finishPing(cmd, stdin, errCh, fmt.Sprintf("write initialize: %v", err))
	}
	if _, err := readResponse(reader); err != nil {
		return finishPing(cmd, stdin, errCh, fmt.Sprintf("initialize: %v", err))
	}

	if err := encoder.Encode(rpcRequest{JSONRPC: "2.0", ID: 2, Method: "tools/list"}); err != nil {
		return finishPing(cmd, stdin, errCh, fmt.Sprintf("write tools/list: %v", err))
	}
	response, err := readResponse(reader)
	if err != nil {
		return finishPing(cmd, stdin, errCh, fmt.Sprintf("tools/list: %v", err))
	}
	if response.Error != nil {
		return finishPing(cmd, stdin, errCh, response.Error.Message)
	}

	var result toolsList
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return finishPing(cmd, stdin, errCh, fmt.Sprintf("decode tools/list: %v", err))
	}
	tools := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		if tool.Name != "" {
			tools = append(tools, tool.Name)
		}
	}
	if len(tools) == 0 {
		return finishPing(cmd, stdin, errCh, "tools/list returned no tools")
	}

	_ = stdin.Close()
	_ = cmd.Wait()
	return MCPPingResult{OK: true, Tools: tools}
}

func readResponse(reader *bufio.Reader) (rpcResponse, error) {
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return rpcResponse{}, err
	}
	var response rpcResponse
	if err := json.Unmarshal(line, &response); err != nil {
		return rpcResponse{}, err
	}
	if response.Error != nil {
		return response, errors.New(response.Error.Message)
	}
	return response, nil
}

func finishPing(cmd *exec.Cmd, stdin io.Closer, errCh <-chan string, message string) MCPPingResult {
	_ = stdin.Close()
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	_ = cmd.Wait()
	select {
	case stderr := <-errCh:
		if stderr != "" {
			message += ": " + stderr
		}
	default:
	}
	return MCPPingResult{OK: false, ErrorText: message}
}
