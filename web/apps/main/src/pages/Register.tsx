import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { TextField } from '@mui/material';
import { authApi, validateRegisterPassword } from '@agentmesh/api';
import { AuthPageLayout } from '../components/AuthPageLayout';

/** 注册表单全部字段 */
interface RegisterForm {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
}

const EMPTY_FORM: RegisterForm = {
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
};

/** 表单字段配置，用于减少重复的 TextField 声明 */
const FORM_FIELDS: Array<{
  key: keyof RegisterForm;
  label: string;
  type?: string;
  autoComplete: string;
  autoFocus?: boolean;
}> = [
  { key: 'username', label: 'Username', autoComplete: 'username', autoFocus: true },
  { key: 'email', label: 'Email Address', type: 'email', autoComplete: 'email' },
  { key: 'password', label: 'Password', type: 'password', autoComplete: 'new-password' },
  {
    key: 'confirmPassword',
    label: 'Confirm Password',
    type: 'password',
    autoComplete: 'new-password',
  },
];

const Register: React.FC = () => {
  const navigate = useNavigate();
  const [form, setForm] = useState<RegisterForm>(EMPTY_FORM);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  /** 更新单个表单字段 */
  const handleFieldChange =
    (key: keyof RegisterForm) => (e: React.ChangeEvent<HTMLInputElement>) => {
      setForm((prev) => ({ ...prev, [key]: e.target.value }));
    };

  /** 提交注册：先客户端校验密码，再调用 authApi 注册 */
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const validationError = validateRegisterPassword(form);
    if (validationError) {
      setError(validationError);
      return;
    }

    setIsLoading(true);
    try {
      await authApi.register({
        username: form.username,
        password: form.password,
        email: form.email,
      });
      navigate('/app/login');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AuthPageLayout
      title="Sign Up"
      subtitle="Create your AgentMesh account"
      error={error}
      onSubmit={handleSubmit}
      submitLabel="Sign Up"
      loadingLabel="Creating account..."
      isLoading={isLoading}
      footerText="Already have an account? Sign In"
      footerTo="/app/login"
    >
      {FORM_FIELDS.map(({ key, label, type, autoComplete, autoFocus }) => (
        <TextField
          key={key}
          margin="normal"
          required
          fullWidth
          id={key}
          name={key}
          label={label}
          type={type ?? 'text'}
          autoComplete={autoComplete}
          autoFocus={autoFocus}
          value={form[key]}
          onChange={handleFieldChange(key)}
        />
      ))}
    </AuthPageLayout>
  );
};

export default Register;
