import React, { createContext, useContext, useEffect, useState } from 'react';
import { useAuth as useClerkAuth, useUser } from '@clerk/clerk-react';
import { setAuthToken } from '@/api/client';

export interface UserProfile {
  email: string;
  name: string;
  workspaceName: string;
}

export interface AuthContextType {
  isAuthenticated: boolean;
  isLoaded: boolean;
  token: string | null;
  user: UserProfile | null;
  signOut?: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextType>({
  isAuthenticated: false,
  isLoaded: true,
  token: null,
  user: null,
});

export function useAuth() {
  return useContext(AuthContext);
}

export function ClerkAuthProviderWrapper({ children }: { children: React.ReactNode }) {
  const clerk = useClerkAuth();
  const { user: clerkUser } = useUser();
  const [token, setToken] = useState<string | null>(null);

  useEffect(() => {
    let isMounted = true;

    async function syncToken() {
      if (clerk.isSignedIn) {
        try {
          const jwt = await clerk.getToken();
          if (isMounted) {
            setToken(jwt);
            setAuthToken(jwt);
          }
        } catch {
          if (isMounted) {
            setToken(null);
            setAuthToken(null);
          }
        }
      } else {
        if (isMounted) {
          setToken(null);
          setAuthToken(null);
        }
      }
    }

    if (clerk.isLoaded) {
      syncToken();
    }

    return () => {
      isMounted = false;
    };
  }, [clerk.isSignedIn, clerk.isLoaded, clerk.getToken, clerk.orgId]);

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated: Boolean(clerk.isSignedIn),
        isLoaded: Boolean(clerk.isLoaded),
        token,
        user: clerk.isSignedIn && clerkUser ? {
          email: clerkUser.primaryEmailAddress?.emailAddress || 'user@example.com',
          name: clerkUser.fullName || clerkUser.firstName || 'User',
          workspaceName: 'My Workspace',
        } : null,
        signOut: clerk.signOut,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}
