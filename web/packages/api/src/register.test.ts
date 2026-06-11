import { describe, expect, it } from 'vitest';
import { MIN_PASSWORD_LENGTH, validateRegisterPassword } from './register';

describe('validateRegisterPassword', () => {
  it('密码不一致时返回错误', () => {
    expect(
      validateRegisterPassword({ password: 'secret1', confirmPassword: 'secret2' })
    ).toBe('Passwords do not match');
  });

  it('密码过短时返回错误', () => {
    expect(
      validateRegisterPassword({ password: '12345', confirmPassword: '12345' })
    ).toBe(`Password must be at least ${MIN_PASSWORD_LENGTH} characters`);
  });

  it('密码合法时返回 null', () => {
    expect(
      validateRegisterPassword({ password: 'secret1', confirmPassword: 'secret1' })
    ).toBeNull();
  });

  it('边界长度（恰好 MIN_PASSWORD_LENGTH）应通过', () => {
    const pwd = 'a'.repeat(MIN_PASSWORD_LENGTH);
    expect(validateRegisterPassword({ password: pwd, confirmPassword: pwd })).toBeNull();
  });
});
