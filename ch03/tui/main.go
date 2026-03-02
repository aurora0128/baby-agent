package main

import (
	"io"
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/joho/godotenv"

	"babyagent/ch03"
	"babyagent/ch03/tool"
	"babyagent/shared"
)

func main() {
	_ = godotenv.Load()

	modelConf := shared.NewModelConfig()

	agent := ch03.NewAgent(
		modelConf,
		ch03.CodingAgentSystemPrompt,
		[]tool.Tool{tool.NewBashTool()},
	)

	log.SetOutput(io.Discard)
	p := tea.NewProgram(newModel(agent, modelConf.Model))
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
