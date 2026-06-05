import React from 'react';
import { BrowserRouter, Routes, Route, useNavigate, Navigate } from 'react-router-dom';
import { Layout } from '@agentmesh/ui';
import { useMenuItems, getRouteConfigs } from '@agentmesh/plugin-sdk';
import { useAuthStore } from '@agentmesh/stores';
import Dashboard from './pages/Dashboard';
import PluginMarket from './pages/PluginMarket';
import Chat from './pages/Chat';
import Login from './pages/Login';
import Register from './pages/Register';
import { ProtectedRoute } from './components/ProtectedRoute';

// Main Layout wrapper with navigation
const MainLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const navigate = useNavigate();
  const menuItems = useMenuItems();
  const routeConfigs = getRouteConfigs();
  const logout = useAuthStore((state) => state.logout);
  const user = useAuthStore((state) => state.user);

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
    logout();
    navigate('/app/login');
  };

  return (
    <Layout
      navItems={navItems}
      user={user ? { name: user.username, email: user.email, avatar: user.avatar } : undefined}
      onNavigate={handleNavigate}
      onLogout={handleLogout}
    >
      {children}
    </Layout>
  );
};

// Auth route wrapper (redirects to /app if already authenticated)
const AuthRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  if (isAuthenticated) {
    return <Navigate to="/app" replace />;
  }

  return <>{children}</>;
};

// App routes component
const AppRoutes: React.FC = () => {
  const menuItems = useMenuItems();
  const routeConfigs = getRouteConfigs();

  return (
    <Routes>
      {/* Auth routes */}
      <Route path="/app/login" element={<AuthRoute><Login /></AuthRoute>} />
      <Route path="/app/register" element={<AuthRoute><Register /></AuthRoute>} />

      {/* Protected routes */}
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
      {/* Catch all - redirect to app */}
      <Route path="*" element={<Navigate to="/app" replace />} />
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