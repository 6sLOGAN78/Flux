import React from 'react';
import { Navigate } from 'react-router-dom';

export function SSOPage() {
  return <Navigate to="/login" replace />;
}

export default SSOPage;
