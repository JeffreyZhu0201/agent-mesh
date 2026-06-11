import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useAuthStore } from './authStore';

describe('authStore', () => {
  beforeEach(() => {
    localStorage.clear();
    useAuthStore.setState({
      user: null,
      token: null,
      isAuthenticated: false,
    });
    vi.restoreAllMocks();
  });

  it('logs in with wrapped API response', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({
          success: true,
          data: {
            token: 'jwt-token',
            user: {
              userId: 7,
              username: 'alice',
              tenantId: 1,
              role: 'viewer',
            },
          },
        }),
      })
    );

    await useAuthStore.getState().login('alice', 'secret1');

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(true);
    expect(state.token).toBe('jwt-token');
    expect(state.user).toMatchObject({
      id: 7,
      username: 'alice',
      tenantId: 1,
      role: 'viewer',
    });
  });

  it('throws backend error message on failed login', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        json: async () => ({ success: false, error: 'invalid username or password' }),
      })
    );

    await expect(useAuthStore.getState().login('alice', 'bad')).rejects.toThrow(
      'invalid username or password'
    );
  });

  it('clears auth state on logout', () => {
    useAuthStore.setState({
      user: { id: 1, username: 'alice', email: '', tenantId: 1, role: 'viewer' },
      token: 'token',
      isAuthenticated: true,
    });

    useAuthStore.getState().logout();
    const state = useAuthStore.getState();
    expect(state.user).toBeNull();
    expect(state.token).toBeNull();
    expect(state.isAuthenticated).toBe(false);
  });

  it('setToken 在有 user 时保持已认证', () => {
    useAuthStore.setState({
      user: { id: 1, username: 'alice', email: '', tenantId: 1, role: 'viewer' },
      token: null,
      isAuthenticated: false,
    });
    useAuthStore.getState().setToken('new-token');
    expect(useAuthStore.getState().token).toBe('new-token');
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
  });

  it('setUser 写入用户并标记已认证', () => {
    const user = { id: 2, username: 'bob', email: 'b@x.com', tenantId: 1, role: 'admin' };
    useAuthStore.getState().setUser(user);
    const state = useAuthStore.getState();
    expect(state.user).toEqual(user);
    expect(state.isAuthenticated).toBe(true);
  });
});
