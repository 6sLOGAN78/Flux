import React, { createContext, useContext, useEffect, useState } from 'react';
import { useAuth as useClerkAuth } from '@clerk/clerk-react';
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
  signIn?: (email?: string, password?: string) => Promise<void>;
  signOut?: () => Promise<void>;
}

const DEFAULT_USER: UserProfile = {
  email: 'dev@localhost',
  name: 'Developer',
  workspaceName: 'Local Dev Workspace',
};

export const AuthContext = createContext<AuthContextType>({
  isAuthenticated: false,
  isLoaded: true,
  token: null,
  user: null,
});

export function useAuth() {
  return useContext(AuthContext);
}

export function StandaloneAuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = useState<string | null>(() => {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('flux_demo_auth') === 'true'
      ? localStorage.getItem('flux_demo_token') || 'demo_jwt_token'
      : null;
  });

  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(() => {
    if (typeof window === 'undefined') return true;
    return localStorage.getItem('flux_demo_auth') === 'true';
  });

  const [user, setUser] = useState<UserProfile | null>(() => {
    if (typeof window === 'undefined') return null;
    if (localStorage.getItem('flux_demo_auth') !== 'true') return null;
    const storedUser = localStorage.getItem('flux_demo_user');
    if (storedUser) {
      try {
        return JSON.parse(storedUser);
      } catch {
        return DEFAULT_USER;
      }
    }
    return DEFAULT_USER;
  });

  useEffect(() => {
    if (isAuthenticated) {
      setAuthToken(token || 'demo_jwt_token');
    } else {
      setAuthToken(null);
    }
  }, [isAuthenticated, token]);

  const signIn = async (email?: string, _password?: string) => {
    const demoToken = 'demo_jwt_token';
    const userEmail = email && email.trim() ? email.trim() : 'dev@localhost';
    const namePart = userEmail.split('@')[0] || 'Developer';
    const capitalizedName = namePart.charAt(0).toUpperCase() + namePart.slice(1);
    
    let domainPart = userEmail.split('@')[1] || 'localhost';
    if (domainPart.includes('.')) {
      domainPart = domainPart.split('.')[0];
    }
    const workspaceName = domainPart.charAt(0).toUpperCase() + domainPart.slice(1) + ' Workspace';

    const profile: UserProfile = {
      email: userEmail,
      name: capitalizedName,
      workspaceName: workspaceName,
    };

    if (typeof window !== 'undefined') {
      localStorage.setItem('flux_demo_auth', 'true');
      localStorage.setItem('flux_demo_token', demoToken);
      localStorage.setItem('flux_demo_user', JSON.stringify(profile));
    }

    setToken(demoToken);
    setUser(profile);
    setIsAuthenticated(true);
    setAuthToken(demoToken);
  };

  const signOut = async () => {
    if (typeof window !== 'undefined') {
      localStorage.setItem('flux_demo_auth', 'false');
      localStorage.removeItem('flux_demo_token');
      localStorage.removeItem('flux_demo_user');
    }
    setIsAuthenticated(false);
    setUser(null);
    setToken(null);
    setAuthToken(null);
  };

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        isLoaded: true,
        token,
        user,
        signIn,
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
        user: clerk.isSignedIn ? DEFAULT_USER : null,
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
