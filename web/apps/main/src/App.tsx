import React, { useEffect } from 'react';
import { BrowserRouter, Routes, Route, useNavigate, Navigate } from 'react-router-dom';
import { Box, CircularProgress } from '@mui/material';
import { Layout } from '@agentmesh/ui';
import { useMenuItems, getRouteConfigs } from '@agentmesh/plugin-sdk';
import { useAuthStore, usePluginsStore } from '@agentmesh/stores';
import Login from './pages/Login';
import Register from './pages/Register';
import { ProtectedRoute } from './components/ProtectedRoute';

// Import built-in plugins — each self-registers when imported.
import './plugins';

// Main Layout wrapper with navigation
const MainLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const navigate = useNavigate();
  const enabledKeys = usePluginsStore((state) => state.enabledKeys);
  const menuItems = useMenuItems(enabledKeys);
  const logout = useAuthStore((state) => state.logout);
  const user = useAuthStore((state) => state.user);

  const navItems = React.useMemo(() => {
    return menuItems.map((item) => ({
      label: item.label,
      icon: <span style={{ fontWeight: 'bold' }}>{item.label.charAt(0)}</span>,
      path: item.path,
    }));
  }, [menuItems]);

  const handleNavigate = (path: string) => navigate(path);
  const handleLogout = () => {
    logout();
    usePluginsStore.getState().reset();
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

// Loads the user's enabled plugins after auth. Renders a spinner until the
// first load completes so we never render plugin routes before knowing what
// the tenant has enabled.
const PluginsLoader: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const loadEnabled = usePluginsStore((state) => state.loadEnabled);
  const loaded = usePluginsStore((state) => state.loaded);

  useEffect(() => {
    if (isAuthenticated && !loaded) {
      loadEnabled();
    }
  }, [isAuthenticated, loaded, loadEnabled]);

  if (!loaded) {
    return (
      <Box sx={{ display: 'flex', height: '100vh', alignItems: 'center', justifyContent: 'center' }}>
        <CircularProgress />
      </Box>
    );
  }

  return <>{children}</>;
};

// Auth route wrapper (redirects to /app if already authenticated)
const AuthRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  if (isAuthenticated) return <Navigate to="/app" replace />;
  return <>{children}</>;
};

// Pick the first enabled plugin's route as the landing target.
// If nothing enabled, show a friendly placeholder.
const LandingRedirect: React.FC = () => {
  const enabledKeys = usePluginsStore((state) => state.enabledKeys);
  const menuItems = useMenuItems(enabledKeys);

  if (menuItems.length === 0) {
    return (
      <Box sx={{ p: 4, textAlign: 'center' }}>
        <h3>No plugins enabled</h3>
        <p>Ask your administrator to enable a plugin to get started.</p>
      </Box>
    );
  }

  return <Navigate to={menuItems[0].path} replace />;
};

// App routes — filters plugin routes by what's enabled for the tenant.
const AppRoutes: React.FC = () => {
  const enabledKeys = usePluginsStore((state) => state.enabledKeys);
  const routeConfigs = getRouteConfigs(enabledKeys);

  return (
    <Routes>
      {/* Auth routes */}
      <Route path="/app/login" element={<AuthRoute><Login /></AuthRoute>} />
      <Route path="/app/register" element={<AuthRoute><Register /></AuthRoute>} />

      {/* Default landing: pick the first enabled plugin */}
      <Route
        path="/app"
        element={
          <ProtectedRoute>
            <MainLayout>
              <LandingRedirect />
            </MainLayout>
          </ProtectedRoute>
        }
      />

      {/* Dynamic plugin routes (only enabled ones) */}
      {routeConfigs.map((route, index) => {
        const PluginComponent = route.component || (() => <div>Plugin Content</div>);
        return (
          <Route
            key={`plugin-route-${index}`}
            path={route.path}
            element={
              <ProtectedRoute>
                <MainLayout>
                  <PluginComponent />
                </MainLayout>
              </ProtectedRoute>
            }
          />
        );
      })}

      {/* Catch all - redirect to app */}
      <Route path="*" element={<Navigate to="/app" replace />} />
    </Routes>
  );
};

// Root App component
export default function App() {
  return (
    <BrowserRouter>
      <PluginsLoader>
        <AppRoutes />
      </PluginsLoader>
    </BrowserRouter>
  );
}
