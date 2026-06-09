import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';

// ==================== Zustand Store 状态管理说明 ====================
// Zustand 是 React 的轻量级状态管理库
// 这个 authStore 管理用户的认证状态
//
// 状态结构 (State):
// - user: User | null - 当前登录用户信息
// - token: string | null - JWT 认证令牌
// - isAuthenticated: boolean - 是否已认证 (由 user 和 token 计算得出)
//
// 操作方法 (Actions):
// - login(username, password): 登录，调用后端 API 获取 token 和 user
// - logout(): 退出登录，清除所有状态
// - setToken(token): 设置 token
// - setUser(user): 设置用户信息
//
// 持久化机制:
// - 使用 zustand/middleware 的 persist 中间件
// - 存储键名: 'auth-storage' (localStorage)
// - partialize: 只持久化 user, token, isAuthenticated (不持久化 loading 等状态)
// - 刷新页面后自动从 localStorage 恢复登录状态
//
// isAuthenticated 计算逻辑:
// - 每次 setState 时动态计算
// - setToken: 如果已有 user，则 isAuthenticated = true
// - setUser: 直接设置 isAuthenticated = true
// - logout: 直接设置 isAuthenticated = false
// ==================== Zustand Store 状态管理说明 END ====================

export interface User {
  id: number;
  username: string;
  email: string;
  tenantId: number;
  role: string;
}

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  setToken: (token: string) => void;
  setUser: (user: User) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,

      login: async (username: string, password: string) => {
        const response = await fetch('http://localhost:8080/api/user/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username, password }),
        });

        if (!response.ok) {
          const error = await response.json();
          throw new Error(error.message || 'Login failed');
        }

        const data = await response.json();
        set({ user: data.user, token: data.token, isAuthenticated: true });
      },

      logout: () => {
        set({ user: null, token: null, isAuthenticated: false });
      },

      setToken: (token: string) => {
        set((state) => ({ token, isAuthenticated: !!state.user }));
      },

      setUser: (user: User) => {
        set({ user, isAuthenticated: true });
      },
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
