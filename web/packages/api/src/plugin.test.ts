import { beforeEach, describe, expect, it, vi } from 'vitest';
import { AUTH_STORAGE_KEY } from './authUtils';
import { fetchAdminPlugins, fetchEnabledPlugins, togglePlugin } from './plugin';

describe('plugin API', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  it('fetchEnabledPlugins 返回 plugins 数组', async () => {
    localStorage.setItem(
      AUTH_STORAGE_KEY,
      JSON.stringify({ state: { token: 'jwt' } })
    );
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({
          plugins: [{ key: 'chat', name: 'Chat', version: '0.1.0', enabled: true }],
        }),
      })
    );

    const plugins = await fetchEnabledPlugins();
    expect(plugins).toHaveLength(1);
    expect(plugins[0].key).toBe('chat');
  });

  it('fetchAdminPlugins 解析管理端列表', async () => {
    localStorage.setItem(
      AUTH_STORAGE_KEY,
      JSON.stringify({ state: { token: 'jwt' } })
    );
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({
          plugins: [
            { key: 'chat', name: 'Chat', version: '0.1.0', enabled: false },
          ],
        }),
      })
    );

    const plugins = await fetchAdminPlugins();
    expect(plugins[0].enabled).toBe(false);
  });

  it('togglePlugin 发送 POST 请求', async () => {
    localStorage.setItem(
      AUTH_STORAGE_KEY,
      JSON.stringify({ state: { token: 'jwt' } })
    );
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({ message: 'ok' }),
      })
    );

    await togglePlugin('chat', true);
    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining('/plugin/admin/toggle'),
      expect.objectContaining({ method: 'POST' })
    );
  });
});
