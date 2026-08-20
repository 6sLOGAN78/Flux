import React, { createContext, useContext, useEffect, useState } from 'react';
import { useAuth as useClerkAuth } from '@clerk/clerk-react';
import { setAuthToken } from '@/api/client';

export interface AuthContextType {
  isAuthenticated: boolean;
  isLoaded: boolean;
  token: string | null;
  signOut?: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextType>({
  isAuthenticated: true,
  isLoaded: true,
  token: 'demo_jwt_token',
});

export function useAuth() {
  return useContext(AuthContext);
}

export function StandaloneAuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = useState<string | null>(() => {
    return typeof window !== 'undefined'
      ? localStorage.getItem('flux_demo_token') || 'demo_jwt_token'
      : 'demo_jwt_token';
  });

  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(() => {
    return typeof window !== 'undefined'
      ? localStorage.getItem('flux_demo_auth') !== 'false'
      : true;
  });

  useEffect(() => {
    if (isAuthenticated) {
      setAuthToken(token || 'demo_jwt_token');
    } else {
      setAuthToken(null);
    }
  }, [isAuthenticated, token]);

  const signOut = async () => {
    if (typeof window !== 'undefined') {
      localStorage.setItem('flux_demo_auth', 'false');
    }
    setIsAuthenticated(false);
    setToken(null);
    setAuthToken(null);
  };

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        isLoaded: true,
        token,
        signOut,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

function ClerkAuthSyncInner({ children }: { children: React.ReactNode }) {
  const clerk = useClerkAuth();
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
  }, [clerk.isSignedIn, clerk.isLoaded, clerk.getToken]);

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated: Boolean(clerk.isSignedIn),
        isLoaded: Boolean(clerk.isLoaded),
        token,
        signOut: clerk.signOut,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function ClerkAuthProviderWrapper({ children }: { children: React.ReactNode }) {
  return <ClerkAuthSyncInner>{children}</ClerkAuthSyncInner>;
}
