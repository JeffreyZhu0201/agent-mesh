import { create } from 'zustand';
import { fetchEnabledPlugins, getAuthToken, type PluginInfo } from '@agentmesh/api';

// ==================== 插件状态 Store ====================
// enabledKeys：当前租户已启用插件的 key 集合，供 plugin-sdk 过滤路由/菜单
// loadEnabled()：登录后从 plugin-svc 拉取列表
// ===========================================================

interface PluginsState {
  enabledKeys: Set<string>;
  plugins: PluginInfo[];
  loaded: boolean;
  loading: boolean;
  error: string | null;
  loadEnabled: () => Promise<void>;
  reset: () => void;
}

export const usePluginsStore = create<PluginsState>((set) => ({
  enabledKeys: new Set<string>(),
  plugins: [],
  loaded: false,
  loading: false,
  error: null,

  /** 从后端加载当前租户已启用的插件；无 Token 时清空并标记 loaded */
  loadEnabled: async () => {
    if (!getAuthToken()) {
      set({ loaded: true, enabledKeys: new Set(), plugins: [] });
      return;
    }

    set({ loading: true, error: null });
    try {
      const plugins = await fetchEnabledPlugins();
      set({
        enabledKeys: new Set(plugins.map((p) => p.key)),
        plugins,
        loaded: true,
        loading: false,
      });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load plugins',
        loaded: true,
        loading: false,
      });
    }
  },

  reset: () => set({ enabledKeys: new Set(), plugins: [], loaded: false, error: null }),
}));
