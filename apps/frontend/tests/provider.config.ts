// Provider evidence is independent from the local signed browser fixtures.
export const requiredProviderRows = [
  "email_password_signup", "password_login", "email_verification", "password_recovery",
  "google_oauth", "github_oauth", "named_workspace_owner", "invitation_onboarding",
  "last_used_workspace", "membership_loss_tabs", "links_empty_state", "token_refresh",
  "token_expiry", "session_revocation", "session_signout", "session_maximum",
  "provider_timeout", "provider_quota_latency", "production_cookie_inspection",
  "provider_key_rotation", "invitation_delivery", "invitation_key_rotation",
] as const;
export const validateAcceptance = (record: unknown, automated = false): boolean => {
  if (!record || typeof record !== "object" || !("rows" in record)) return false;
  const rows = (record as { rows?: unknown }).rows;
  return Array.isArray(rows) && rows.every((row) => row.status === "passed" || (automated && row.status === "human_pending"));
};
