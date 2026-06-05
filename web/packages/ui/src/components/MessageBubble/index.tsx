import React from 'react';
import { Box, Typography, Chip, Avatar, Paper } from '@mui/material';
import {
  Person as PersonIcon,
  SmartToy as BotIcon,
} from '@mui/icons-material';

export interface MessageBubbleProps {
  content: string;
  role: 'user' | 'assistant';
  model?: string;
  timestamp?: Date;
  isStreaming?: boolean;
}

const formatTime = (date: Date): string => {
  return date.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
  });
};

const MessageBubble: React.FC<MessageBubbleProps> = ({
  content,
  role,
  model,
  timestamp = new Date(),
  isStreaming = false,
}) => {
  const isUser = role === 'user';

  return (
    <Box
      sx={{
        display: 'flex',
        justifyContent: isUser ? 'flex-end' : 'flex-start',
        mb: 2,
        px: 2,
      }}
    >
      <Box
        sx={{
          display: 'flex',
          flexDirection: isUser ? 'row-reverse' : 'row',
          alignItems: 'flex-start',
          maxWidth: '70%',
          gap: 1.5,
        }}
      >
        <Avatar
          sx={{
            bgcolor: isUser ? 'primary.main' : 'secondary.main',
            width: 36,
            height: 36,
          }}
        >
          {isUser ? <PersonIcon /> : <BotIcon />}
        </Avatar>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.5 }}>
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              gap: 1,
              flexDirection: isUser ? 'row-reverse' : 'row',
            }}
          >
            {!isUser && model && (
              <Chip
                label={model}
                size="small"
                sx={{
                  height: 20,
                  fontSize: '0.7rem',
                  bgcolor: 'secondary.light',
                  color: 'secondary.contrastText',
                }}
              />
            )}
            {isStreaming && (
              <Chip
                label="Streaming..."
                size="small"
                sx={{
                  height: 20,
                  fontSize: '0.7rem',
                  bgcolor: 'info.light',
                  color: 'info.contrastText',
                }}
              />
            )}
          </Box>
          <Paper
            elevation={0}
            sx={{
              p: 2,
              borderRadius: isUser
                ? '16px 16px 4px 16px'
                : '16px 16px 16px 4px',
              bgcolor: isUser ? 'primary.main' : 'grey.100',
              color: isUser ? 'primary.contrastText' : 'text.primary',
              borderBottomRightRadius: isUser ? 4 : 16,
              borderBottomLeftRadius: isUser ? 16 : 4,
            }}
          >
            <Typography
              variant="body1"
              sx={{
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word',
              }}
            >
              {content}
              {isStreaming && (
                <Box
                  component="span"
                  sx={{
                    display: 'inline-block',
                    width: 8,
                    height: 16,
                    bgcolor: isUser ? 'primary.contrastText' : 'text.primary',
                    ml: 0.5,
                    animation: 'blink 1s step-end infinite',
                    '@keyframes blink': {
                      '0%, 100%': { opacity: 1 },
                      '50%': { opacity: 0 },
                    },
                  }}
                />
              )}
            </Typography>
          </Paper>
          <Typography
            variant="caption"
            sx={{
              color: 'text.secondary',
              px: 0.5,
              alignSelf: isUser ? 'flex-end' : 'flex-start',
            }}
          >
            {formatTime(timestamp)}
          </Typography>
        </Box>
      </Box>
    </Box>
  );
};

export default MessageBubble;