import React from 'react';
import { Box, Card, CardContent, Typography, Grid, List, ListItem, ListItemText, Avatar, Chip } from '@mui/material';
import { usePlugins } from '@agentmesh/plugin-sdk';

// Stat card component
interface StatCardProps {
  title: string;
  value: string | number;
  icon: React.ReactNode;
  color?: string;
}

const StatCard: React.FC<StatCardProps> = ({ title, value, icon, color = '#1976d2' }) => {
  return (
    <Card sx={{ height: '100%' }}>
      <CardContent>
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
          <Typography variant="h6" color="text.secondary" sx={{ fontSize: '0.875rem' }}>
            {title}
          </Typography>
          <Box
            sx={{
              width: 40,
              height: 40,
              borderRadius: 2,
              bgcolor: `${color}15`,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: color,
            }}
          >
            {icon}
          </Box>
        </Box>
        <Typography variant="h3" sx={{ fontWeight: 'bold' }}>
          {value}
        </Typography>
      </CardContent>
    </Card>
  );
};

// Recent conversation item
interface ConversationItem {
  id: string;
  title: string;
  lastMessage: string;
  timestamp: string;
  avatar?: string;
}

const mockConversations: ConversationItem[] = [
  {
    id: '1',
    title: 'Code Review Discussion',
    lastMessage: 'Can you review the authentication PR?',
    timestamp: '2 hours ago',
  },
  {
    id: '2',
    title: 'Plugin Development',
    lastMessage: 'The new plugin SDK is working great!',
    timestamp: '5 hours ago',
  },
  {
    id: '3',
    title: 'Documentation Update',
    lastMessage: 'I updated the API docs with new endpoints',
    timestamp: 'Yesterday',
  },
  {
    id: '4',
    title: 'Bug Fix Discussion',
    lastMessage: 'Fixed the memory leak in the plugin loader',
    timestamp: '2 days ago',
  },
];

const Dashboard: React.FC = () => {
  const plugins = usePlugins();
  const activePlugins = plugins.length;

  const stats = [
    {
      title: 'Total Chats',
      value: 247,
      icon: <span>💬</span>,
      color: '#1976d2',
    },
    {
      title: 'Active Plugins',
      value: activePlugins,
      icon: <span>🔌</span>,
      color: '#2e7d32',
    },
    {
      title: 'Messages',
      value: '12.4K',
      icon: <span>✉️</span>,
      color: '#ed6c02',
    },
    {
      title: 'Team Members',
      value: 8,
      icon: <span>👥</span>,
      color: '#9c27b0',
    },
  ];

  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 3, fontWeight: 'bold' }}>
        Dashboard
      </Typography>

      {/* Stats Cards */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        {stats.map((stat, index) => (
          <Grid item xs={12} sm={6} md={3} key={index}>
            <StatCard {...stat} />
          </Grid>
        ))}
      </Grid>

      {/* Recent Conversations */}
      <Card>
        <CardContent>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
            <Typography variant="h6" sx={{ fontWeight: 'bold' }}>
              Recent Conversations
            </Typography>
            <Chip label="View All" size="small" clickable />
          </Box>
          <List disablePadding>
            {mockConversations.map((conversation, index) => (
              <ListItem
                key={conversation.id}
                divider={index < mockConversations.length - 1}
                sx={{ px: 0 }}
              >
                <Avatar sx={{ mr: 2, bgcolor: 'primary.main' }}>
                  {conversation.title.charAt(0)}
                </Avatar>
                <ListItemText
                  primary={
                    <Typography variant="subtitle1" sx={{ fontWeight: 500 }}>
                      {conversation.title}
                    </Typography>
                  }
                  secondary={
                    <Box component="span">
                      <Typography variant="body2" color="text.secondary" component="span">
                        {conversation.lastMessage}
                      </Typography>
                      <Typography
                        variant="caption"
                        color="text.secondary"
                        component="span"
                        sx={{ ml: 2 }}
                      >
                        {conversation.timestamp}
                      </Typography>
                    </Box>
                  }
                />
              </ListItem>
            ))}
          </List>
        </CardContent>
      </Card>
    </Box>
  );
};

export default Dashboard;
