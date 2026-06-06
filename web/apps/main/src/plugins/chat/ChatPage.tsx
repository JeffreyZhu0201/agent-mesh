import React, { useState, useCallback, useRef, useEffect } from 'react';
import { Box, Typography, Select, MenuItem, FormControl, SelectChangeEvent } from '@mui/material';
import { ChatSidebar } from '@agentmesh/ui';
import { MessageList, MessageItem } from './components/chat/MessageList';
import { ChatComposer } from './components/chat/ChatComposer';
import { WelcomeScreen } from './components/chat/WelcomeScreen';

export interface Conversation {
  id: string;
  title: string;
  preview?: string;
  updatedAt: Date;
  model?: string;
  messages: MessageItem[];
}

const AVAILABLE_MODELS = [
  { id: 'doubao-seed-code', name: 'Doubao Seed Code' },
  { id: 'doubao-seed-2.0-code', name: 'Doubao Seed 2.0 Code' },
  { id: 'doubao-seed-2.0-pro', name: 'Doubao Seed 2.0 Pro' },
  { id: 'doubao-seed-2.0-lite', name: 'Doubao Seed 2.0 Lite' },
  { id: 'minimax-m2.7', name: 'MiniMax M2.7' },
  { id: 'glm-5.1', name: 'GLM-5.1' },
  { id: 'kimi-k2.6', name: 'Kimi K2.6' },
  { id: 'deepseek-v4-pro', name: 'DeepSeek V4 Pro' },
  { id: 'deepseek-v4-flash', name: 'DeepSeek V4 Flash' },
];

const DEFAULT_MODEL = 'doubao-seed-code';
const API_BASE = 'http://localhost:8080/api';
const generateId = () => Math.random().toString(36).substring(2, 11);

// Stream chat response from backend via SSE-over-fetch
async function streamChat(
  conversationId: string,
  content: string,
  model: string,
  onChunk: (chunk: string) => void,
  onDone: (messageId: string) => void,
  onError: (msg: string) => void,
  signal?: AbortSignal
) {
  const res = await fetch(`${API_BASE}/agent/chat`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ conversationId, content, model }),
    signal,
  });

  if (!res.ok || !res.body) {
    onError(`Request failed: ${res.status}`);
    return;
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';

  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });

    const events = buffer.split('\n\n');
    buffer = events.pop() || '';

    for (const evt of events) {
      const line = evt.split('\n').find((l) => l.startsWith('data: '));
      if (!line) continue;
      try {
        const data = JSON.parse(line.slice(6));
        if (data.type === 'chunk') onChunk(data.content);
        else if (data.type === 'done') onDone(data.messageId);
        else if (data.type === 'error') onError(data.message);
      } catch {
        // ignore parse errors
      }
    }
  }
}

const Chat: React.FC = () => {
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [selectedConversationId, setSelectedConversationId] = useState<string>('');
  const [selectedModel, setSelectedModel] = useState(DEFAULT_MODEL);
  const [isStreaming, setIsStreaming] = useState(false);
  const [streamingMessageId, setStreamingMessageId] = useState<string | null>(null);

  const messagesEndRef = useRef<HTMLDivElement>(null);

  const selectedConversation = conversations.find((c) => c.id === selectedConversationId);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [selectedConversation?.messages, isStreaming]);

  // Load conversations from backend on mount
  useEffect(() => {
    (async () => {
      try {
        const res = await fetch(`${API_BASE}/agent/conversations`);
        if (!res.ok) return;
        const data = await res.json();
        const list: Conversation[] = (data.conversations || []).map((c: any) => ({
          id: c.id,
          title: c.title,
          preview: c.preview,
          updatedAt: new Date(c.updatedAt),
          model: c.model,
          messages: (c.messages || []).map((m: any) => ({
            id: m.id,
            role: m.role,
            content: m.content,
            model: m.model,
            timestamp: new Date(m.timestamp),
          })),
        }));
        list.sort((a, b) => b.updatedAt.getTime() - a.updatedAt.getTime());
        setConversations(list);
        if (list[0]) setSelectedConversationId(list[0].id);
      } catch (err) {
        console.error('Failed to load conversations', err);
      }
    })();
  }, []);

  const createConversation = useCallback(async (): Promise<string | null> => {
    try {
      const res = await fetch(`${API_BASE}/agent/conversations`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title: 'New Chat', model: selectedModel }),
      });
      const conv = await res.json();
      const newConversation: Conversation = {
        id: conv.id,
        title: conv.title,
        updatedAt: new Date(conv.updatedAt),
        model: conv.model,
        messages: [],
      };
      setConversations((prev) => [newConversation, ...prev]);
      setSelectedConversationId(newConversation.id);
      return conv.id;
    } catch (err) {
      console.error('Failed to create conversation', err);
      return null;
    }
  }, [selectedModel]);

  const handleNewChat = useCallback(() => {
    void createConversation();
  }, [createConversation]);

  const handleDeleteConversation = useCallback(
    async (id: string) => {
      try {
        await fetch(`${API_BASE}/agent/conversations/${id}`, { method: 'DELETE' });
      } catch (err) {
        console.error('Failed to delete conversation', err);
      }
      setConversations((prev) => prev.filter((c) => c.id !== id));
      if (selectedConversationId === id) {
        setSelectedConversationId((prev) => {
          const remaining = conversations.filter((c) => c.id !== id);
          return remaining[0]?.id || '';
        });
      }
    },
    [selectedConversationId, conversations]
  );

  const handleModelChange = useCallback((event: SelectChangeEvent<string>) => {
    setSelectedModel(event.target.value);
  }, []);

  const sendMessage = useCallback(
    async (content: string) => {
      // Auto-create a conversation if none selected
      let convId = selectedConversationId;
      if (!convId) {
        const newId = await createConversation();
        if (!newId) return;
        convId = newId;
      }

      const userMessage: MessageItem = {
        id: generateId(),
        role: 'user',
        content,
        timestamp: new Date(),
      };

      setConversations((prev) =>
        prev.map((c) =>
          c.id === convId
            ? {
                ...c,
                messages: [...c.messages, userMessage],
                updatedAt: new Date(),
                preview: content.substring(0, 50),
                title: c.title === 'New Chat' || !c.title ? content.substring(0, 40) : c.title,
              }
            : c
        )
      );

      setIsStreaming(true);
      const assistantMessageId = generateId();
      setStreamingMessageId(assistantMessageId);

      setConversations((prev) =>
        prev.map((c) =>
          c.id === convId
            ? {
                ...c,
                messages: [
                  ...c.messages,
                  {
                    id: assistantMessageId,
                    role: 'assistant' as const,
                    content: '',
                    model: selectedModel,
                    timestamp: new Date(),
                  },
                ],
              }
            : c
        )
      );

      let fullResponse = '';
      try {
        await streamChat(
          convId,
          content,
          selectedModel,
          (chunk) => {
            fullResponse += chunk;
            setConversations((prev) =>
              prev.map((c) =>
                c.id !== convId
                  ? c
                  : {
                      ...c,
                      messages: c.messages.map((m) =>
                        m.id === assistantMessageId ? { ...m, content: fullResponse } : m
                      ),
                    }
              )
            );
          },
          () => {
            setIsStreaming(false);
            setStreamingMessageId(null);
          },
          (msg) => {
            setConversations((prev) =>
              prev.map((c) =>
                c.id !== convId
                  ? c
                  : {
                      ...c,
                      messages: c.messages.map((m) =>
                        m.id === assistantMessageId ? { ...m, content: `⚠️ Error: ${msg}` } : m
                      ),
                    }
              )
            );
            setIsStreaming(false);
            setStreamingMessageId(null);
          }
        );
      } catch (err) {
        console.error('Chat stream failed', err);
        setIsStreaming(false);
        setStreamingMessageId(null);
      }
    },
    [selectedConversationId, selectedModel, createConversation]
  );

  const messages = selectedConversation?.messages || [];
  const showWelcome = messages.length === 0;

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
        onSelectConversation={setSelectedConversationId}
        onNewChat={handleNewChat}
        onDeleteConversation={handleDeleteConversation}
      />

      {/* Main chat area */}
      <Box
        sx={{
          flexGrow: 1,
          display: 'flex',
          flexDirection: 'column',
          bgcolor: 'background.default',
          minWidth: 0,
        }}
      >
        {/* Top bar */}
        <Box
          sx={{
            px: 3,
            py: 1.5,
            borderBottom: 1,
            borderColor: 'divider',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            bgcolor: 'background.paper',
          }}
        >
          <Typography variant="subtitle1" sx={{ fontWeight: 600, color: 'text.primary' }}>
            {selectedConversation?.title || 'New Chat'}
          </Typography>
          <FormControl size="small" sx={{ minWidth: 160 }}>
            <Select
              value={selectedModel}
              onChange={handleModelChange}
              disabled={isStreaming}
              sx={{ fontSize: 14, borderRadius: 2 }}
            >
              {AVAILABLE_MODELS.map((model) => (
                <MenuItem key={model.id} value={model.id} sx={{ fontSize: 14 }}>
                  {model.name}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        </Box>

        {/* Messages area (or welcome screen) */}
        <Box sx={{ flexGrow: 1, overflow: 'auto', minHeight: 0 }}>
          {showWelcome ? (
            <WelcomeScreen onSendPrompt={sendMessage} />
          ) : (
            <>
              <MessageList messages={messages} streamingMessageId={streamingMessageId} />
              <Box ref={messagesEndRef} />
            </>
          )}
        </Box>

        {/* Composer */}
        <ChatComposer
          onSend={sendMessage}
          disabled={isStreaming}
          placeholder={isStreaming ? 'Waiting for response...' : 'Message AgentMesh...'}
        />
      </Box>
    </Box>
  );
};

export default Chat;
