const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'ApiError';
  }
}

function getAuthToken(): string | null {
  const stored = localStorage.getItem('auth-storage');
  if (stored) {
    try {
      const parsed = JSON.parse(stored);
      return parsed.state?.token || null;
    } catch {
      return null;
    }
  }
  return null;
}

function getTenantId(): string | null {
  return localStorage.getItem('tenant-id');
}

export async function apiClient(
  endpoint: string,
  options: RequestInit = {}
): Promise<unknown> {
  const token = getAuthToken();
  const tenantId = getTenantId();

  const headers = new Headers(options.headers);
  if (!headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  if (tenantId) {
    headers.set('X-Tenant-ID', tenantId);
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ message: 'Request failed' }));
    throw new ApiError(response.status, errorData.message || 'Request failed');
  }

  return response.json();
}

export { API_BASE_URL, ApiError };
