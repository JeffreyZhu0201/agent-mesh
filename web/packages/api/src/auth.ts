/*
 * @Author: Jeffrey Zhu JeffreyZhu0201@gmail.com
 * @Date: 2026-06-05 15:57:39
 * @LastEditors: Jeffrey Zhu JeffreyZhu0201@gmail.com
 * @LastEditTime: 2026-06-11 17:51:08
 * @FilePath: /AgentMesh/web/packages/api/src/auth.ts
 * @Description: 
 * 
 * Copyright (c) 2026 by JeffreyZhu, All Rights Reserved. 
 */
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

export interface LoginResponse {
  token: string;
  user: {
    id: string;
    username: string;
    email: string;
    avatar?: string;
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

class AuthApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = 'AuthApiError';
  }
}

async function fetchWithAuth(url: string, options: RequestInit = {}) {
  const token = localStorage.getItem('auth-storage');
  const headers = new Headers(options.headers);
  headers.set('Content-Type', 'application/json');

  if (token) {
    try {
      const parsed = JSON.parse(token);
      if (parsed.state?.token) {
        headers.set('Authorization', `Bearer ${parsed.state.token}`);
      }
    } catch {
      // ignore parse errors
    }
  }

  const response = await fetch(url, {
    ...options,
    headers,
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }));
    throw new AuthApiError(response.status, error.message || 'Request failed');
  }

  return response.json();
}

async function fetchPublic(url: string, options: RequestInit = {}) {
  const headers = new Headers(options.headers);
  headers.set('Content-Type', 'application/json');

  const response = await fetch(url, { ...options, headers });
  const body = await response.json().catch(() => ({ error: 'Request failed' }));

  if (!response.ok) {
    throw new AuthApiError(response.status, body.error || body.message || 'Request failed');
  }

  return body.data ?? body;
}

export const authApi = {
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    const payload = await fetchPublic(`${API_BASE_URL}/user/login`, {
      method: 'POST',
      body: JSON.stringify(credentials),
    });
    return {
      token: payload.token,
      user: {
        id: String(payload.user?.id ?? payload.user?.userId ?? ''),
        username: payload.user?.username ?? '',
        email: payload.user?.email ?? '',
        avatar: payload.user?.avatar,
      },
    };
  },

  async register(data: RegisterRequest): Promise<RegisterResponse> {
    const payload = await fetchPublic(`${API_BASE_URL}/user/register`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
    return {
      user: {
        id: String(payload.id ?? ''),
        username: payload.username ?? '',
        email: payload.email ?? '',
      },
    };
  },
};
