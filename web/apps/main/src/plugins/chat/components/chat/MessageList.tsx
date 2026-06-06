import React, { memo } from 'react';
import { Box, Avatar, IconButton, Tooltip } from '@mui/material';
import {
  Person as PersonIcon,
  SmartToy as BotIcon,
  ContentCopy as CopyIcon,
} from '@mui/icons-material';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism';

export interface MessageItem {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  model?: string;
  timestamp: Date;
}

interface MessageProps {
  message: MessageItem;
  isStreaming?: boolean;
}

const CodeBlock: React.FC<{
  inline?: boolean;
  className?: string;
  children?: React.ReactNode;
}> = ({ inline, className, children }) => {
  const match = /language-(\w+)/.exec(className || '');
  const codeString = String(children).replace(/\n$/, '');

  if (inline) {
    return (
      <code
        style={{
          background: 'rgba(135, 131, 120, 0.15)',
          color: '#eb5757',
          padding: '0.18em 0.35em',
          borderRadius: 4,
          fontSize: '0.875em',
          fontFamily: '"JetBrains Mono", "SFMono-Regular", Menlo, monospace',
        }}
      >
        {children}
      </code>
    );
  }

  const language = match?.[1] || 'text';

  return (
    <Box
      sx={{
        position: 'relative',
        my: 1.5,
        borderRadius: 2,
        overflow: 'hidden',
        bgcolor: '#282c34',
      }}
    >
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          px: 1.5,
          py: 0.5,
          bgcolor: 'rgba(255,255,255,0.06)',
          fontSize: 12,
          fontFamily: 'monospace',
          color: 'rgba(255,255,255,0.65)',
        }}
      >
        <span>{language}</span>
        <Tooltip title="Copy code">
          <IconButton
            size="small"
            onClick={() => navigator.clipboard.writeText(codeString)}
            sx={{ color: 'rgba(255,255,255,0.65)', '&:hover': { color: '#fff' } }}
          >
            <CopyIcon sx={{ fontSize: 14 }} />
          </IconButton>
        </Tooltip>
      </Box>
      <SyntaxHighlighter
        language={language}
        style={oneDark as any}
        customStyle={{
          margin: 0,
          padding: '12px 16px',
          fontSize: '0.875rem',
          background: 'transparent',
        }}
        wrapLongLines
      >
        {codeString}
      </SyntaxHighlighter>
    </Box>
  );
};

const MarkdownContent: React.FC<{ content: string }> = memo(({ content }) => (
  <ReactMarkdown
    remarkPlugins={[remarkGfm]}
    components={{
      code: CodeBlock as any,
      a: ({ node, ...props }) => (
        <a {...props} target="_blank" rel="noopener noreferrer" style={{ color: '#6366F1' }} />
      ),
      p: ({ node, ...props }) => (
        <p {...props} style={{ margin: '0.5em 0', lineHeight: 1.7 }} />
      ),
      ul: ({ node, ...props }) => (
        <ul {...props} style={{ paddingLeft: '1.5em', margin: '0.5em 0' }} />
      ),
      ol: ({ node, ...props }) => (
        <ol {...props} style={{ paddingLeft: '1.5em', margin: '0.5em 0' }} />
      ),
      blockquote: ({ node, ...props }) => (
        <blockquote
          {...props}
          style={{
            borderLeft: '3px solid #d1d5db',
            paddingLeft: '1em',
            color: '#6b7280',
            margin: '0.5em 0',
          }}
        />
      ),
      table: ({ node, ...props }) => (
        <Box sx={{ overflowX: 'auto', my: 1 }}>
          <table {...props} style={{ borderCollapse: 'collapse', width: '100%' }} />
        </Box>
      ),
      th: ({ node, ...props }) => (
        <th
          {...props}
          style={{
            border: '1px solid #e5e7eb',
            padding: '0.5em',
            background: '#f9fafb',
            textAlign: 'left',
          }}
        />
      ),
      td: ({ node, ...props }) => (
        <td {...props} style={{ border: '1px solid #e5e7eb', padding: '0.5em' }} />
      ),
    }}
  >
    {content}
  </ReactMarkdown>
));

const Message: React.FC<MessageProps> = ({ message, isStreaming }) => {
  const isUser = message.role === 'user';

  return (
    <Box
      sx={{
        width: '100%',
        bgcolor: isUser ? 'transparent' : 'rgba(247, 247, 248, 0.6)',
        py: 3,
      }}
    >
      <Box
        sx={{
          maxWidth: 820,
          mx: 'auto',
          px: { xs: 2, md: 3 },
          display: 'flex',
          gap: 2,
          alignItems: 'flex-start',
        }}
      >
        <Avatar
          sx={{
            width: 32,
            height: 32,
            bgcolor: isUser ? 'primary.main' : '#10a37f',
            flexShrink: 0,
          }}
        >
          {isUser ? <PersonIcon sx={{ fontSize: 18 }} /> : <BotIcon sx={{ fontSize: 18 }} />}
        </Avatar>
        <Box sx={{ flex: 1, minWidth: 0, color: 'text.primary', fontSize: '0.95rem' }}>
          {isUser ? (
            <Box sx={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word', lineHeight: 1.7 }}>
              {message.content}
            </Box>
          ) : (
            <Box
              sx={{
                '& > *:first-of-type': { mt: 0 },
                '& > *:last-child': { mb: 0 },
                wordBreak: 'break-word',
              }}
            >
              {message.content ? (
                <MarkdownContent content={message.content} />
              ) : (
                <TypingDots />
              )}
              {isStreaming && message.content && <BlinkingCursor />}
            </Box>
          )}
        </Box>
      </Box>
    </Box>
  );
};

const BlinkingCursor: React.FC = () => (
  <Box
    component="span"
    sx={{
      display: 'inline-block',
      width: 8,
      height: 16,
      bgcolor: 'text.primary',
      ml: 0.5,
      verticalAlign: 'text-bottom',
      animation: 'blink 1s step-end infinite',
      '@keyframes blink': {
        '0%, 100%': { opacity: 1 },
        '50%': { opacity: 0 },
      },
    }}
  />
);

const TypingDots: React.FC = () => (
  <Box sx={{ display: 'flex', gap: 0.5, py: 1 }}>
    {[0, 1, 2].map((i) => (
      <Box
        key={i}
        sx={{
          width: 6,
          height: 6,
          borderRadius: '50%',
          bgcolor: 'text.secondary',
          animation: 'bounce 1.4s infinite ease-in-out both',
          animationDelay: `${i * 0.16}s`,
          '@keyframes bounce': {
            '0%, 80%, 100%': { transform: 'scale(0)', opacity: 0.4 },
            '40%': { transform: 'scale(1)', opacity: 1 },
          },
        }}
      />
    ))}
  </Box>
);

export interface MessageListProps {
  messages: MessageItem[];
  streamingMessageId?: string | null;
}

export const MessageList: React.FC<MessageListProps> = ({ messages, streamingMessageId }) => (
  <Box sx={{ width: '100%' }}>
    {messages.map((msg) => (
      <Message
        key={msg.id}
        message={msg}
        isStreaming={streamingMessageId === msg.id}
      />
    ))}
  </Box>
);

export default MessageList;
