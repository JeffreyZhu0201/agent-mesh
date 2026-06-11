import { API_BASE_URL, fetchJson } from './http';
import { getAuthToken } from './authUtils';

/** 构建带 Authorization 的请求头 */
function authHeaders(extra?: HeadersInit): Headers {
  const headers = new Headers(extra);
  const token = getAuthToken();
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }
  if (!headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }
  return headers;
}

/**
 * 已认证 API 请求（自动附加 Bearer Token）。
 * 适用于 admin、plugin 等需要登录的接口。
 */
export async function apiRequest<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE_URL}${endpoint.startsWith('/') ? endpoint : `/${endpoint}`}`;
  const { data } = await fetchJson<T>(url, {
    ...options,
    headers: authHeaders(options.headers),
  });
  return data;
}
