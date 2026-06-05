/**
 * Plugin SDK - Public API
 */

// Types
export * from './types';

// Context management
export { setPluginContext, getPluginContext } from './hooks';

// Registry functions
export {
  registerPlugin,
  unregisterPlugin,
  getRegisteredPlugins,
  getMenuItems,
  getRouteConfigs,
} from './hooks';

// React hooks
export { usePlugin, usePlugins, useMenuItems } from './hooks';

import { PluginContext } from './types';

/**
 * Make an authenticated API call through the plugin system
 */
export async function pluginApiCall<T = unknown>(
  endpoint: string,
  options: RequestInit = {},
  context?: PluginContext
): Promise<T> {
  const ctx = context ?? getPluginContext();

  if (!ctx) {
    throw new Error('Plugin context not set. Call setPluginContext() first.');
  }

  const { tenantId, userId, token, apiBase } = ctx;

  const url = `${apiBase.replace(/\/$/, '')}/${endpoint.replace(/^\//, '')}`;

  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    'X-Tenant-ID': tenantId,
    'X-User-ID': userId,
    ...(options.headers || {}),
  };

  if (token) {
    (headers as Record<string, string>)['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(url, {
    ...options,
    headers,
  });

  if (!response.ok) {
    throw new Error(`API call failed: ${response.status} ${response.statusText}`);
  }

  return response.json() as Promise<T>;
}
