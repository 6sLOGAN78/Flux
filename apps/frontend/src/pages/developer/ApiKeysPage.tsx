import React, { useState } from 'react';
import { Key, Plus, Code, Check, ShieldCheck, ExternalLink } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { ApiKeyTable, ApiKeyItem } from '@/components/developer/ApiKeyTable';
import { CreateApiKeyModal } from '@/components/developer/CreateApiKeyModal';
import {
  OAuthClientsCard,
  OAuthClientItem,
} from '@/components/developer/OAuthClientsCard';

const INITIAL_KEYS: ApiKeyItem[] = [];

const INITIAL_OAUTH_CLIENTS: OAuthClientItem[] = [];

export function ApiKeysPage() {
  const [keys, setKeys] = useState<ApiKeyItem[]>(INITIAL_KEYS);
  const [oauthClients, setOauthClients] =
    useState<OAuthClientItem[]>(INITIAL_OAUTH_CLIENTS);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [generatedSecret, setGeneratedSecret] = useState<string | undefined>();
  const [notice, setNotice] = useState<string | null>(null);

  const handleCreateKey = (data: { name: string; scopes: string[] }) => {
    const rawSecret = `flx_live_${Math.random().toString(36).substring(2)}${Math.random().toString(36).substring(2)}`;
    const prefix = `${rawSecret.substring(0, 12)}...`;

    const newKey: ApiKeyItem = {
      id: `key_${Date.now()}`,
      name: data.name,
      tokenPrefix: prefix,
      scopes: data.scopes,
      createdAt: new Date().toISOString(),
    };

    setKeys((prev) => [newKey, ...prev]);
    setGeneratedSecret(rawSecret);
  };

  const handleRevokeKey = (id: string) => {
    setKeys((prev) => prev.filter((k) => k.id !== id));
    setNotice('API key successfully revoked');
    setTimeout(() => setNotice(null), 3000);
  };

  const handleRotateSecret = (id: string) => {
    setNotice(`Rotated OAuth client secret for app ${id}`);
    setTimeout(() => setNotice(null), 3000);
  };

  const handleRegisterClient = () => {
    const newClient: OAuthClientItem = {
      id: `oauth_${Date.now()}`,
      name: `OAuth App ${oauthClients.length + 1}`,
      clientId: `flux_client_${Math.random().toString(36).substring(2, 8)}`,
      redirectUris: ['https://localhost:3000/callback'],
      createdAt: new Date().toISOString(),
    };
    setOauthClients((prev) => [...prev, newClient]);
    setNotice('Registered new OAuth 2.0 application');
    setTimeout(() => setNotice(null), 3000);
  };

  const handleCloseModal = () => {
    setIsCreateOpen(false);
    setGeneratedSecret(undefined);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
              Developer API Keys &amp; OAuth 2.0
            </h1>
            <Badge variant="zinc" size="sm">
              REST &amp; OAuth
            </Badge>
          </div>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            Generate scoped tokens for programmatic API access and integrate OAuth client applications.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="primary"
            size="md"
            onClick={() => setIsCreateOpen(true)}
            leftIcon={<Plus className="h-4 w-4" />}
          >
            Generate API Key
          </Button>
        </div>
      </div>

      {notice && (
        <div className="flex items-center gap-2 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-xs font-semibold text-emerald-800 dark:border-emerald-900/50 dark:bg-emerald-950/30 dark:text-emerald-300 animate-in fade-in">
          <Check className="h-4 w-4" />
          <span>{notice}</span>
        </div>
      )}

      {/* API Key Table */}
      <ApiKeyTable keys={keys} onRevokeKey={handleRevokeKey} />

      {/* OAuth 2.0 Clients */}
      <OAuthClientsCard
        clients={oauthClients}
        onRotateSecret={handleRotateSecret}
        onRegisterClient={handleRegisterClient}
      />

      {/* Create Modal */}
      <CreateApiKeyModal
        isOpen={isCreateOpen}
        onClose={handleCloseModal}
        onSubmit={handleCreateKey}
        generatedSecret={generatedSecret}
      />
    </div>
  );
}

export default ApiKeysPage;
