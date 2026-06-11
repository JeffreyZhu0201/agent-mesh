import { apiRequest } from './client';

/** 插件元信息（与 plugin-svc 返回字段一致） */
export interface PluginInfo {
  key: string;
  name: string;
  version: string;
  type?: string;
  description?: string;
  author?: string;
  enabled: boolean;
}

/** 获取当前租户已启用的插件列表（前端主应用用） */
export async function fetchEnabledPlugins(): Promise<PluginInfo[]> {
  const data = await apiRequest<{ plugins: PluginInfo[] }>('/plugin/list');
  return data.plugins ?? [];
}

/** 获取所有插件及启用状态（管理后台用） */
export async function fetchAdminPlugins(): Promise<PluginInfo[]> {
  const data = await apiRequest<{ plugins: PluginInfo[] }>('/plugin/admin/list');
  return data.plugins ?? [];
}

/** 切换租户级插件启用状态 */
export async function togglePlugin(pluginKey: string, enabled: boolean): Promise<void> {
  await apiRequest('/plugin/admin/toggle', {
    method: 'POST',
    body: JSON.stringify({ pluginKey, enabled }),
  });
}
