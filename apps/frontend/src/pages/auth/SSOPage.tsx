import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '@/components/auth/AuthContext';

export function SSOPage() {
  const [domain, setDomain] = useState('');
  const [isRedirecting, setIsRedirecting] = useState(false);
  const { signIn } = useAuth();
  const navigate = useNavigate();

  const handleSSO = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsRedirecting(true);
    try {
      if (signIn) {
        await signIn();
      }
      navigate('/dashboard');
    } finally {
      setIsRedirecting(false);
    }
  };

  return (
    <div className="flex min-h-[100dvh] items-center justify-center bg-zinc-50 p-4 dark:bg-zinc-950">
      <div className="w-full max-w-sm rounded-xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-900">
        <div className="text-center">
          <Link to="/" className="inline-flex items-center gap-2 font-semibold text-zinc-900 dark:text-zinc-50">
            <span className="flex h-6 w-6 items-center justify-center rounded-md bg-zinc-900 text-xs font-bold text-white dark:bg-zinc-100 dark:text-zinc-900">
              F
            </span>
            <span>Flux</span>
          </Link>
          <h2 className="mt-4 text-xl font-semibold tracking-tight text-zinc-900 dark:text-zinc-50">
            Enterprise Single Sign-On
          </h2>
          <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
            Sign in via your organization's SAML 2.0 or OIDC Identity Provider
          </p>
        </div>

        <form onSubmit={handleSSO} className="mt-6 space-y-4">
          <div>
            <label className="block text-xs font-medium text-zinc-700 dark:text-zinc-300">
              Company Domain
            </label>
            <input
              type="text"
              required
              value={domain}
              onChange={(e) => setDomain(e.target.value)}
              placeholder="acme.com"
              className="mt-1.5 w-full rounded-lg border border-zinc-200 bg-transparent px-3 py-2 text-sm text-zinc-900 placeholder:text-zinc-400 focus:border-zinc-900 focus:outline-hidden focus:ring-1 focus:ring-zinc-900 dark:border-zinc-800 dark:text-zinc-50 dark:focus:border-zinc-100 dark:focus:ring-zinc-100"
            />
          </div>

          <button
            type="submit"
            disabled={isRedirecting}
            className="w-full rounded-lg bg-zinc-900 py-2 text-sm font-medium text-white transition-colors hover:bg-zinc-800 disabled:opacity-50 dark:bg-zinc-100 dark:text-zinc-900 dark:hover:bg-zinc-200"
          >
            {isRedirecting ? 'Redirecting to IdP...' : 'Continue with SSO'}
          </button>
        </form>

        <p className="mt-6 text-center text-xs text-zinc-500 dark:text-zinc-400">
          <Link to="/sign-in" className="font-medium text-zinc-900 underline underline-offset-4 dark:text-zinc-50">
            ← Back to email sign in
          </Link>
        </p>
      </div>
    </div>
  );
}

export default SSOPage;
