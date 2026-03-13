package tool

import (
	"context"
	"encoding/json"
	"os/exec"
	"runtime"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

type BashTool struct{}

func NewBashTool() *BashTool {
	return &BashTool{}
}

type BashToolParam struct {
	Command string `json:"command"`
}

func (t *BashTool) ToolName() AgentTool {
	return AgentToolBash
}

/*
详细描述每个工具需要什么字段
*/
func (t *BashTool) Info() openai.ChatCompletionToolUnionParam {
	return openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
		Name: string(AgentToolBash),
		/*
			本质是返回了一个指针
			目的是为了区分字符串的特殊情况 “”传了但为空值 以及没传的情况
			desc := "execute bash command"
			Description: &desc
			Name 是必填，直接给字符串就行（类型就是 string）
			Description 是可选，SDK 设计成 *string，所以用 openai.String(...) 最省事
		*/
		Description: openai.String("execute bash command"),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "the bash command to execute",
				},
			},
			"required": []string{"command"},
		},
	})
}

func (t *BashTool) Execute(ctx context.Context, argumentsInJSON string) (string, error) {
	p := BashToolParam{}
	err := json.Unmarshal([]byte(argumentsInJSON), &p)
	if err != nil {
		return "", err
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Windows: use cmd.exe to interpret the command line
		cmd = exec.CommandContext(ctx, "cmd", "/C", p.Command)
	} else {
		// Linux/macOS: use POSIX sh (more universal than assuming bash exists)
		cmd = exec.CommandContext(ctx, "sh", "-c", p.Command)

	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
