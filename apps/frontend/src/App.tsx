import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import { queryClient } from '@/api/queryClient';
import { ClerkProvider } from '@clerk/clerk-react';
import { env } from '@/config/env';
import { ClerkAuthProviderWrapper } from '@/components/auth/AuthContext';
import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { PublicRoute } from '@/components/auth/PublicRoute';
import { AppLayout } from '@/components/layout/AppLayout';
import { LandingPage } from '@/pages/public/LandingPage';
import { PricingPage } from '@/pages/public/PricingPage';
import { SignInPage } from '@/pages/auth/SignInPage';
import { SignUpPage } from '@/pages/auth/SignUpPage';
import { SSOPage } from '@/pages/auth/SSOPage';

export function App() {
  const clerkPubKey = env.VITE_CLERK_PUBLISHABLE_KEY;

  return (
    <QueryClientProvider client={queryClient}>
      <ClerkProvider publishableKey={clerkPubKey || 'pk_test_placeholder'}>
        <ClerkAuthProviderWrapper>
          <BrowserRouter>
            <Routes>
              {/* Public Marketing Routes */}
              <Route path="/" element={<LandingPage />} />
              <Route path="/pricing" element={<PricingPage />} />

              {/* Public Auth Routes */}
              <Route
                path="/sign-in"
                element={
                  <PublicRoute>
                    <SignInPage />
                  </PublicRoute>
                }
              />
              <Route
                path="/sign-up"
                element={
                  <PublicRoute>
                    <SignUpPage />
                  </PublicRoute>
                }
              />
              <Route
                path="/sso"
                element={
                  <PublicRoute>
                    <SSOPage />
                  </PublicRoute>
                }
              />

              {/* Protected App Routes */}
              <Route
                path="/dashboard"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <div className="space-y-6">
                        <div>
                          <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
                            Overview
                          </h1>
                          <p className="text-xs text-zinc-500 dark:text-zinc-400">
                            Real-time platform metrics and active link routing status.
                          </p>
                        </div>
                        <div className="rounded-xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
                          <p className="text-xs text-zinc-500">
                            Select a workspace to view your real-time ClickHouse metrics and links.
                          </p>
                        </div>
                      </div>
                    </AppLayout>
                  </ProtectedRoute>
                }
              />

              <Route
                path="/links"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <div className="space-y-6">
                        <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
                          Links
                        </h1>
                      </div>
                    </AppLayout>
                  </ProtectedRoute>
                }
              />

              <Route
                path="/analytics"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <div className="space-y-6">
                        <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
                          Analytics
                        </h1>
                      </div>
                    </AppLayout>
                  </ProtectedRoute>
                }
              />

              {/* Catch-all fallback */}
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </BrowserRouter>
        </ClerkAuthProviderWrapper>
      </ClerkProvider>
    </QueryClientProvider>
  );
}

export default App;
