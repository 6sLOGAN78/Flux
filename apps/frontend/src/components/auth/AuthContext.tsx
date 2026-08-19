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
  isAuthenticated: false,
  isLoaded: true,
  token: null,
});

export function useAuth() {
  return useContext(AuthContext);
}

export function ClerkAuthProviderWrapper({ children }: { children: React.ReactNode }) {
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
