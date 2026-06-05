package logic

import (
	"context"
	"io"
	"strings"

	"agentmesh/agent-svc/model"
)

type ChatLogic struct {
	conversationModel *model.ConversationModel
	messageModel      *model.MessageModel
}

func NewChatLogic() *ChatLogic {
	return &ChatLogic{
		conversationModel: model.NewConversationModel(),
		messageModel:      model.NewMessageModel(),
	}
}

// ChatStream handles the chat streaming logic
func (l *ChatLogic) ChatStream(ctx context.Context, tenantID, userID, conversationID, userMessage string, stream chan string) error {
	// Save user message
	userMsg := &model.Message{
		ConversationID: conversationID,
		Role:            "user",
		Content:         userMessage,
	}
	if err := l.messageModel.Insert(ctx, userMsg); err != nil {
		return err
	}

	// Call Eino for AI response (mock implementation)
	aiResponse, err := l.callEino(ctx, tenantID, userID, conversationID, userMessage)
	if err != nil {
		return err
	}

	// Stream the response
	reader := strings.NewReader(aiResponse)
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			stream <- string(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	// Save AI response
	assistantMsg := &model.Message{
		ConversationID: conversationID,
		Role:            "assistant",
		Content:         aiResponse,
	}
	if err := l.messageModel.Insert(ctx, assistantMsg); err != nil {
		return err
	}

	close(stream)
	return nil
}

// callEino calls the Eino SDK for AI response
// This is a mock implementation for now
func (l *ChatLogic) callEino(ctx context.Context, tenantID, userID, conversationID, userMessage string) (string, error) {
	// TODO: Replace with actual Eino SDK integration
	// Mock response for demonstration purposes
	mockResponses := []string{
		"Hello! I'm your AI assistant. ",
		"I'm here to help you with any questions you might have. ",
		"How can I assist you today?",
	}

	var response strings.Builder
	for _, msg := range mockResponses {
		response.WriteString(msg)
	}

	return response.String(), nil
}
