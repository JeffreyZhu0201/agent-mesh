import { create } from 'zustand';

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
