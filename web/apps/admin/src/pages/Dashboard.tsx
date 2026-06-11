import { useEffect, useState } from 'react';
import { Building2, Puzzle, Users, Activity } from 'lucide-react';
import { fetchTenants, fetchUsers, getAuthToken, getUserRole } from '@agentmesh/api';

interface Stats {
  totalTenants: number;
  activeTenants: number;
  totalUsers: number;
  platformAdminCount: number;
  adminCount: number;
  userCount: number;
  viewerCount: number;
}

export function Dashboard() {
  const [stats, setStats] = useState<Stats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const userRole = getUserRole();

  useEffect(() => {
    if (userRole !== 'platform_admin') {
      setLoading(false);
      return;
    }

    async function fetchData() {
      const token = getAuthToken();
      if (!token) {
        setError('Not authenticated');
        setLoading(false);
        return;
      }

      try {
        const [tenants, users] = await Promise.all([fetchTenants(), fetchUsers('platform')]);

        setStats({
          totalTenants: tenants.length,
          activeTenants: tenants.filter((t) => t.status === 1).length,
          totalUsers: users.length,
          platformAdminCount: users.filter((u) => u.role === 'platform_admin').length,
          adminCount: users.filter((u) => u.role === 'admin').length,
          userCount: users.filter((u) => u.role === 'user').length,
          viewerCount: users.filter((u) => u.role === 'viewer').length,
        });
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    }

    fetchData();
  }, [userRole]);

  if (userRole !== 'platform_admin') {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-red-600 mb-2">Access Denied</h2>
          <p className="text-gray-500">Only platform_admin can view this dashboard.</p>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">Loading statistics...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-red-600 mb-2">Error</h2>
          <p className="text-gray-500">{error}</p>
        </div>
      </div>
    );
  }

  const statCards = [
    {
      name: 'Total Tenants',
      value: stats?.totalTenants ?? 0,
      change: `${stats?.activeTenants ?? 0} active`,
      icon: Building2,
      bgColor: 'bg-blue-500',
    },
    {
      name: 'Active Tenants',
      value: stats?.activeTenants ?? 0,
      change: 'Currently operational',
      icon: Activity,
      bgColor: 'bg-emerald-500',
    },
    {
      name: 'Total Users',
      value: stats?.totalUsers ?? 0,
      change: 'All roles',
      icon: Users,
      bgColor: 'bg-purple-500',
    },
    {
      name: 'Plugin-like Stats',
      value: stats?.platformAdminCount ?? 0,
      change: `${stats?.adminCount ?? 0} admins, ${stats?.userCount ?? 0} users, ${stats?.viewerCount ?? 0} viewers`,
      icon: Puzzle,
      bgColor: 'bg-green-500',
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-800">Dashboard</h2>
        <p className="text-gray-500">Overview of your AgentMesh system</p>
      </div>

      {/* Stats cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((stat) => (
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

      {/* Users by role breakdown */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-semibold text-gray-800">Users by Role</h3>
        </div>
        <div className="px-6 py-4">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="text-center p-4 bg-gray-50 rounded-lg">
              <p className="text-2xl font-bold text-gray-900">{stats?.platformAdminCount ?? 0}</p>
              <p className="text-sm text-gray-500">Platform Admins</p>
            </div>
            <div className="text-center p-4 bg-gray-50 rounded-lg">
              <p className="text-2xl font-bold text-gray-900">{stats?.adminCount ?? 0}</p>
              <p className="text-sm text-gray-500">Admins</p>
            </div>
            <div className="text-center p-4 bg-gray-50 rounded-lg">
              <p className="text-2xl font-bold text-gray-900">{stats?.userCount ?? 0}</p>
              <p className="text-sm text-gray-500">Users</p>
            </div>
            <div className="text-center p-4 bg-gray-50 rounded-lg">
              <p className="text-2xl font-bold text-gray-900">{stats?.viewerCount ?? 0}</p>
              <p className="text-sm text-gray-500">Viewers</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
