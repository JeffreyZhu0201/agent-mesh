import React from 'react';
import { Navigate } from 'react-router-dom';
import { useAuthStore } from '@agentmesh/stores';

interface ProtectedRouteProps {
  children: React.ReactNode;
}

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children }) => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  if (!isAuthenticated) {
    return <Navigate to="/app/login" replace />;
  }

  return <>{children}</>;
};

export default ProtectedRoute;
