package main

import (
	"context"
	"io"
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/joho/godotenv"

	"babyagent/ch05"
	"babyagent/ch05/tool"
	"babyagent/shared"
)

func main() {
	_ = godotenv.Load()

	ctx := context.Background()
	modelConf := shared.NewModelConfig()

	mcpServerMap, err := shared.LoadMcpServerConfig("mcp-server.json")
	if err != nil {
		log.Printf("Failed to load MCP server configuration: %v", err)
	}
	mcpClients := make([]*ch05.McpClient, 0)
	for k, v := range mcpServerMap {
		mcpClient := ch05.NewMcpToolProvider(k, v)
		if err := mcpClient.RefreshTools(ctx); err != nil {
			log.Printf("Failed to refresh tools for MCP server %s: %v", k, err)
			continue
		}
		mcpClients = append(mcpClients, mcpClient)
	}

	agent := ch05.NewAgent(
		modelConf,
		ch05.CodingAgentSystemPrompt,
		[]tool.Tool{tool.NewBashTool()},
		mcpClients,
	)

	log.SetOutput(io.Discard)
	p := tea.NewProgram(newModel(agent, modelConf.Model))
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
