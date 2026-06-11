import { beforeEach, describe, expect, it, vi } from 'vitest';
import { usePluginsStore } from './pluginsStore';

describe('pluginsStore', () => {
  beforeEach(() => {
    localStorage.clear();
    usePluginsStore.setState({
      enabledKeys: new Set(),
      plugins: [],
      loaded: false,
      loading: false,
      error: null,
    });
    vi.restoreAllMocks();
  });

  it('marks loaded with empty plugins when token is missing', async () => {
    await usePluginsStore.getState().loadEnabled();
    const state = usePluginsStore.getState();
    expect(state.loaded).toBe(true);
    expect(state.plugins).toEqual([]);
    expect(state.enabledKeys.size).toBe(0);
  });

  it('loads enabled plugin keys from API', async () => {
    localStorage.setItem(
      'auth-storage',
      JSON.stringify({ state: { token: 'jwt-token' } })
    );

    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: true,
        json: async () => ({
          plugins: [
            { key: 'chat', name: 'Chat', version: '0.1.0', enabled: true },
            { key: 'dashboard', name: 'Dashboard', version: '0.1.0', enabled: true },
          ],
        }),
      })
    );

    await usePluginsStore.getState().loadEnabled();
    const state = usePluginsStore.getState();
    expect(state.loaded).toBe(true);
    expect(state.enabledKeys).toEqual(new Set(['chat', 'dashboard']));
    expect(state.plugins).toHaveLength(2);
  });

  it('stores error when plugin API fails', async () => {
    localStorage.setItem(
      'auth-storage',
      JSON.stringify({ state: { token: 'jwt-token' } })
    );

    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        json: async () => ({}),
      })
    );

    await usePluginsStore.getState().loadEnabled();
    const state = usePluginsStore.getState();
    expect(state.loaded).toBe(true);
    expect(state.error).toBe('Request failed');
  });
});
