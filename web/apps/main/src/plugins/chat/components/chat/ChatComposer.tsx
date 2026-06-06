import React, { useState, useCallback, useRef, useEffect } from 'react';
import { Box, IconButton, Paper, Tooltip } from '@mui/material';
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';

interface ChatComposerProps {
  onSend: (text: string) => void;
  disabled?: boolean;
  placeholder?: string;
}

export const ChatComposer: React.FC<ChatComposerProps> = ({
  onSend,
  disabled = false,
  placeholder = 'Message AgentMesh...',
}) => {
  const [value, setValue] = useState('');
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Auto-resize textarea
  useEffect(() => {
    const ta = textareaRef.current;
    if (!ta) return;
    ta.style.height = 'auto';
    ta.style.height = Math.min(ta.scrollHeight, 200) + 'px';
  }, [value]);

  const handleSend = useCallback(() => {
    const trimmed = value.trim();
    if (!trimmed || disabled) return;
    onSend(trimmed);
    setValue('');
  }, [value, disabled, onSend]);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const canSend = value.trim().length > 0 && !disabled;

  return (
    <Box sx={{ width: '100%', px: { xs: 2, md: 3 }, pb: 2, pt: 1 }}>
      <Box sx={{ maxWidth: 820, mx: 'auto' }}>
        <Paper
          elevation={0}
          sx={{
            display: 'flex',
            alignItems: 'flex-end',
            gap: 1,
            p: 1,
            border: '1px solid',
            borderColor: 'divider',
            borderRadius: 3,
            bgcolor: 'background.paper',
            transition: 'border-color 0.15s',
            '&:focus-within': {
              borderColor: 'primary.main',
              boxShadow: '0 0 0 1px rgba(99,102,241,0.25)',
            },
          }}
        >
          <textarea
            ref={textareaRef}
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            disabled={disabled}
            rows={1}
            style={{
              flex: 1,
              border: 'none',
              outline: 'none',
              resize: 'none',
              background: 'transparent',
              fontFamily: 'inherit',
              fontSize: '0.95rem',
              lineHeight: 1.5,
              padding: '8px 10px',
              maxHeight: 200,
              color: 'inherit',
            }}
          />
          <Tooltip title={canSend ? 'Send (Enter)' : 'Type something...'}>
            <span>
              <IconButton
                onClick={handleSend}
                disabled={!canSend}
                sx={{
                  width: 36,
                  height: 36,
                  bgcolor: canSend ? 'primary.main' : 'action.disabledBackground',
                  color: canSend ? 'primary.contrastText' : 'action.disabled',
                  '&:hover': { bgcolor: canSend ? 'primary.dark' : 'action.disabledBackground' },
                  borderRadius: 2,
                  transition: 'background-color 0.15s',
                }}
              >
                <ArrowUpwardIcon sx={{ fontSize: 18 }} />
              </IconButton>
            </span>
          </Tooltip>
        </Paper>
        <Box
          sx={{
            textAlign: 'center',
            mt: 1,
            fontSize: 11,
            color: 'text.secondary',
          }}
        >
          AgentMesh can make mistakes. Verify important info.
        </Box>
      </Box>
    </Box>
  );
};

export default ChatComposer;