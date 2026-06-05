package logic

import (
	"context"
	"fmt"
	"io"
	"strings"

	"agentmesh/agent-svc/internal/config"
	"agentmesh/agent-svc/model"

	einmodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

type ChatLogic struct {
	conversationModel *model.ConversationModel
	messageModel      *model.MessageModel
	chatModel         einmodel.ChatModel
	cfg               *config.Config
}

func NewChatLogic(cfg *config.Config) *ChatLogic {
	l := &ChatLogic{
		conversationModel: model.NewConversationModel(),
		messageModel:      model.NewMessageModel(),
		cfg:               cfg,
	}

	// Initialize Eino chat model based on config
	if cfg != nil {
		l.initChatModel(context.Background())
	}

	return l
}

// initChatModel initializes the Eino ChatModel based on configuration
func (l *ChatLogic) initChatModel(ctx context.Context) error {
	var chatModel einmodel.ChatModel
	var err error

	provider := l.cfg.LLM.Provider
	if provider == "" {
		provider = "openai" // default provider
	}

	switch provider {
	case "claude":
		chatModel, err = l.initClaudeModel(ctx)
	case "openai":
		chatModel, err = l.initOpenAIModel(ctx)
	default:
		return fmt.Errorf("unsupported LLM provider: %s", provider)
	}

	if err != nil {
		return fmt.Errorf("failed to initialize chat model: %w", err)
	}

	l.chatModel = chatModel
	return nil
}

// initOpenAIModel initializes OpenAI chat model
func (l *ChatLogic) initOpenAIModel(ctx context.Context) (einmodel.ChatModel, error) {
	cfg := &openai.ChatModelConfig{
		APIKey:  l.cfg.LLM.OpenAI.APIKey,
		Model:   l.cfg.LLM.OpenAI.Model,
		BaseURL: l.cfg.LLM.OpenAI.Endpoint,
	}

	if l.cfg.LLM.Temperature > 0 {
		cfg.Temperature = &l.cfg.LLM.Temperature
	}
	if l.cfg.LLM.MaxTokens > 0 {
		cfg.MaxTokens = &l.cfg.LLM.MaxTokens
	}

	return openai.NewChatModel(ctx, cfg)
}

// initClaudeModel initializes Claude chat model
func (l *ChatLogic) initClaudeModel(ctx context.Context) (einmodel.ChatModel, error) {
	cfg := &claude.Config{
		APIKey:    l.cfg.LLM.Claude.APIKey,
		Model:     l.cfg.LLM.Claude.Model,
		MaxTokens: l.cfg.LLM.MaxTokens,
	}

	if l.cfg.LLM.Temperature > 0 {
		cfg.Temperature = &l.cfg.LLM.Temperature
	}
	if l.cfg.LLM.Claude.Endpoint != "" {
		cfg.BaseURL = &l.cfg.LLM.Claude.Endpoint
	}

	return claude.NewChatModel(ctx, cfg)
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

	// Call Eino for AI response
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
func (l *ChatLogic) callEino(ctx context.Context, tenantID, userID, conversationID, userMessage string) (string, error) {
	// If no chat model is initialized, return mock response
	if l.chatModel == nil {
		return l.mockResponse(), nil
	}

	// Build messages for Eino
	messages := []*schema.Message{
		{
			Role:    schema.User,
			Content: userMessage,
		},
	}

	// Generate response using Eino
	resp, err := l.chatModel.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("Eino Generate failed: %w", err)
	}

	return resp.Content, nil
}

// mockResponse returns a mock response when no LLM is configured
func (l *ChatLogic) mockResponse() string {
	return "Hello! I'm your AI assistant. How can I help you today?"
}

// SetChatModel allows setting a custom chat model (useful for testing)
func (l *ChatLogic) SetChatModel(model einmodel.ChatModel) {
	l.chatModel = model
}
