import React, { useEffect } from 'react';
import { BrowserRouter, Routes, Route, useNavigate, Navigate } from 'react-router-dom';
import { Box, CircularProgress } from '@mui/material';
import { Layout } from '@agentmesh/ui';
import { useMenuItems, getRouteConfigs } from '@agentmesh/plugin-sdk';
import { useAuthStore, usePluginsStore } from '@agentmesh/stores';
import Login from './pages/Login';
import Register from './pages/Register';
import { ProtectedRoute } from './components/ProtectedRoute';

// 导入内置插件 — 每次导入都会触发插件的自我注册
import './plugins';

// ==================== 组件结构说明 ====================
// 整体应用结构: BrowserRouter -> PluginsLoader -> AppRoutes
// 1. BrowserRouter: React Router 的根路由器，提供路由上下文
// 2. PluginsLoader: 插件加载器，确保在渲染插件路由前先加载插件列表
// 3. AppRoutes: 应用路由配置，包含认证路由和动态插件路由
// ==================== 组件结构说明 END ====================

// ==================== MainLayout 说明 ====================
// MainLayout 是带有导航功能的主布局组件
// 功能包括:
// - 显示左侧/顶部导航菜单 (来自已启用插件的 menuItems)
// - 显示当前用户信息 (用户名、邮箱、头像)
// - 处理导航点击事件
// - 处理退出登录
// ==================== MainLayout 说明 END ====================
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

// ==================== PluginsLoader 说明 ====================
// PluginsLoader 插件加载器组件
// 作用: 等待插件列表加载完成后再渲染子组件
// 原因: 需要先知道当前租户启用了哪些插件，才能正确渲染对应的路由
// 工作流程:
// 1. 检查用户是否已认证 (isAuthenticated)
// 2. 如果已认证但插件列表尚未加载 (loaded === false)，调用 loadEnabled()
// 3. 在插件列表加载完成前显示加载动画 (CircularProgress)
// 4. 加载完成后才渲染 children (即 AppRoutes)
//
// 注意: 这个组件确保我们永远不会在不知道启用哪些插件的情况下渲染插件路由
// ==================== PluginsLoader 说明 END ====================
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

// ==================== LandingRedirect 说明 ====================
// LandingRedirect 落地页重定向组件
// 作用: 当用户访问 /app 时，自动跳转到第一个已启用插件的页面
// 工作流程:
// 1. 从 pluginsStore 获取当前租户已启用的插件 keys (enabledKeys)
// 2. 通过 useMenuItems(enabledKeys) 获取所有已启用插件的菜单项
// 3. 如果有已启用插件，自动重定向到第一个插件的路径 (menuItems[0].path)
// 4. 如果没有任何已启用插件，显示友好提示信息
//
// 这提供了一种智能落地页功能：用户访问首页自动进入其有权限使用的第一个插件
// ==================== LandingRedirect 说明 END ====================
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

// ==================== AppRoutes 说明 ====================
// AppRoutes 应用路由配置组件
// 作用: 定义所有应用路由，并根据 enabledKeys 过滤插件路由
//
// 路由结构:
// 1. /app/login - 登录页 (AuthRoute 包装：已登录用户会重定向到 /app)
// 2. /app/register - 注册页 (AuthRoute 包装：已登录用户会重定向到 /app)
// 3. /app - 主应用入口 (ProtectedRoute 包装，包含 MainLayout 和 LandingRedirect)
// 4. /app/plugin/:path - 动态插件路由 (由各插件提供的 routeConfigs 生成)
// 5. * - 兜底路由，重定向到 /app
//
// 如何通过 plugin-sdk 驱动路由:
// - getRouteConfigs(enabledKeys) 从 plugin-sdk 获取所有已注册插件的路由配置
// - 仅返回 enabledKeys 中包含的插件路由 (实现了多租户插件隔离)
// - 每个路由使用 MainLayout 包装，确保一致的导航体验
// ==================== AppRoutes 说明 END ====================
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
