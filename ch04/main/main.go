package main

import (
	"context"
	"io"
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/joho/godotenv"

	"babyagent/ch04"
	"babyagent/ch04/tool"
	"babyagent/ch04/tui"
	"babyagent/shared"
)

func main() {
	_ = godotenv.Load()

	ctx := context.Background()
	modelConf := shared.NewModelConfig()
	// 先拿到现有什么mcp，相当于一个通讯录
	mcpServerMap, err := shared.LoadMcpServerConfig("mcp-server.json")
	if err != nil {
		log.Printf("Failed to load MCP server configuration: %v", err)
	}
	mcpClients := make([]*ch04.McpClient, 0)
	// 将网络或者本地发现的mcp，注册到本agent
	for k, v := range mcpServerMap {
		mcpClient := ch04.NewMcpToolProvider(k, v)
		if err := mcpClient.RefreshTools(ctx); err != nil {
			log.Printf("Failed to refresh tools for MCP server %s: %v", k, err)
			continue
		}
		mcpClients = append(mcpClients, mcpClient)
	}

	agent := ch04.NewAgent(
		modelConf,
		ch04.CodingAgentSystemPrompt,
		[]tool.Tool{tool.NewBashTool()},
		mcpClients,
	)

	log.SetOutput(io.Discard)
	p := tea.NewProgram(tui.NewModel(agent, modelConf.Model))
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
