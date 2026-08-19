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
import { OverviewPage } from '@/pages/dashboard/OverviewPage';
import { LinksListPage } from '@/pages/links/LinksListPage';
import { LinkDetailPage } from '@/pages/links/LinkDetailPage';
import { CategoriesPage } from '@/pages/links/CategoriesPage';
import { CampaignsPage } from '@/pages/growth/CampaignsPage';
import { SmartRoutingPage } from '@/pages/growth/SmartRoutingPage';
import { ABTestingPage } from '@/pages/growth/ABTestingPage';
import { DomainsPage } from '@/pages/growth/DomainsPage';
import { AnalyticsPage } from '@/pages/analytics/AnalyticsPage';
import { AttributionPage } from '@/pages/analytics/AttributionPage';
import { QRStudioCanvas } from '@/components/qr/QRStudioCanvas';

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
                      <OverviewPage />
                    </AppLayout>
                  </ProtectedRoute>
                }
              />

              <Route
                path="/links"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <LinksListPage />
                    </AppLayout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/links/:id"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <LinkDetailPage />
                    </AppLayout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/qr-studio"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <div className="space-y-6">
                        <div>
                          <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
                            QR Studio
                          </h1>
                          <p className="text-xs text-zinc-500 dark:text-zinc-400">
                            Design, brand, and export high-resolution vector QR codes.
                          </p>
                        </div>
                        <QRStudioCanvas />
                      </div>
                    </AppLayout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/categories"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <CategoriesPage />
                    </AppLayout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/campaigns"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <CampaignsPage />
                    </AppLayout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/routing"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <SmartRoutingPage />
                    </AppLayout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/traffic-splits"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <ABTestingPage />
                    </AppLayout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/ab-testing"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <ABTestingPage />
                    </AppLayout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/domains"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <DomainsPage />
                    </AppLayout>
                  </ProtectedRoute>
                }
              />

              <Route
                path="/analytics"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <AnalyticsPage />
                    </AppLayout>
                  </ProtectedRoute>
                }
              />
              <Route
                path="/attribution"
                element={
                  <ProtectedRoute>
                    <AppLayout>
                      <AttributionPage />
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
