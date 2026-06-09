import { create } from 'zustand';

// ==================== 插件状态管理说明 ====================
// 这个 store 管理插件的启用状态和列表
//
// 核心概念:
// 1. enabledKeys: Set<string> - 已启用插件的 key 集合
//    - 用于在 plugin-sdk 中过滤哪些插件应该显示
//    - 传递给 getRouteConfigs(enabledKeys) 和 useMenuItems(enabledKeys)
//    - 实现多租户插件隔离
//
// 2. plugins: PluginInfo[] - 从后端获取的插件信息数组
//    - 包含 key, name, version, enabled 等信息
//
// 3. loaded: boolean - 是否已完成首次加载
//    - 在 PluginsLoader 中等待此状态变为 true
//
// 调用流程:
// 1. App.tsx 的 PluginsLoader 组件在用户登录后调用 loadEnabled()
// 2. loadEnabled() 从后端获取当前租户可用的插件列表
// 3. 将 plugins 转换为 enabledKeys (Set) 并存储
// 4. UI 组件通过 enabledKeys 过滤显示哪些插件
// ==================== 插件状态管理说明 END ====================

interface PluginInfo {
  key: string;
  name: string;
  version: string;
  type?: string;
  description?: string;
  author?: string;
  enabled: boolean;
}

interface PluginsState {
  enabledKeys: Set<string>;
  plugins: PluginInfo[];
  loaded: boolean;
  loading: boolean;
  error: string | null;
  loadEnabled: () => Promise<void>;
  reset: () => void;
}

// ==================== getAuthToken() 说明 ====================
// 从 authStore 的 localStorage 中读取 token
// 因为 authStore 使用 'auth-storage' 作为存储键名
// 我们需要读取同一个键来获取认证 token
// ==================== getAuthToken() 说明 END ====================
// Read auth token from the same localStorage key the auth store persists to.
function getAuthToken(): string | null {
  try {
    const stored = localStorage.getItem('auth-storage');
    if (!stored) return null;
    const parsed = JSON.parse(stored);
    return parsed?.state?.token || null;
  } catch {
    return null;
  }
}

const API_BASE = 'http://localhost:8080/api';

// ==================== loadEnabled() 说明 ====================
// loadEnabled(): Promise<void>
// 作用: 从后端 API 获取当前租户已启用的插件列表
// 调用时机: App.tsx 的 PluginsLoader 组件在用户登录后调用
//
// 工作流程:
// 1. 先从 localStorage 获取 auth token
// 2. 如果没有 token，设置 loaded=true 并清空插件列表
// 3. 调用后端 API /api/plugin/list (需要 Bearer token)
// 4. 将返回的 plugins 数组转换为 enabledKeys (Set<string>)
// 5. 存储到 state 中，供 UI 过滤使用
// ==================== loadEnabled() 说明 END ====================
export const usePluginsStore = create<PluginsState>((set) => ({
  enabledKeys: new Set<string>(),
  plugins: [],
  loaded: false,
  loading: false,
  error: null,

  loadEnabled: async () => {
    const token = getAuthToken();
    if (!token) {
      set({ loaded: true, enabledKeys: new Set(), plugins: [] });
      return;
    }

    set({ loading: true, error: null });
    try {
      const response = await fetch(`${API_BASE}/plugin/list`, {
        headers: { Authorization: `Bearer ${token}` },
      });

      if (!response.ok) {
        throw new Error(`Failed to load plugins: ${response.status}`);
      }

      const data = await response.json();
      const plugins: PluginInfo[] = data.plugins || [];
      const enabledKeys = new Set(plugins.map((p) => p.key));

      set({ enabledKeys, plugins, loaded: true, loading: false });
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
