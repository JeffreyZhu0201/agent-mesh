/** 与后端 user-svc 保持一致的最低密码长度 */
export const MIN_PASSWORD_LENGTH = 6;

/** 注册表单中需要客户端校验的字段 */
export interface RegisterFormInput {
  password: string;
  confirmPassword: string;
}

/**
 * 校验注册表单的密码规则（确认密码一致、长度达标）。
 * 用户名和邮箱由 HTML required 与后端二次校验，此处只处理密码逻辑。
 *
 * @returns 错误提示文案；全部通过则返回 null
 */
export function validateRegisterPassword(input: RegisterFormInput): string | null {
  if (input.password !== input.confirmPassword) {
    return 'Passwords do not match';
  }
  if (input.password.length < MIN_PASSWORD_LENGTH) {
    return `Password must be at least ${MIN_PASSWORD_LENGTH} characters`;
  }
  return null;
}
