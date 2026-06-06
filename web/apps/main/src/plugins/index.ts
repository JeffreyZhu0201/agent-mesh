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
import './chat';
