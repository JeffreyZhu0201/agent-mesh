/** authStore persist 使用的 localStorage 键名 */
export const AUTH_STORAGE_KEY = 'auth-storage';

interface PersistedAuthState {
  state?: {
    token?: string;
    user?: { role?: string };
  };
}

/** 从 localStorage 读取 JWT（与 zustand persist 结构一致） */
export function getAuthToken(): string | null {
  try {
    const stored = localStorage.getItem(AUTH_STORAGE_KEY);
    if (!stored) return null;
    const parsed = JSON.parse(stored) as PersistedAuthState;
    return parsed.state?.token ?? null;
  } catch {
    return null;
  }
}

/** 从 localStorage 读取当前用户角色 */
export function getUserRole(): string | null {
  try {
    const stored = localStorage.getItem(AUTH_STORAGE_KEY);
    if (!stored) return null;
    const parsed = JSON.parse(stored) as PersistedAuthState;
    return parsed.state?.user?.role ?? null;
  } catch {
    return null;
  }
}
