import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { authApi } from '@agentmesh/api';

// ==================== 认证状态 Store ====================
// 使用 zustand + persist 将 user/token 持久化到 localStorage（键名 auth-storage）
// Actions: login / logout / setToken / setUser
// ===========================================================

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

      /** 调用 authApi 登录，成功后写入全局认证状态 */
      login: async (username: string, password: string) => {
        const { token, user: apiUser } = await authApi.login({ username, password });
        set({
          token,
          user: {
            id: Number(apiUser.id) || 0,
            username: apiUser.username,
            email: apiUser.email,
            tenantId: apiUser.tenantId,
            role: apiUser.role,
          },
          isAuthenticated: true,
        });
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
