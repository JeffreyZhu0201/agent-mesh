import React from 'react';
import { Box, Typography, Paper } from '@mui/material';
import SmartToyOutlinedIcon from '@mui/icons-material/SmartToyOutlined';
import CodeIcon from '@mui/icons-material/Code';
import AutoFixHighIcon from '@mui/icons-material/AutoFixHigh';
import QuestionAnswerIcon from '@mui/icons-material/QuestionAnswer';

const promptSuggestions = [
  {
    icon: <CodeIcon />,
    title: 'Write code',
    description: 'Generate a React component, API endpoint, or script',
  },
  {
    icon: <AutoFixHighIcon />,
    title: 'Explain & debug',
    description: 'Help me understand or fix a piece of code',
  },
  {
    icon: <QuestionAnswerIcon />,
    title: 'Answer questions',
    description: 'Ask me anything about the project or architecture',
  },
];

interface WelcomeScreenProps {
  onSendPrompt: (text: string) => void;
}

export const WelcomeScreen: React.FC<WelcomeScreenProps> = ({ onSendPrompt }) => (
  <Box
    sx={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      height: '100%',
      px: 2,
    }}
  >
    {/* Logo / greeting */}
    <Box
      sx={{
        mb: 2,
        width: 48,
        height: 48,
        borderRadius: 2,
        bgcolor: 'primary.main',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        color: '#fff',
      }}
    >
      <SmartToyOutlinedIcon sx={{ fontSize: 28 }} />
    </Box>
    <Typography variant="h5" sx={{ fontWeight: 600, mb: 0.5 }}>
      What can I help you with?
    </Typography>
    <Typography variant="body2" color="text.secondary" sx={{ mb: 4 }}>
      Powered by Doubao Seed Code
    </Typography>

    {/* Suggestion cards */}
    <Box
      sx={{
        display: 'flex',
        gap: 2,
        flexWrap: 'wrap',
        justifyContent: 'center',
        maxWidth: 720,
      }}
    >
      {promptSuggestions.map((card) => (
        <Paper
          key={card.title}
          elevation={0}
          onClick={() => onSendPrompt(card.description)}
          sx={{
            p: 2.5,
            width: 200,
            cursor: 'pointer',
            border: '1px solid',
            borderColor: 'divider',
            borderRadius: 3,
            transition: 'all 0.15s ease',
            '&:hover': {
              borderColor: 'primary.main',
              bgcolor: 'rgba(99, 102, 241, 0.04)',
              transform: 'translateY(-1px)',
            },
          }}
        >
          <Box sx={{ color: 'primary.main', mb: 1 }}>{card.icon}</Box>
          <Typography variant="subtitle2" sx={{ fontWeight: 600, mb: 0.5 }}>
            {card.title}
          </Typography>
          <Typography variant="caption" color="text.secondary">
            {card.description}
          </Typography>
        </Paper>
      ))}
    </Box>
  </Box>
);

export default WelcomeScreen;