import React from 'react';
import { BrowserRouter, Routes, Route, useNavigate } from 'react-router-dom';
import { Layout } from '@agentmesh/ui';
import { useMenuItems, getRouteConfigs } from '@agentmesh/plugin-sdk';
import Dashboard from './pages/Dashboard';
import PluginMarket from './pages/PluginMarket';

// Chat placeholder component
const Chat: React.FC = () => {
  return (
    <div style={{ padding: 20 }}>
      <h2>Chat</h2>
      <p>Chat functionality coming soon...</p>
    </div>
  );
};

// Mock user for demo
const MOCK_USER = {
  name: 'Demo User',
  email: 'demo@agentmesh.io',
  avatar: undefined,
};

// Main Layout wrapper with navigation
const MainLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const navigate = useNavigate();
  const menuItems = useMenuItems();
  const routeConfigs = getRouteConfigs();

  // Map plugin menu items to nav items
  const navItems = React.useMemo(() => {
    const baseNav = [
      { label: 'Dashboard', icon: <span style={{ fontWeight: 'bold' }}>D</span>, path: '/app' },
      { label: 'Plugin Market', icon: <span style={{ fontWeight: 'bold' }}>P</span>, path: '/app/plugins' },
    ];

    const pluginNav = menuItems.map(item => ({
      label: item.label,
      icon: <span style={{ fontWeight: 'bold' }}>{item.label.charAt(0)}</span>,
      path: item.path,
    }));

    return [...baseNav, ...pluginNav];
  }, [menuItems]);

  const handleNavigate = (path: string) => {
    navigate(path);
  };

  const handleLogout = () => {
    console.log('Logout clicked');
  };

  return (
    <Layout
      navItems={navItems}
      user={MOCK_USER}
      onNavigate={handleNavigate}
      onLogout={handleLogout}
    >
      {children}
    </Layout>
  );
};

// Protected route wrapper
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  // In a real app, check authentication here
  return <>{children}</>;
};

// App routes component
const AppRoutes: React.FC = () => {
  const menuItems = useMenuItems();
  const routeConfigs = getRouteConfigs();

  return (
    <Routes>
      <Route path="/app" element={<ProtectedRoute><MainLayout><Dashboard /></MainLayout></ProtectedRoute>} />
      <Route path="/app/chat" element={<ProtectedRoute><MainLayout><Chat /></MainLayout></ProtectedRoute>} />
      <Route path="/app/plugins" element={<ProtectedRoute><MainLayout><PluginMarket /></MainLayout></ProtectedRoute>} />
      {/* Dynamic plugin routes */}
      {routeConfigs.map((route, index) => (
        <Route
          key={`plugin-route-${index}`}
          path={route.path}
          element={<ProtectedRoute><MainLayout>{route.component ? <route.component /> : <div>Plugin Content</div>}</MainLayout></ProtectedRoute>}
        />
      ))}
    </Routes>
  );
};

// Root App component
export default function App() {
  return (
    <BrowserRouter>
      <AppRoutes />
    </BrowserRouter>
  );
}
