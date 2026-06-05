import React, { useState } from 'react';
import {
  Box,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  Typography,
  TextField,
  InputAdornment,
  Badge,
  Divider,
  Button,
  Avatar,
  IconButton,
} from '@mui/material';
import {
  Search as SearchIcon,
  Add as AddIcon,
  Chat as ChatIcon,
  Delete as DeleteIcon,
} from '@mui/icons-material';

export interface Conversation {
  id: string;
  title: string;
  preview?: string;
  updatedAt: Date;
  model?: string;
}

export interface ChatSidebarProps {
  conversations: Conversation[];
  selectedId?: string;
  onSelectConversation: (id: string) => void;
  onNewChat: () => void;
  onSearch?: (query: string) => void;
  onDeleteConversation?: (id: string) => void;
}

const ChatSidebar: React.FC<ChatSidebarProps> = ({
  conversations,
  selectedId,
  onSelectConversation,
  onNewChat,
  onSearch,
  onDeleteConversation,
}) => {
  const [searchQuery, setSearchQuery] = useState('');

  const handleSearchChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setSearchQuery(event.target.value);
    onSearch?.(event.target.value);
  };

  const formatDate = (date: Date): string => {
    const now = new Date();
    const diffDays = Math.floor(
      (now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24)
    );

    if (diffDays === 0) {
      return date.toLocaleTimeString('en-US', {
        hour: '2-digit',
        minute: '2-digit',
      });
    } else if (diffDays === 1) {
      return 'Yesterday';
    } else if (diffDays < 7) {
      return date.toLocaleDateString('en-US', { weekday: 'short' });
    } else {
      return date.toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
      });
    }
  };

  return (
    <Box
      sx={{
        width: 280,
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        bgcolor: 'background.paper',
        borderRight: 1,
        borderColor: 'divider',
      }}
    >
      {/* New Chat Button */}
      <Box sx={{ p: 2 }}>
        <Button
          fullWidth
          variant="contained"
          startIcon={<AddIcon />}
          onClick={onNewChat}
          sx={{
            justifyContent: 'flex-start',
            py: 1.5,
            borderRadius: 2,
          }}
        >
          New Chat
        </Button>
      </Box>

      <Divider />

      {/* Search */}
      <Box sx={{ px: 2, py: 1.5 }}>
        <TextField
          fullWidth
          size="small"
          placeholder="Search conversations..."
          value={searchQuery}
          onChange={handleSearchChange}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon sx={{ color: 'text.secondary', fontSize: 20 }} />
              </InputAdornment>
            ),
          }}
          sx={{
            '& .MuiOutlinedInput-root': {
              borderRadius: 2,
              bgcolor: 'background.default',
            },
          }}
        />
      </Box>

      <Divider />

      {/* Conversation List */}
      <Box sx={{ flexGrow: 1, overflow: 'auto' }}>
        <List disablePadding>
          {conversations.length === 0 ? (
            <Box
              sx={{
                p: 3,
                textAlign: 'center',
                color: 'text.secondary',
              }}
            >
              <ChatIcon sx={{ fontSize: 40, mb: 1, opacity: 0.5 }} />
              <Typography variant="body2">No conversations yet</Typography>
              <Typography variant="caption" color="text.disabled">
                Start a new chat to begin
              </Typography>
            </Box>
          ) : (
            conversations.map((conversation) => (
              <ListItem
                      key={conversation.id}
                      disablePadding
                      secondaryAction={
                        onDeleteConversation && (
                          <IconButton
                            edge="end"
                            aria-label="delete"
                            onClick={(e) => {
                              e.stopPropagation();
                              onDeleteConversation(conversation.id);
                            }}
                            sx={{
                              mr: 1,
                              opacity: 0.7,
                              '&:hover': { opacity: 1, color: 'error.main' },
                            }}
                          >
                            <DeleteIcon fontSize="small" />
                          </IconButton>
                        )
                      }
                    >
                <ListItemButton
                  selected={selectedId === conversation.id}
                  onClick={() => onSelectConversation(conversation.id)}
                  sx={{
                    py: 1.5,
                    px: 2,
                    '&.Mui-selected': {
                      bgcolor: 'primary.light',
                      color: 'primary.contrastText',
                      '& .MuiListItemText-primary': {
                        color: 'inherit',
                      },
                      '& .MuiListItemText-secondary': {
                        color: 'inherit',
                        opacity: 0.8,
                      },
                      '&:hover': {
                        bgcolor: 'primary.main',
                      },
                    },
                    '&:hover': {
                      bgcolor: 'action.hover',
                    },
                  }}
                >
                  <Avatar
                    sx={{
                      width: 36,
                      height: 36,
                      mr: 1.5,
                      bgcolor: selectedId === conversation.id
                        ? 'primary.contrastText'
                        : 'grey.200',
                      color: selectedId === conversation.id
                        ? 'primary.main'
                        : 'text.secondary',
                    }}
                  >
                    <ChatIcon sx={{ fontSize: 20 }} />
                  </Avatar>
                  <ListItemText
                    primary={
                      <Typography
                        variant="body2"
                        sx={{
                          fontWeight: 500,
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                        }}
                      >
                        {conversation.title}
                      </Typography>
                    }
                    secondary={
                      <Box
                        component="span"
                        sx={{
                          display: 'flex',
                          justifyContent: 'space-between',
                          alignItems: 'center',
                          width: '100%',
                        }}
                      >
                        <Typography
                          variant="caption"
                          sx={{
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                            maxWidth: '60%',
                          }}
                        >
                          {conversation.preview || 'No messages'}
                        </Typography>
                        <Typography
                          variant="caption"
                          sx={{ opacity: 0.7 }}
                        >
                          {formatDate(conversation.updatedAt)}
                        </Typography>
                      </Box>
                    }
                  />
                </ListItemButton>
              </ListItem>
            ))
          )}
        </List>
      </Box>
    </Box>
  );
};

export default ChatSidebar;