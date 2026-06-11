import { beforeEach, describe, expect, it } from 'vitest';
import {
  getMenuItems,
  getRegisteredPlugins,
  getRouteConfigs,
  registerPlugin,
  unregisterPlugin,
} from './hooks';
import type { Plugin } from './types';

function makePlugin(name: string, order = 0): Plugin {
  return {
    metadata: { name, version: '1.0.0' },
    menuItems: [{ key: `${name}-menu`, label: name, path: `/${name}`, order }],
    routes: [{ path: `/${name}` }],
  };
}

describe('plugin-sdk hooks', () => {
  beforeEach(() => {
    getRegisteredPlugins().forEach((plugin) => {
      unregisterPlugin(plugin.metadata.name);
    });
  });

  it('registers and lists plugins', () => {
    registerPlugin(makePlugin('chat'));
    expect(getRegisteredPlugins()).toHaveLength(1);
  });

  it('filters menu items by enabled keys', () => {
    registerPlugin(makePlugin('chat', 2));
    registerPlugin(makePlugin('dashboard', 1));

    const enabled = new Set(['dashboard']);
    const items = getMenuItems(enabled);

    expect(items).toHaveLength(1);
    expect(items[0].key).toBe('dashboard-menu');
  });

  it('sorts menu items by order', () => {
    registerPlugin(makePlugin('chat', 2));
    registerPlugin(makePlugin('dashboard', 1));

    const items = getMenuItems(null);
    expect(items.map((item) => item.key)).toEqual(['dashboard-menu', 'chat-menu']);
  });

  it('returns route configs for enabled plugins only', () => {
    registerPlugin(makePlugin('chat'));
    registerPlugin(makePlugin('dashboard'));

    const routes = getRouteConfigs(new Set(['chat']));
    expect(routes).toHaveLength(1);
    expect(routes[0].path).toBe('/chat');
  });
});
