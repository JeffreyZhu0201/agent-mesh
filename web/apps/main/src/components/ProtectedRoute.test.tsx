import { describe, expect, it } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { ProtectedRoute } from './ProtectedRoute';
import { useAuthStore } from '@agentmesh/stores';

/** 渲染带路由守卫的测试树 */
function renderGuard(isAuthenticated: boolean) {
  useAuthStore.setState({ isAuthenticated, user: null, token: null });
  return render(
    <MemoryRouter initialEntries={['/app/dashboard']}>
      <Routes>
        <Route
          path="/app/dashboard"
          element={
            <ProtectedRoute>
              <div>受保护内容</div>
            </ProtectedRoute>
          }
        />
        <Route path="/app/login" element={<div>登录页</div>} />
      </Routes>
    </MemoryRouter>
  );
}

describe('ProtectedRoute', () => {
  it('未登录时重定向到登录页', () => {
    renderGuard(false);
    expect(screen.getByText('登录页')).toBeInTheDocument();
  });

  it('已登录时渲染子组件', () => {
    renderGuard(true);
    expect(screen.getByText('受保护内容')).toBeInTheDocument();
  });
});
