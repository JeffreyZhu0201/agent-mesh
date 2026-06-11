import React from 'react';
import {
  Box,
  Button,
  Container,
  Link,
  Typography,
  Alert,
  Paper,
} from '@mui/material';
import { Link as RouterLink } from 'react-router-dom';

interface AuthPageLayoutProps {
  /** 页面主标题 */
  title: string;
  /** 副标题说明 */
  subtitle: string;
  /** 错误提示（有值时显示 Alert） */
  error: string | null;
  /** 表单提交回调 */
  onSubmit: (e: React.FormEvent) => void;
  /** 提交按钮文案 */
  submitLabel: string;
  /** 提交中按钮文案 */
  loadingLabel: string;
  /** 是否正在提交 */
  isLoading: boolean;
  /** 表单字段区域 */
  children: React.ReactNode;
  /** 底部链接文案 */
  footerText: string;
  /** 底部链接目标路径 */
  footerTo: string;
}

/** 登录/注册页共享布局：居中 Paper + 标题 + 错误提示 + 表单 + 底部链接 */
export const AuthPageLayout: React.FC<AuthPageLayoutProps> = ({
  title,
  subtitle,
  error,
  onSubmit,
  submitLabel,
  loadingLabel,
  isLoading,
  children,
  footerText,
  footerTo,
}) => (
  <Container maxWidth="sm" sx={{ minHeight: '100vh', display: 'flex', alignItems: 'center' }}>
    <Paper sx={{ p: 4, width: '100%' }}>
      <Box sx={{ textAlign: 'center', mb: 3 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          {title}
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {subtitle}
        </Typography>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      <Box component="form" onSubmit={onSubmit} sx={{ mt: 1 }}>
        {children}
        <Button
          type="submit"
          fullWidth
          variant="contained"
          sx={{ mt: 3, mb: 2 }}
          disabled={isLoading}
        >
          {isLoading ? loadingLabel : submitLabel}
        </Button>

        <Box sx={{ textAlign: 'center' }}>
          <Link component={RouterLink} to={footerTo} variant="body2">
            {footerText}
          </Link>
        </Box>
      </Box>
    </Paper>
  </Container>
);

export default AuthPageLayout;
