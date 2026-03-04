package context

import (
	"log"

	"github.com/tiktoken-go/tokenizer"

	"babyagent/shared"
)

var tokenEnc tokenizer.Codec

func init() {
	var err error
	tokenEnc, err = tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		log.Fatal(err)
	}
}

func CountTokens(message shared.OpenAIMessage) int {
	contentAny := message.GetContent().AsAny()
	switch contentAny.(type) {
	case *string:
		count, _ := tokenEnc.Count(*contentAny.(*string))
		return count
	}
	return 0
}

// GetRoleName 从消息中获取角色名称（不依赖 GetRole()）
func GetRoleName(message shared.OpenAIMessage) string {
	if message.OfSystem != nil {
		return "system"
	}
	if message.OfUser != nil {
		return "user"
	}
	if message.OfAssistant != nil {
		return "assistant"
	}
	if message.OfTool != nil {
		return "tool"
	}
	if message.OfDeveloper != nil {
		return "developer"
	}
	if message.OfFunction != nil {
		return "function"
	}
	return "unknown"
}
