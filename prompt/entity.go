package prompt

type PromptChatList []PromptChat

var (
	RoleSystem string = "system"
	RoleUser   string = "user"
)

type PromptResponse[V string | PromptChatList] struct {
	ID              string   `json:"id"`
	CreatedAt       string   `json:"createdAt"`
	UpdatedAt       string   `json:"updatedAt"`
	ProjectId       string   `json:"projectId"`
	CreatedBy       string   `json:"createdBy"`
	Prompt          V        `json:"prompt"`
	Name            string   `json:"name"`
	Version         int      `json:"version"`
	Type            string   `json:"type"`
	IsActive        bool     `json:"isActive"`
	Config          any      `json:"config"`
	Tags            []string `json:"tags"`
	Labels          []string `json:"labels"`
	CommitMessage   any      `json:"commitMessage"`
	ResolutionGraph any      `json:"resolutionGraph"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type PromptChat struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Careful as this return the first system message found.
// If your prompt contains more than 1 system message. This method is not suitable.
func (p PromptChatList) GetSystemMessage() string {
	if len(p) == 0 {
		return ""
	}

	for _, chat := range p {
		if chat.Role == RoleSystem {
			return chat.Content
		}
	}

	return ""
}

// Careful as this return the first user message found.
// If your prompt contains more than 1 user message. This method is not suitable.
func (p PromptChatList) GetUserMessage() string {
	if len(p) == 0 {
		return ""
	}

	for _, chat := range p {
		if chat.Role == RoleUser {
			return chat.Content
		}
	}

	return ""
}

type PromptType int

const (
	PromptTypeChat = iota
	PromptTypeText
)

type Prompt struct {
	Type    PromptType `json:"type"`
	Role    string     `json:"role"`
	Content string     `json:"content"`
}

type PromptSpec struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	Prompt        []Prompt `json:"prompt"`
	CommitMessage *string  `json:"commitMessage,omitempty"`
	Labels        []string `json:"labels,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

type LangfuseSpec struct {
	PublicKey  string
	PrivateKey string
	BaseUrl    string
}
