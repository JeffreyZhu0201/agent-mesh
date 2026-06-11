import { API_BASE_URL, fetchJson } from './http';

export interface LoginResponse {
  token: string;
  user: {
    id: string;
    username: string;
    email: string;
    avatar?: string;
    tenantId: number;
    role: string;
  };
}

export interface RegisterResponse {
  user: {
    id: string;
    username: string;
    email: string;
  };
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  password: string;
  email: string;
}

/** 将 user-svc 返回的用户对象规范化为前端 LoginResponse.user 结构 */
function normalizeUser(rawUser: Record<string, unknown> = {}) {
  return {
    id: String(rawUser.id ?? rawUser.userId ?? ''),
    username: String(rawUser.username ?? ''),
    email: String(rawUser.email ?? ''),
    avatar: rawUser.avatar as string | undefined,
    tenantId: Number(rawUser.tenantId ?? 0),
    role: String(rawUser.role ?? 'viewer'),
  };
}

/** 公开接口（登录/注册）：无需 Token */
async function fetchPublic<T>(path: string, options: RequestInit): Promise<T> {
  const { data } = await fetchJson<T>(`${API_BASE_URL}${path}`, options);
  return data;
}

export const authApi = {
  /** 用户登录，返回 JWT 与用户信息 */
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    const payload = await fetchPublic<{ token: string; user?: Record<string, unknown> }>(
      '/user/login',
      { method: 'POST', body: JSON.stringify(credentials) }
    );
    return { token: payload.token, user: normalizeUser(payload.user) };
  },

  /** 用户注册，成功后需自行跳转登录页 */
  async register(data: RegisterRequest): Promise<RegisterResponse> {
    const payload = await fetchPublic<Record<string, unknown>>('/user/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    return {
      user: {
        id: String(payload.id ?? ''),
        username: String(payload.username ?? ''),
        email: String(payload.email ?? ''),
      },
    };
  },
};

// 兼容旧导出
export { ApiError, AuthApiError } from './http';
