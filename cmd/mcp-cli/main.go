package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("MCP Platform CLI - Debugging & Testing Tool")
	fmt.Println("-------------------------------------------")

	// Start the server sub-process
	cmd := exec.Command("go", "run", "cmd/server/main.go")
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		return
	}
	defer cmd.Process.Kill()

	reader := bufio.NewReader(stdout)
	writer := bufio.NewWriter(stdin)

	// Simple loop to call platform.health
	fmt.Println("Calling platform.health to verify connectivity...")
	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "platform.health",
			"arguments": map[string]any{"_api_key": "admin-token"},
		},
	}

	data, _ := json.Marshal(req)
	writer.Write(data)
	writer.WriteByte('\n')
	writer.Flush()

	line, _, _ := reader.ReadLine()
	fmt.Printf("Server Response: %s\n", string(line))
}
