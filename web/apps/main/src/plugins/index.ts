/**
 * ==================== 插件注册中心说明 ====================
 *
 * 这是内置插件的注册中心文件。
 *
 * 核心机制 - 导入即注册:
 * 导入此文件会触发下方列出的所有插件的自我注册。
 * 每个插件在自己的 index.ts 中调用 registerPlugin()，将自己添加到全局注册表中。
 * 注册后，插件的菜单项和路由配置会通过 App.tsx 中的 useMenuItems() 和 getRouteConfigs() 暴露出来。
 *
 * 插件过滤机制:
 * - 所有插件都会注册到全局注册表
 * - 但只有 enabledKeys (后端返回的已启用插件列表) 中的插件会被显示
 * - 这是通过 getRouteConfigs(enabledKeys) 和 useMenuItems(enabledKeys) 实现的
 *
 * 如何添加新插件:
 * 1. 在 src/plugins/<name>/ 目录下创建插件文件夹
 * 2. 创建 index.ts，在其中调用 registerPlugin() 注册插件
 * 3. 在本文件中导入该插件 (如: import './chat';)
 *
 * ==================== 插件注册中心说明 END ====================
 */

/**
 * Built-in plugins registry.
 *
 * Importing this module self-registers every plugin listed below
 * via the @agentmesh/plugin-sdk registry. Plugins are then surfaced
 * through useMenuItems() and getRouteConfigs() in App.tsx.
 *
 * To add a new plugin, drop a folder under src/plugins/<name>/
 * with an `index.ts` that calls registerPlugin(), then import it here.
 */
// Chat plugin is temporarily disabled (agent-svc kept as optional)
// import './chat';
