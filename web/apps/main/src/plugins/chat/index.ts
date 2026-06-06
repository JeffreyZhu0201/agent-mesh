import { registerPlugin, Plugin } from '@agentmesh/plugin-sdk';
import ChatPage from './ChatPage';

export const chatPlugin: Plugin = {
  metadata: {
    name: 'chat',
    version: '0.1.0',
    description: 'Conversational AI chat with multi-model support',
    type: 'page',
    author: 'AgentMesh',
  },
  menuItems: [
    {
      key: 'chat',
      label: 'Chat',
      icon: 'C',
      path: '/app/chat',
      order: 10,
    },
  ],
  routes: [
    {
      path: '/app/chat',
      component: ChatPage,
      exact: true,
    },
  ],
  component: ChatPage,
};

// Self-register when this module is imported
registerPlugin(chatPlugin);

export default chatPlugin;
export { ChatPage };
