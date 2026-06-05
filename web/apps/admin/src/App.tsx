import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AdminLayout } from './layouts/AdminLayout';
import { Dashboard } from './pages/Dashboard';
import { TenantManagement } from './pages/TenantManagement';
import { PluginManagement } from './pages/PluginManagement';
import { UserManagement } from './pages/UserManagement';

export function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/admin" element={<AdminLayout />}>
          <Route index element={<Dashboard />} />
          <Route path="tenants" element={<TenantManagement />} />
          <Route path="plugins" element={<PluginManagement />} />
          <Route path="users" element={<UserManagement />} />
        </Route>
        <Route path="/" element={<Navigate to="/admin" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;