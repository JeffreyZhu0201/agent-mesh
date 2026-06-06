/**
 * Plugin SDK Hooks for React
 */

import { useState, useEffect, useCallback, useMemo } from 'react';
import { Plugin, MenuItem, RouteConfig, PluginContext } from './types';

// Module-level global registry
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
