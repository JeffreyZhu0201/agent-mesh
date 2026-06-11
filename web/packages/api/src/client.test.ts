import { beforeEach, describe, expect, it, vi } from 'vitest';
import { AUTH_STORAGE_KEY } from './authUtils';
import { apiRequest } from './client';

describe('apiRequest', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  it('自动附加 Bearer Token 并解包 user-svc 响应', async () => {
    localStorage.setItem(
      AUTH_STORAGE_KEY,
      JSON.stringify({ state: { token: 'jwt-token' } })
    );

    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({ success: true, data: [{ id: 1, name: 'Default' }] }),
      })
    );

    const tenants = await apiRequest<Array<{ id: number; name: string }>>('/admin/tenants');
    expect(tenants).toHaveLength(1);
    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining('/admin/tenants'),
      expect.objectContaining({
        headers: expect.any(Headers),
      })
    );
    const call = vi.mocked(fetch).mock.calls[0];
    const headers = call[1]?.headers as Headers;
    expect(headers.get('Authorization')).toBe('Bearer jwt-token');
  });

  it('HTTP 失败时抛出带 error 字段的消息', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 403,
        json: async () => ({ error: 'platform admin access required' }),
      })
    );

    await expect(apiRequest('/admin/tenants')).rejects.toThrow('platform admin access required');
  });
});
