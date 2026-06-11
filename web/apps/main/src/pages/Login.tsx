import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { TextField } from '@mui/material';
import { useAuthStore } from '@agentmesh/stores';
import { AuthPageLayout } from '../components/AuthPageLayout';

/** 登录表单字段配置 */
const LOGIN_FIELDS = [
  { key: 'username' as const, label: 'Username', type: 'text', autoComplete: 'username', autoFocus: true },
  { key: 'password' as const, label: 'Password', type: 'password', autoComplete: 'current-password' },
];

const Login: React.FC = () => {
  const navigate = useNavigate();
  const login = useAuthStore((state) => state.login);
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const values = { username, password };
  const setters = { username: setUsername, password: setPassword };

  /** 提交登录表单 */
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsLoading(true);

    try {
      await login(username, password);
      navigate('/app');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AuthPageLayout
      title="Sign In"
      subtitle="Welcome back to AgentMesh"
      error={error}
      onSubmit={handleSubmit}
      submitLabel="Sign In"
      loadingLabel="Signing in..."
      isLoading={isLoading}
      footerText="Don't have an account? Sign Up"
      footerTo="/app/register"
    >
      {LOGIN_FIELDS.map(({ key, label, type, autoComplete, autoFocus }) => (
        <TextField
          key={key}
          margin="normal"
          required
          fullWidth
          id={key}
          name={key}
          label={label}
          type={type}
          autoComplete={autoComplete}
          autoFocus={autoFocus}
          value={values[key]}
          onChange={(e) => setters[key](e.target.value)}
        />
      ))}
    </AuthPageLayout>
  );
};

export default Login;
