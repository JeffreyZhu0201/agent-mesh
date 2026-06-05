import React, { useState } from 'react';
import {
  Box,
  Card,
  CardContent,
  CardActions,
  Typography,
  Grid,
  Button,
  TextField,
  InputAdornment,
  Chip,
  IconButton,
} from '@mui/material';
import { Search as SearchIcon, InstallDesktop as InstallIcon, Check as CheckIcon } from '@mui/icons-material';

// Plugin data interface
interface PluginData {
  id: string;
  name: string;
  description: string;
  author: string;
  version: string;
  installed: boolean;
  category: string;
}

// Mock plugin data
const mockPlugins: PluginData[] = [
  {
    id: '1',
    name: 'Slack Integration',
    description: 'Connect your AgentMesh workspace with Slack for team notifications and messaging.',
    author: 'AgentMesh Team',
    version: '1.2.0',
    installed: false,
    category: 'Communication',
  },
  {
    id: '2',
    name: 'GitHub Assistant',
    description: 'AI-powered code review and PR management directly from your chat interface.',
    author: 'AgentMesh Team',
    version: '2.0.1',
    installed: true,
    category: 'Development',
  },
  {
    id: '3',
    name: 'Analytics Dashboard',
    description: 'Real-time analytics and insights for your team productivity and usage patterns.',
    author: 'AgentMesh Team',
    version: '1.5.0',
    installed: false,
    category: 'Analytics',
  },
  {
    id: '4',
    name: 'Calendar Sync',
    description: 'Sync your calendar events and schedule meetings seamlessly across platforms.',
    author: 'AgentMesh Team',
    version: '1.0.3',
    installed: false,
    category: 'Productivity',
  },
  {
    id: '5',
    name: 'Database Explorer',
    description: 'Visual query builder and database management for your data workflows.',
    author: 'AgentMesh Team',
    version: '0.9.2',
    installed: false,
    category: 'Development',
  },
  {
    id: '6',
    name: 'Email Assistant',
    description: 'Smart email drafting and scheduling with natural language commands.',
    author: 'AgentMesh Team',
    version: '1.1.0',
    installed: true,
    category: 'Communication',
  },
];

// Plugin card component
interface PluginCardProps {
  plugin: PluginData;
  onInstall: (id: string) => void;
  onUninstall: (id: string) => void;
}

const PluginCard: React.FC<PluginCardProps> = ({ plugin, onInstall, onUninstall }) => {
  return (
    <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <CardContent sx={{ flexGrow: 1 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 2 }}>
          <Typography variant="h6" sx={{ fontWeight: 'bold' }}>
            {plugin.name}
          </Typography>
          <Chip label={plugin.category} size="small" variant="outlined" />
        </Box>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          {plugin.description}
        </Typography>
        <Box sx={{ display: 'flex', gap: 1, alignItems: 'center' }}>
          <Typography variant="caption" color="text.secondary">
            By {plugin.author}
          </Typography>
          <Typography variant="caption" color="text.secondary">
            v{plugin.version}
          </Typography>
        </Box>
      </CardContent>
      <CardActions sx={{ p: 2, pt: 0 }}>
        {plugin.installed ? (
          <Button
            variant="outlined"
            color="success"
            startIcon={<CheckIcon />}
            onClick={() => onUninstall(plugin.id)}
            fullWidth
          >
            Installed
          </Button>
        ) : (
          <Button
            variant="contained"
            startIcon={<InstallIcon />}
            onClick={() => onInstall(plugin.id)}
            fullWidth
          >
            Install
          </Button>
        )}
      </CardActions>
    </Card>
  );
};

const PluginMarket: React.FC = () => {
  const [searchQuery, setSearchQuery] = useState('');
  const [plugins, setPlugins] = useState<PluginData[]>(mockPlugins);

  const handleInstall = (id: string) => {
    setPlugins(prev =>
      prev.map(plugin =>
        plugin.id === id ? { ...plugin, installed: true } : plugin
      )
    );
  };

  const handleUninstall = (id: string) => {
    setPlugins(prev =>
      prev.map(plugin =>
        plugin.id === id ? { ...plugin, installed: false } : plugin
      )
    );
  };

  const filteredPlugins = plugins.filter(
    plugin =>
      plugin.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      plugin.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
      plugin.category.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4" sx={{ fontWeight: 'bold' }}>
          Plugin Market
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {plugins.filter(p => p.installed).length} of {plugins.length} installed
        </Typography>
      </Box>

      {/* Search Bar */}
      <TextField
        fullWidth
        placeholder="Search plugins..."
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        InputProps={{
          startAdornment: (
            <InputAdornment position="start">
              <SearchIcon />
            </InputAdornment>
          ),
        }}
        sx={{ mb: 3 }}
      />

      {/* Plugin Cards Grid */}
      <Grid container spacing={3}>
        {filteredPlugins.map((plugin) => (
          <Grid item xs={12} sm={6} md={4} key={plugin.id}>
            <PluginCard
              plugin={plugin}
              onInstall={handleInstall}
              onUninstall={handleUninstall}
            />
          </Grid>
        ))}
      </Grid>

      {filteredPlugins.length === 0 && (
        <Box sx={{ textAlign: 'center', py: 8 }}>
          <Typography variant="h6" color="text.secondary">
            No plugins found matching "{searchQuery}"
          </Typography>
        </Box>
      )}
    </Box>
  );
};

export default PluginMarket;
