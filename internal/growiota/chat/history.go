package chat

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	Role    Role
	Content string
}

type History struct {
	Messages []Message
}

func NewHistory() *History {
	return &History{
		Messages: make([]Message, 0),
	}
}

func (h *History) AddMessage(role Role, content string) {
	h.Messages = append(h.Messages, Message{
		Role:    role,
		Content: content,
	})
}
