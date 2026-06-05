import React, { useState, useCallback, useRef, useEffect } from 'react';
import {
  Box,
  Typography,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Paper,
  CircularProgress,
  SelectChangeEvent,
} from '@mui/material';
import { ChatSidebar, ChatInput, MessageBubble } from '@agentmesh/ui';

export interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  model?: string;
  timestamp: Date;
}

export interface Conversation {
  id: string;
  title: string;
  preview?: string;
  updatedAt: Date;
  model?: string;
  messages: Message[];
}

const AVAILABLE_MODELS = [
  { id: 'gpt-4', name: 'GPT-4' },
  { id: 'gpt-3.5-turbo', name: 'GPT-3.5 Turbo' },
  { id: 'claude-3-opus', name: 'Claude-3 Opus' },
  { id: 'claude-3-sonnet', name: 'Claude-3 Sonnet' },
];

const generateId = () => Math.random().toString(36).substring(2, 11);

// Mock streaming response generator
const generateStreamingResponse = async (
  message: string,
  onChunk: (chunk: string) => void,
  onComplete: () => void
) => {
  const responses: Record<string, string> = {
    default: `I understand you said "${message}". This is a simulated streaming response to demonstrate the chat UI with real-time message streaming. The response appears character by character to simulate an actual LLM response stream.`,
    hello: 'Hello! How can I assist you today? Feel free to ask me anything about the AgentMesh platform or any other questions you might have.',
    help: 'I can help you with various tasks! I can:\n\n1. Answer questions about your projects\n2. Write and debug code\n3. Search through documentation\n4. And much more!\n\nWhat would you like help with?',
  };

  const response =
    responses[message.toLowerCase()] || responses.default;

  for (let i = 0; i < response.length; i++) {
    await new Promise((resolve) =>
      setTimeout(resolve, 20 + Math.random() * 30)
    );
    onChunk(response[i]);
  }

  onComplete();
};

const Chat: React.FC = () => {
  const [conversations, setConversations] = useState<Conversation[]>([
    {
      id: generateId(),
      title: 'Welcome Chat',
      preview: 'Hello! How can I help?',
      updatedAt: new Date(),
      model: 'gpt-4',
      messages: [
        {
          id: generateId(),
          role: 'assistant',
          content:
            'Hello! Welcome to AgentMesh. I am your AI assistant. How can I help you today?',
          model: 'gpt-4',
          timestamp: new Date(),
        },
      ],
    },
  ]);

  const [selectedConversationId, setSelectedConversationId] = useState<string>(
    conversations[0]?.id || ''
  );
  const [selectedModel, setSelectedModel] = useState('gpt-4');
  const [isStreaming, setIsStreaming] = useState(false);
  const [streamingMessageId, setStreamingMessageId] = useState<string | null>(
    null
  );

  const messagesEndRef = useRef<HTMLDivElement>(null);

  const selectedConversation = conversations.find(
    (c) => c.id === selectedConversationId
  );

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [selectedConversation?.messages, isStreaming]);

  const handleNewChat = useCallback(() => {
    const newConversation: Conversation = {
      id: generateId(),
      title: 'New Chat',
      updatedAt: new Date(),
      model: selectedModel,
      messages: [],
    };

    setConversations((prev) => [newConversation, ...prev]);
    setSelectedConversationId(newConversation.id);
  }, [selectedModel]);

  const handleSelectConversation = useCallback((id: string) => {
    setSelectedConversationId(id);
  }, []);

  const handleSearch = useCallback((query: string) => {
    // Filter conversations based on search query
    console.log('Searching for:', query);
  }, []);

  const handleModelChange = useCallback((event: SelectChangeEvent<string>) => {
    setSelectedModel(event.target.value);
    setConversations((prev) =>
      prev.map((c) =>
        c.id === selectedConversationId ? { ...c, model: event.target.value } : c
      )
    );
  }, [selectedConversationId]);

  const handleSendMessage = useCallback(
    (content: string) => {
      if (!selectedConversationId) return;

      const userMessage: Message = {
        id: generateId(),
        role: 'user',
        content,
        timestamp: new Date(),
      };

      // Add user message
      setConversations((prev) =>
        prev.map((c) =>
          c.id === selectedConversationId
            ? {
                ...c,
                messages: [...c.messages, userMessage],
                updatedAt: new Date(),
                preview: content.substring(0, 50),
              }
            : c
        )
      );

      // Start streaming response
      setIsStreaming(true);
      const assistantMessageId = generateId();
      setStreamingMessageId(assistantMessageId);

      let fullResponse = '';

      generateStreamingResponse(
        content,
        (chunk) => {
          fullResponse += chunk;
          // Update with streaming content
          setConversations((prev) =>
            prev.map((c) => {
              if (c.id !== selectedConversationId) return c;
              const existingStreamingIndex = c.messages.findIndex(
                (m) => m.id === assistantMessageId
              );
              if (existingStreamingIndex >= 0) {
                const updatedMessages = [...c.messages];
                updatedMessages[existingStreamingIndex] = {
                  ...updatedMessages[existingStreamingIndex],
                  content: fullResponse,
                };
                return { ...c, messages: updatedMessages };
              } else {
                return {
                  ...c,
                  messages: [
                    ...c.messages,
                    {
                      id: assistantMessageId,
                      role: 'assistant' as const,
                      content: fullResponse,
                      model: selectedModel,
                      timestamp: new Date(),
                    },
                  ],
                };
              }
            })
          );
        },
        () => {
          // Complete
          setIsStreaming(false);
          setStreamingMessageId(null);
        }
      );
    },
    [selectedConversationId, selectedModel]
  );

  return (
    <Box sx={{ display: 'flex', height: 'calc(100vh - 64px)' }}>
      {/* Sidebar */}
      <ChatSidebar
        conversations={conversations.map((c) => ({
          id: c.id,
          title: c.title,
          preview: c.preview,
          updatedAt: c.updatedAt,
          model: c.model,
        }))}
        selectedId={selectedConversationId}
        onSelectConversation={handleSelectConversation}
        onNewChat={handleNewChat}
        onSearch={handleSearch}
      />

      {/* Main Chat Area */}
      <Box
        sx={{
          flexGrow: 1,
          display: 'flex',
          flexDirection: 'column',
          bgcolor: 'background.default',
        }}
      >
        {/* Header with Model Selection */}
        <Box
          sx={{
            p: 2,
            borderBottom: 1,
            borderColor: 'divider',
            bgcolor: 'background.paper',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <Typography variant="h6" sx={{ fontWeight: 600 }}>
            {selectedConversation?.title || 'Chat'}
          </Typography>
          <FormControl size="small" sx={{ minWidth: 180 }}>
            <InputLabel>Model</InputLabel>
            <Select
              value={selectedModel}
              onChange={handleModelChange}
              label="Model"
              disabled={isStreaming}
            >
              {AVAILABLE_MODELS.map((model) => (
                <MenuItem key={model.id} value={model.id}>
                  {model.name}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        </Box>

        {/* Messages */}
        <Box
          sx={{
            flexGrow: 1,
            overflow: 'auto',
            py: 2,
          }}
        >
          {selectedConversation?.messages.map((message) => (
            <MessageBubble
              key={message.id}
              content={message.content}
              role={message.role}
              model={message.model}
              timestamp={message.timestamp}
              isStreaming={message.id === streamingMessageId && isStreaming}
            />
          ))}
          <div ref={messagesEndRef} />
        </Box>

        {/* Input */}
        <ChatInput
          onSend={handleSendMessage}
          disabled={isStreaming}
          placeholder={
            isStreaming
              ? 'Waiting for response...'
              : 'Type a message... (Enter to send, Shift+Enter for newline)'
          }
        />
      </Box>
    </Box>
  );
};

export default Chat;