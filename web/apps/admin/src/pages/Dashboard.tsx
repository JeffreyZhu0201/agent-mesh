import { Building2, Puzzle, Users, Activity } from 'lucide-react';

const stats = [
  {
    name: 'Total Tenants',
    value: '12',
    change: '+2 this month',
    icon: Building2,
    bgColor: 'bg-blue-500',
  },
  {
    name: 'Active Plugins',
    value: '48',
    change: '+5 this week',
    icon: Puzzle,
    bgColor: 'bg-green-500',
  },
  {
    name: 'Total Users',
    value: '1,234',
    change: '+128 this month',
    icon: Users,
    bgColor: 'bg-purple-500',
  },
  {
    name: 'System Health',
    value: '99.9%',
    change: 'All systems operational',
    icon: Activity,
    bgColor: 'bg-emerald-500',
  },
];

const recentActivity = [
  { id: 1, action: 'New tenant created', target: 'Acme Corp', time: '5 minutes ago' },
  { id: 2, action: 'Plugin enabled', target: 'Auth Plugin v2.1', time: '15 minutes ago' },
  { id: 3, action: 'User added', target: 'john@acme.com', time: '1 hour ago' },
  { id: 4, action: 'Tenant schema updated', target: 'TechStart Inc', time: '2 hours ago' },
  { id: 5, action: 'Plugin disabled', target: 'Old Logger v1.0', time: '3 hours ago' },
];

export function Dashboard() {
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-800">Dashboard</h2>
        <p className="text-gray-500">Overview of your AgentMesh system</p>
      </div>

      {/* Stats cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {stats.map((stat) => (
          <div
            key={stat.name}
            className="bg-white rounded-lg shadow-sm border border-gray-200 p-6"
          >
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-500">{stat.name}</p>
                <p className="text-3xl font-bold text-gray-900 mt-1">{stat.value}</p>
              </div>
              <div className={`${stat.bgColor} p-3 rounded-lg`}>
                <stat.icon className="w-6 h-6 text-white" />
              </div>
            </div>
            <p className="text-sm text-gray-500 mt-4">{stat.change}</p>
          </div>
        ))}
      </div>

      {/* Recent activity */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-semibold text-gray-800">Recent Activity</h3>
        </div>
        <div className="divide-y divide-gray-100">
          {recentActivity.map((activity) => (
            <div
              key={activity.id}
              className="px-6 py-4 flex items-center justify-between"
            >
              <div>
                <p className="text-sm font-medium text-gray-800">{activity.action}</p>
                <p className="text-sm text-gray-500">{activity.target}</p>
              </div>
              <span className="text-sm text-gray-400">{activity.time}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
