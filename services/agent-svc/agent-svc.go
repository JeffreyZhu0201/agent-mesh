package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var configFile = flag.String("f", "agent-svc.yaml", "the config file")

// Message in a conversation
type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // user | assistant | system
	Content   string    `json:"content"`
	Model     string    `json:"model,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Conversation holds the chat history
type Conversation struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Preview   string    `json:"preview,omitempty"`
	Model     string    `json:"model,omitempty"`
	UserID    uint      `json:"userId"`
	TenantID  uint      `json:"tenantId"`
	Messages  []Message `json:"messages"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// In-memory store keyed by conversation ID
var (
	conversations = make(map[string]*Conversation)
	convMu        sync.RWMutex
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// buildLLMModel creates a fresh Eino OpenAI-compatible chat model.
// Returns nil if no API key (mock mode).
func buildLLMModel(ctx context.Context) (model.ChatModel, error) {
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		// Backward-compat with ANTHROPIC_API_KEY env var name
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey == "" {
		return nil, nil
	}
	modelName := getEnv("LLM_MODEL", getEnv("ANTHROPIC_MODEL", "gpt-4o-mini"))
	baseURL := getEnv("LLM_BASE_URL", os.Getenv("ANTHROPIC_BASE_URL"))

	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		Model:   modelName,
		BaseURL: baseURL,
	})
}

func main() {
	flag.Parse()

	r := gin.Default()

	// (CORS is handled by the api-gateway; setting it here too causes duplicate headers.)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// List conversations (optionally filter by user)
	r.GET("/api/agent/conversations", func(c *gin.Context) {
		convMu.RLock()
		defer convMu.RUnlock()
		result := make([]*Conversation, 0, len(conversations))
		for _, conv := range conversations {
			result = append(result, conv)
		}
		c.JSON(http.StatusOK, gin.H{"conversations": result})
	})

	// Create a new conversation
	r.POST("/api/agent/conversations", func(c *gin.Context) {
		var req struct {
			Title string `json:"title"`
			Model string `json:"model"`
		}
		_ = c.ShouldBindJSON(&req)
		if req.Title == "" {
			req.Title = "New Chat"
		}
		if req.Model == "" {
			req.Model = getEnv("ANTHROPIC_MODEL", "claude-3-5-sonnet-20241022")
		}
		conv := &Conversation{
			ID:        uuid.New().String(),
			Title:     req.Title,
			Model:     req.Model,
			Messages:  []Message{},
			UpdatedAt: time.Now(),
		}
		convMu.Lock()
		conversations[conv.ID] = conv
		convMu.Unlock()
		c.JSON(http.StatusCreated, conv)
	})

	// Get a single conversation
	r.GET("/api/agent/conversations/:id", func(c *gin.Context) {
		convMu.RLock()
		conv, ok := conversations[c.Param("id")]
		convMu.RUnlock()
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"message": "conversation not found"})
			return
		}
		c.JSON(http.StatusOK, conv)
	})

	// Delete a conversation
	r.DELETE("/api/agent/conversations/:id", func(c *gin.Context) {
		convMu.Lock()
		delete(conversations, c.Param("id"))
		convMu.Unlock()
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	})

	// Chat endpoint with SSE streaming
	// POST /api/agent/chat  {conversationId, content, model}
	r.POST("/api/agent/chat", handleChat)

	fmt.Println("Starting Agent Service at port 8082...")
	r.Run(":8082")
}

// handleChat appends a user message and streams an assistant reply via SSE.
// Events: {type:"chunk", content:"..."}  {type:"done", messageId:"..."}
func handleChat(c *gin.Context) {
	var req struct {
		ConversationID string `json:"conversationId" binding:"required"`
		Content        string `json:"content" binding:"required"`
		Model          string `json:"model"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request: " + err.Error()})
		return
	}

	convMu.Lock()
	conv, ok := conversations[req.ConversationID]
	if !ok {
		convMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"message": "conversation not found"})
		return
	}
	// Append user message
	userMsg := Message{
		ID:        uuid.New().String(),
		Role:      "user",
		Content:   req.Content,
		Timestamp: time.Now(),
	}
	conv.Messages = append(conv.Messages, userMsg)
	conv.UpdatedAt = time.Now()
	if conv.Title == "New Chat" || conv.Title == "" {
		conv.Title = truncate(req.Content, 40)
	}
	conv.Preview = truncate(req.Content, 60)

	// Build the message history for the model
	history := make([]*schema.Message, 0, len(conv.Messages))
	for _, m := range conv.Messages {
		var role schema.RoleType
		switch m.Role {
		case "user":
			role = schema.User
		case "assistant":
			role = schema.Assistant
		case "system":
			role = schema.System
		default:
			role = schema.User
		}
		history = append(history, &schema.Message{Role: role, Content: m.Content})
	}
	modelName := req.Model
	if modelName == "" {
		modelName = conv.Model
	}
	convMu.Unlock()

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "streaming unsupported"})
		return
	}

	assistantMsgID := uuid.New().String()
	var fullReply string

	send := func(event map[string]any) {
		data, _ := json.Marshal(event)
		fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		flusher.Flush()
	}

	ctx := c.Request.Context()
	cm, err := buildLLMModel(ctx)
	if err != nil {
		send(map[string]any{"type": "error", "message": "model init failed: " + err.Error()})
		return
	}

	if cm == nil {
		// Mock mode fallback when no API key is configured
		mockReply := fmt.Sprintf("(mock) Set ANTHROPIC_API_KEY to enable real responses. You said: %s", req.Content)
		for _, ch := range mockReply {
			send(map[string]any{"type": "chunk", "content": string(ch)})
			fullReply += string(ch)
			time.Sleep(15 * time.Millisecond)
		}
	} else {
		stream, err := cm.Stream(ctx, history)
		if err != nil {
			send(map[string]any{"type": "error", "message": err.Error()})
			return
		}
		defer stream.Close()
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				send(map[string]any{"type": "error", "message": err.Error()})
				return
			}
			if msg.Content != "" {
				fullReply += msg.Content
				send(map[string]any{"type": "chunk", "content": msg.Content})
			}
		}
	}

	// Persist assistant message
	convMu.Lock()
	if conv2, ok := conversations[req.ConversationID]; ok {
		conv2.Messages = append(conv2.Messages, Message{
			ID:        assistantMsgID,
			Role:      "assistant",
			Content:   fullReply,
			Model:     modelName,
			Timestamp: time.Now(),
		})
		conv2.UpdatedAt = time.Now()
	}
	convMu.Unlock()

	send(map[string]any{"type": "done", "messageId": assistantMsgID})
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
