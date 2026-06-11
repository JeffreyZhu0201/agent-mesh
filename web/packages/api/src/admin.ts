import { apiRequest } from './client';

/** 租户实体（与 user-svc 返回字段一致） */
export interface TenantRecord {
  id: number;
  name: string;
  code: string;
  status: number;
  createdAt?: string;
}

/** 用户实体（与 user-svc 返回字段一致） */
export interface UserRecord {
  id: number;
  username: string;
  email: string;
  tenantId: number;
  role: string;
  status: number;
  createdAt?: string;
}

/** 平台管理员：获取所有租户 */
export async function fetchTenants(): Promise<TenantRecord[]> {
  return apiRequest<TenantRecord[]>('/admin/tenants');
}

/** 平台管理员：创建租户 */
export async function createTenant(name: string, code: string): Promise<TenantRecord> {
  return apiRequest<TenantRecord>('/admin/tenants', {
    method: 'POST',
    body: JSON.stringify({ name, code }),
  });
}

/** 平台管理员：更新租户 */
export async function updateTenant(
  id: number,
  data: Partial<Pick<TenantRecord, 'name' | 'code' | 'status'>>
): Promise<TenantRecord> {
  return apiRequest<TenantRecord>(`/admin/tenants/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

/** 平台管理员：删除租户 */
export async function deleteTenant(id: number): Promise<void> {
  await apiRequest(`/admin/tenants/${id}`, { method: 'DELETE' });
}

/** 获取用户列表（平台管理员或租户管理员，由后端路由区分） */
export async function fetchUsers(scope: 'platform' | 'tenant'): Promise<UserRecord[]> {
  const endpoint = scope === 'platform' ? '/admin/users' : '/tenant/users';
  return apiRequest<UserRecord[]>(endpoint);
}

/** 创建用户 */
export async function createUser(
  scope: 'platform' | 'tenant',
  data: { username: string; password: string; email?: string; tenantId?: number; role?: string }
): Promise<UserRecord> {
  const endpoint = scope === 'platform' ? '/admin/users' : '/tenant/users';
  return apiRequest<UserRecord>(endpoint, {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

/** 更新用户 */
export async function updateUser(
  scope: 'platform' | 'tenant',
  id: number,
  data: Partial<Pick<UserRecord, 'email' | 'role' | 'status'>>
): Promise<UserRecord> {
  const endpoint = scope === 'platform' ? `/admin/users/${id}` : `/tenant/users/${id}`;
  return apiRequest<UserRecord>(endpoint, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

/** 删除用户 */
export async function deleteUser(scope: 'platform' | 'tenant', id: number): Promise<void> {
  const endpoint = scope === 'platform' ? `/admin/users/${id}` : `/tenant/users/${id}`;
  await apiRequest(endpoint, { method: 'DELETE' });
}
