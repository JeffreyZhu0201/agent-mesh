/**
 * Plugin SDK Type Definitions for React
 */

/**
 * Metadata describing a plugin
 */
export interface PluginMetadata {
  name: string;
  version: string;
  description?: string;
  type?: string;
  author?: string;
}

/**
 * A menu item that can be registered by a plugin
 */
export interface MenuItem {
  key: string;
  label: string;
  icon?: string;
  path: string;
  parent?: string;
  order?: number;
}

/**
 * A plugin with its metadata, menu items, and optional routes
 */
export interface Plugin {
  metadata: PluginMetadata;
  menuItems: MenuItem[];
  routes?: RouteConfig[];
  component?: React.ComponentType;
}

/**
 * Route configuration for a plugin's routes
 */
export interface RouteConfig {
  path: string;
  component?: React.ComponentType;
  exact?: boolean;
  meta?: Record<string, unknown>;
}

/**
 * Context provided to plugins at runtime
 */
export interface PluginContext {
  tenantId: string;
  userId: string;
  token: string;
  apiBase: string;
}
