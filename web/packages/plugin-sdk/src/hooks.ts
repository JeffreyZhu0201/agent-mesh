/**
 * ==================== Plugin SDK Hooks 说明 ====================
 *
 * 这是插件系统的核心 SDK，提供了以下功能:
 * 1. registerPlugin() - 注册插件到全局注册表
 * 2. getRouteConfigs(enabledKeys) - 获取已启用插件的路由配置
 * 3. useMenuItems(enabledKeys) - 获取已启用插件的菜单项
 *
 * 核心概念:
 * - 全局注册表: 所有已注册的插件存储在 registeredPlugins Map 中
 * - enabledKeys 过滤: 只有 enabledKeys 中包含的插件才会被暴露给前端
 * - 多租户隔离: 不同租户有不同的 enabledKeys，实现插件级别的访问控制
 *
 * ==================== Plugin SDK Hooks 说明 END ====================
 */

/**
 * Plugin SDK Hooks for React
 */

import { useState, useEffect, useCallback, useMemo } from 'react';
import { Plugin, MenuItem, RouteConfig, PluginContext } from './types';

// ==================== 全局注册表 ====================
// registeredPlugins: Map<string, Plugin>
// - 键: 插件名称 (plugin.metadata.name)
// - 值: 插件对象 (包含 metadata, menuItems, routes 等)
// - 这是模块级别的全局注册表，所有插件都在此处注册
// ==================== 全局注册表 END ====================
const registeredPlugins: Map<string, Plugin> = new Map();
let globalPluginContext: PluginContext | null = null;

/**
 * Set the global plugin context
 */
export function setPluginContext(context: PluginContext): void {
  globalPluginContext = context;
}

/**
 * Get the global plugin context
 */
export function getPluginContext(): PluginContext | null {
  return globalPluginContext;
}

// ==================== registerPlugin() 说明 ====================
// registerPlugin(plugin: Plugin): void
// 作用: 将插件添加到全局注册表
// 参数: plugin - 插件对象，包含:
//   - metadata: { name, version, description }
//   - menuItems: 菜单项数组
//   - routes: 路由配置数组
// 调用时机: 插件的 index.ts 导入时自动调用
// ==================== registerPlugin() 说明 END ====================
/**
 * Register a plugin with its menu items and routes
 */
export function registerPlugin(plugin: Plugin): void {
  registeredPlugins.set(plugin.metadata.name, plugin);
}

/**
 * Unregister a plugin by name
 */
export function unregisterPlugin(name: string): void {
  registeredPlugins.delete(name);
}

/**
 * Get all registered plugins
 */
export function getRegisteredPlugins(): Plugin[] {
  return Array.from(registeredPlugins.values());
}

/**
 * Get all menu items from all registered plugins, sorted by order.
 * If `enabledKeys` is provided, only items from plugins whose metadata.name
 * is in the set are returned. Pass null to skip filtering (default).
 */
export function getMenuItems(enabledKeys?: Set<string> | null): MenuItem[] {
  const allMenuItems: MenuItem[] = [];

  registeredPlugins.forEach((plugin) => {
    if (enabledKeys && !enabledKeys.has(plugin.metadata.name)) return;
    allMenuItems.push(...plugin.menuItems);
  });

  return allMenuItems.sort((a, b) => (a.order ?? 0) - (b.order ?? 0));
}

// ==================== getRouteConfigs(enabledKeys) 说明 ====================
// getRouteConfigs(enabledKeys?: Set<string> | null): RouteConfig[]
// 作用: 获取已注册插件的路由配置数组
// 参数: enabledKeys - 已启用插件名称的 Set，用于过滤
//       - 如果传入 null 或 undefined，返回所有已注册插件的路由
//       - 如果传入 Set，只返回该 Set 中包含的插件路由
// 返回值: RouteConfig[] - 路由配置数组，每个配置包含:
//   - path: 路由路径 (如 /app/chat)
//   - component: 插件组件
//
// 使用场景: 在 App.tsx 的 AppRoutes 组件中调用，根据后端返回的 enabledKeys
//           动态生成只有已启用插件才有的路由
// ==================== getRouteConfigs(enabledKeys) 说明 END ====================
/**
 * Get all route configs from all registered plugins.
 * If `enabledKeys` is provided, only routes from enabled plugins are returned.
 */
export function getRouteConfigs(enabledKeys?: Set<string> | null): RouteConfig[] {
  const allRoutes: RouteConfig[] = [];

  registeredPlugins.forEach((plugin) => {
    if (enabledKeys && !enabledKeys.has(plugin.metadata.name)) return;
    if (plugin.routes) {
      allRoutes.push(...plugin.routes);
    }
  });

  return allRoutes;
}

/**
 * Hook to get a specific plugin by name
 */
export function usePlugin(name: string): Plugin | undefined {
  const [plugin, setPlugin] = useState<Plugin | undefined>(() => registeredPlugins.get(name));

  useEffect(() => {
    // Subscribe to changes by re-reading on each render
    const currentPlugin = registeredPlugins.get(name);
    if (currentPlugin !== plugin) {
      setPlugin(currentPlugin);
    }
  }, [name]);

  useEffect(() => {
    // Set up a simple observer pattern for updates
    const checkPlugin = setInterval(() => {
      const currentPlugin = registeredPlugins.get(name);
      if (currentPlugin !== plugin) {
        setPlugin(currentPlugin);
      }
    }, 1000);

    return () => clearInterval(checkPlugin);
  }, [name, plugin]);

  return plugin;
}

/**
 * Hook to get all registered plugins (reactive)
 */
export function usePlugins(): Plugin[] {
  const [plugins, setPlugins] = useState<Plugin[]>(() => getRegisteredPlugins());

  useEffect(() => {
    const updatePlugins = () => {
      setPlugins(getRegisteredPlugins());
    };

    const interval = setInterval(updatePlugins, 1000);
    return () => clearInterval(interval);
  }, []);

  return plugins;
}

// ==================== useMenuItems(enabledKeys) 说明 ====================
// useMenuItems(enabledKeys?: Set<string> | null): MenuItem[]
// 作用: React Hook，获取已启用插件的菜单项
// 参数: enabledKeys - 已启用插件名称的 Set，用于过滤
//       - 如果传入 null 或 undefined，返回所有已注册插件的菜单项
//       - 如果传入 Set，只返回该 Set 中包含的插件的菜单项
// 返回值: MenuItem[] - 菜单项数组，按 order 排序
//
// 过滤机制:
// enabledKeys 实现了"多租户插件隔离"。例如:
// - 租户 A 启用了 [chat, analysis]，则只看到这两个插件的菜单
// - 租户 B 启用了 [chat, image], 则只看到这两个插件的菜单
// - admin 用户可能有不同的 enabledKeys，看到不同的菜单
//
// 这确保了插件级别的访问控制，后端控制前端的插件可见性
// ==================== useMenuItems(enabledKeys) 说明 END ====================
/**
 * Hook to get all menu items (sorted by order, reactive).
 * If `enabledKeys` is provided, only menu items from enabled plugins are shown.
 */
export function useMenuItems(enabledKeys?: Set<string> | null): MenuItem[] {
  const [menuItems, setMenuItems] = useState<MenuItem[]>(() => getMenuItems(enabledKeys));

  useEffect(() => {
    setMenuItems(getMenuItems(enabledKeys));
    const updateMenuItems = () => {
      setMenuItems(getMenuItems(enabledKeys));
    };

    const interval = setInterval(updateMenuItems, 1000);
    return () => clearInterval(interval);
    // Re-run when the set of enabled keys changes
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [enabledKeys]);

  return menuItems;
}
