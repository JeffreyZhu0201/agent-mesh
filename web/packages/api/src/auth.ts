const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:3001/api';

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

export const authApi = {
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    return fetchWithAuth(`${API_BASE_URL}/auth/login`, {
      method: 'POST',
      body: JSON.stringify(credentials),
    });
  },

  async register(data: RegisterRequest): Promise<RegisterResponse> {
    return fetchWithAuth(`${API_BASE_URL}/auth/register`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },
};
