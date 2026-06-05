package logic

import (
	"context"

	"agentmesh/agent-svc/model"
)

type ConversationLogic struct {
	conversationModel *model.ConversationModel
	messageModel      *model.MessageModel
}

func NewConversationLogic() *ConversationLogic {
	return &ConversationLogic{
		conversationModel: model.NewConversationModel(),
		messageModel:      model.NewMessageModel(),
	}
}

// CreateConversation creates a new conversation
func (l *ConversationLogic) CreateConversation(ctx context.Context, tenantID, userID, title string) (*model.Conversation, error) {
	conv := &model.Conversation{
		TenantID: tenantID,
		UserID:   userID,
		Title:    title,
	}

	if err := l.conversationModel.Insert(ctx, conv); err != nil {
		return nil, err
	}

	return conv, nil
}

// ListConversations returns all conversations for a tenant and user
func (l *ConversationLogic) ListConversations(ctx context.Context, tenantID, userID string) ([]*model.Conversation, error) {
	return l.conversationModel.FindAllByTenantAndUser(ctx, tenantID, userID)
}

// GetMessages returns all messages for a conversation
func (l *ConversationLogic) GetMessages(ctx context.Context, conversationID string) ([]*model.Message, error) {
	return l.messageModel.FindAllByConversation(ctx, conversationID)
}
