import { beforeEach, describe, expect, it } from 'vitest';
import { AUTH_STORAGE_KEY, getAuthToken, getUserRole } from './authUtils';

describe('authUtils', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('无存储时返回 null', () => {
    expect(getAuthToken()).toBeNull();
    expect(getUserRole()).toBeNull();
  });

  it('从 persist 结构读取 token 与 role', () => {
    localStorage.setItem(
      AUTH_STORAGE_KEY,
      JSON.stringify({
        state: { token: 'jwt-abc', user: { role: 'admin' } },
      })
    );
    expect(getAuthToken()).toBe('jwt-abc');
    expect(getUserRole()).toBe('admin');
  });

  it('JSON 损坏时返回 null', () => {
    localStorage.setItem(AUTH_STORAGE_KEY, 'not-json');
    expect(getAuthToken()).toBeNull();
    expect(getUserRole()).toBeNull();
  });
});
