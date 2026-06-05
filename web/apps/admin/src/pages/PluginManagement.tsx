import { useState } from 'react';

interface Plugin {
  id: string;
  name: string;
  version: string;
  type: string;
  status: 'enabled' | 'disabled';
}

const initialPlugins: Plugin[] = [
  { id: '1', name: 'Authentication Plugin', version: '2.1.0', type: 'security', status: 'enabled' },
  { id: '2', name: 'Logger Plugin', version: '1.5.2', type: 'monitoring', status: 'enabled' },
  { id: '3', name: 'Cache Plugin', version: '3.0.1', type: 'performance', status: 'disabled' },
  { id: '4', name: 'Notification Plugin', version: '1.2.0', type: 'communication', status: 'enabled' },
  { id: '5', name: 'Analytics Plugin', version: '2.0.0', type: 'analytics', status: 'enabled' },
  { id: '6', name: 'Legacy Connector', version: '0.9.5', type: 'integration', status: 'disabled' },
];

export function PluginManagement() {
  const [plugins, setPlugins] = useState<Plugin[]>(initialPlugins);

  const toggleStatus = (id: string) => {
    setPlugins(
      plugins.map((plugin) =>
        plugin.id === id
          ? {
              ...plugin,
              status: plugin.status === 'enabled' ? 'disabled' : 'enabled',
            }
          : plugin
      )
    );
  };

  const enabledCount = plugins.filter((p) => p.status === 'enabled').length;
  const disabledCount = plugins.filter((p) => p.status === 'disabled').length;

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-800">Plugin Management</h2>
        <p className="text-gray-500">Manage system plugins and their status</p>
      </div>

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
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                ID
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Name
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Version
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Type
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Status
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {plugins.map((plugin) => (
              <tr key={plugin.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {plugin.id}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  {plugin.name}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  v{plugin.version}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className="px-2 py-1 text-xs font-medium bg-gray-100 text-gray-700 rounded">
                    {plugin.type}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span
                    className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                      plugin.status === 'enabled'
                        ? 'bg-green-100 text-green-800'
                        : 'bg-red-100 text-red-800'
                    }`}
                  >
                    {plugin.status}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <button
                    onClick={() => toggleStatus(plugin.id)}
                    className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                      plugin.status === 'enabled' ? 'bg-green-500' : 'bg-gray-300'
                    }`}
                  >
                    <span
                      className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                        plugin.status === 'enabled' ? 'translate-x-6' : 'translate-x-1'
                      }`}
                    />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
