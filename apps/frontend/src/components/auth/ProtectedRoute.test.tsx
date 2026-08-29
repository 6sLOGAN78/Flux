import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { ProtectedRoute } from './ProtectedRoute';
import { PublicRoute } from './PublicRoute';
import { AuthContext } from './AuthContext';
import { SignInPage } from '@/pages/auth/SignInPage';
import { SignUpPage } from '@/pages/auth/SignUpPage';
import { SSOPage } from '@/pages/auth/SSOPage';

describe('Auth Route Guards', () => {
  it('shows loading spinner when auth state is not loaded', () => {
    const html = renderToString(
      <AuthContext.Provider value={{ isAuthenticated: false, isLoaded: false, token: null }}>
        <MemoryRouter initialEntries={['/dashboard']}>
          <ProtectedRoute>
            <div data-testid="protected-content">Secret Dashboard</div>
          </ProtectedRoute>
        </MemoryRouter>
      </AuthContext.Provider>
    );

    expect(html).toContain('animate-spin');
    expect(html).not.toContain('Secret Dashboard');
  });

  it('redirects unauthenticated users away from protected routes', () => {
    const html = renderToString(
      <AuthContext.Provider value={{ isAuthenticated: false, isLoaded: true, token: null }}>
        <MemoryRouter initialEntries={['/dashboard']}>
          <ProtectedRoute>
            <div data-testid="protected-content">Secret Dashboard</div>
          </ProtectedRoute>
        </MemoryRouter>
      </AuthContext.Provider>
    );

    expect(html).not.toContain('Secret Dashboard');
  });

  it('renders protected content when user is authenticated', () => {
    const html = renderToString(
      <AuthContext.Provider value={{ isAuthenticated: true, isLoaded: true, token: 'mock-jwt-token' }}>
        <MemoryRouter initialEntries={['/dashboard']}>
          <ProtectedRoute>
            <div data-testid="protected-content">Secret Dashboard</div>
          </ProtectedRoute>
        </MemoryRouter>
      </AuthContext.Provider>
    );

    expect(html).toContain('Secret Dashboard');
  });

  it('shows loading spinner in PublicRoute when auth is not loaded', () => {
    const html = renderToString(
      <AuthContext.Provider value={{ isAuthenticated: false, isLoaded: false, token: null }}>
        <MemoryRouter initialEntries={['/sign-in']}>
          <PublicRoute>
            <div data-testid="public-content">Sign In Form</div>
          </PublicRoute>
        </MemoryRouter>
      </AuthContext.Provider>
    );

    expect(html).toContain('animate-spin');
    expect(html).not.toContain('Sign In Form');
  });

  it('renders public content when user is unauthenticated', () => {
    const html = renderToString(
      <AuthContext.Provider value={{ isAuthenticated: false, isLoaded: true, token: null }}>
        <MemoryRouter initialEntries={['/sign-in']}>
          <PublicRoute>
            <div data-testid="public-content">Sign In Form</div>
          </PublicRoute>
        </MemoryRouter>
      </AuthContext.Provider>
    );

    expect(html).toContain('Sign In Form');
  });

  it('redirects authenticated users away from public auth routes', () => {
    const html = renderToString(
      <AuthContext.Provider value={{ isAuthenticated: true, isLoaded: true, token: 'mock-jwt-token' }}>
        <MemoryRouter initialEntries={['/sign-in']}>
          <PublicRoute>
            <div data-testid="public-content">Sign In Form</div>
          </PublicRoute>
        </MemoryRouter>
      </AuthContext.Provider>
    );

    expect(html).not.toContain('Sign In Form');
  });
});

describe('Auth Pages', () => {
  it('renders SignInPage with email form and SSO link', () => {
    const html = renderToString(
      <MemoryRouter>
        <SignInPage />
      </MemoryRouter>
    );

    expect(html).toContain('Welcome back');
    expect(html).toContain('Work Email');
    expect(html).toContain('Single Sign-On (SAML / SSO)');
  });

  it('renders SignUpPage with registration form', () => {
    const html = renderToString(
      <MemoryRouter>
        <SignUpPage />
      </MemoryRouter>
    );

    expect(html).toContain('Create your account');
    expect(html).toContain('Full Name');
    expect(html).toContain('Work Email');
  });

  it('renders SSOPage with company domain input for SAML 2.0', () => {
    const html = renderToString(
      <MemoryRouter>
        <SSOPage />
      </MemoryRouter>
    );

    expect(html).toContain('Enterprise Single Sign-On');
    expect(html).toContain('Company Domain');
    expect(html).toContain('Continue with SSO');
  });

  it('renders with StandaloneAuthProvider providing auth context', () => {
    const { StandaloneAuthProvider } = require('./AuthContext');
    const html = renderToString(
      <StandaloneAuthProvider>
        <MemoryRouter initialEntries={['/dashboard']}>
          <ProtectedRoute>
            <div data-testid="dashboard-view">Flux Main Dashboard</div>
          </ProtectedRoute>
        </MemoryRouter>
      </StandaloneAuthProvider>
    );

    expect(html).toContain('Flux Main Dashboard');
  });
});
