package prompt

import (
	"errors"
	"fmt"
	"strings"
)

type Builder interface {
	CompileText(params map[string]string) (string, error)
	CompileChat(params map[string]string) (PromptChatList, error)
}

type promptBuilder struct {
	activePromptChat []PromptChat
	activePromptText string
}

func newBuilder(activePromptChat []PromptChat, activePromptText string) Builder {
	return &promptBuilder{
		activePromptChat: activePromptChat,
		activePromptText: activePromptText,
	}
}

func (b *promptBuilder) CompileText(params map[string]string) (string, error) {
	if b.activePromptText == "" {
		return "", errors.New("active prompt text is empty")
	}

	result := b.activePromptText
	usedVars := make(map[string]struct{})

	for key, value := range params {
		promptVar := fmt.Sprintf("{{%s}}", key)

		if strings.Contains(result, promptVar) {
			usedVars[key] = struct{}{}
		}

		result = strings.ReplaceAll(result, promptVar, value)
	}

	// Validate unused params.
	for key := range params {
		if _, ok := usedVars[key]; !ok {
			return "", fmt.Errorf("prompt variable with key %s does not exist", key)
		}
	}

	return result, nil
}

func (b *promptBuilder) CompileChat(params map[string]string) (PromptChatList, error) {
	if len(b.activePromptChat) == 0 {
		return nil, errors.New("active prompt chat is empty")
	}

	result := make([]PromptChat, len(b.activePromptChat))
	copy(result, b.activePromptChat)

	usedVars := make(map[string]struct{})

	for i, chat := range result {
		content := chat.Content

		for key, value := range params {
			promptVar := fmt.Sprintf("{{%s}}", key)

			if strings.Contains(content, promptVar) {
				usedVars[key] = struct{}{}
			}

			content = strings.ReplaceAll(content, promptVar, value)
		}

		result[i].Content = content
	}

	// Validate unused params.
	for key := range params {
		if _, ok := usedVars[key]; !ok {
			return nil, fmt.Errorf("prompt variable with key %s does not exist", key)
		}
	}

	return result, nil
}
