import { useEffect, useState } from 'react';

interface Plugin {
  key: string;
  name: string;
  version: string;
  type?: string;
  description?: string;
  author?: string;
  enabled: boolean;
}

const API_BASE = 'http://localhost:8080/api';

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

export function PluginManagement() {
  const [plugins, setPlugins] = useState<Plugin[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [togglingKey, setTogglingKey] = useState<string | null>(null);

  const loadPlugins = async () => {
    const token = getAuthToken();
    if (!token) {
      setError('Not authenticated. Please log in via the main app first.');
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`${API_BASE}/plugin/admin/list`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      setPlugins(data.plugins || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadPlugins();
  }, []);

  const toggleStatus = async (pluginKey: string, currentEnabled: boolean) => {
    const token = getAuthToken();
    if (!token) return;
    setTogglingKey(pluginKey);
    // Optimistic update
    setPlugins((prev) =>
      prev.map((p) => (p.key === pluginKey ? { ...p, enabled: !currentEnabled } : p))
    );
    try {
      const res = await fetch(`${API_BASE}/plugin/admin/toggle`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ pluginKey, enabled: !currentEnabled }),
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
    } catch (err) {
      // Revert on error
      setPlugins((prev) =>
        prev.map((p) => (p.key === pluginKey ? { ...p, enabled: currentEnabled } : p))
      );
      setError(err instanceof Error ? err.message : 'Toggle failed');
    } finally {
      setTogglingKey(null);
    }
  };

  const enabledCount = plugins.filter((p) => p.enabled).length;
  const disabledCount = plugins.filter((p) => !p.enabled).length;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-800">Plugin Management</h2>
          <p className="text-gray-500">Enable or disable plugins for your tenant</p>
        </div>
        <button
          onClick={loadPlugins}
          className="px-3 py-1.5 text-sm bg-indigo-600 text-white rounded-md hover:bg-indigo-700"
        >
          Reload
        </button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-2 rounded text-sm">
          {error}
        </div>
      )}

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
          <p className="text-sm text-gray-500">Total Plugins</p>
          <p className="text-2xl font-bold text-gray-800">{plugins.length}</p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
          <p className="text-sm text-gray-500">Enabled</p>
          <p className="text-2xl font-bold text-green-600">{enabledCount}</p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
          <p className="text-sm text-gray-500">Disabled</p>
          <p className="text-2xl font-bold text-red-600">{disabledCount}</p>
        </div>
      </div>

      {/* Table */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        {loading ? (
          <div className="p-8 text-center text-gray-500">Loading plugins...</div>
        ) : plugins.length === 0 ? (
          <div className="p-8 text-center text-gray-500">No plugins installed</div>
        ) : (
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Key</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Version</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Toggle</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {plugins.map((plugin) => (
                <tr key={plugin.key} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 font-mono">{plugin.key}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{plugin.name}</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">v{plugin.version}</td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className="px-2 py-1 text-xs font-medium bg-gray-100 text-gray-700 rounded">
                      {plugin.type || 'plugin'}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-600 max-w-md truncate">{plugin.description}</td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span
                      className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                        plugin.enabled ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                      }`}
                    >
                      {plugin.enabled ? 'enabled' : 'disabled'}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <button
                      onClick={() => toggleStatus(plugin.key, plugin.enabled)}
                      disabled={togglingKey === plugin.key}
                      className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                        plugin.enabled ? 'bg-green-500' : 'bg-gray-300'
                      } ${togglingKey === plugin.key ? 'opacity-50 cursor-wait' : ''}`}
                    >
                      <span
                        className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                          plugin.enabled ? 'translate-x-6' : 'translate-x-1'
                        }`}
                      />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
