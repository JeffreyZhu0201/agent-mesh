import { beforeEach, describe, expect, it, vi } from 'vitest';
import { authApi } from './auth';

describe('authApi', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('logs in via user service endpoint', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({
          success: true,
          data: {
            token: 'jwt',
            user: { userId: 1, username: 'alice', email: 'a@example.com' },
          },
        }),
      })
    );

    const result = await authApi.login({ username: 'alice', password: 'secret1' });
    expect(result.token).toBe('jwt');
    expect(result.user.username).toBe('alice');
  });

  it('registers via user service endpoint', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({
          success: true,
          data: { id: 2, username: 'bob', email: 'b@example.com' },
        }),
      })
    );

    const result = await authApi.register({
      username: 'bob',
      password: 'secret1',
      email: 'b@example.com',
    });
    expect(result.user.username).toBe('bob');
  });

  it('throws AuthApiError on failure', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        json: async () => ({ error: 'invalid username or password' }),
      })
    );

    await expect(authApi.login({ username: 'x', password: 'y' })).rejects.toThrow(
      'invalid username or password'
    );
  });
});
