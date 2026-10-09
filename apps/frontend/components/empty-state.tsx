export function EmptyState({ canCreate }: { canCreate: boolean }) {
  return (
    <section aria-labelledby="empty-links-heading">
      <h2 id="empty-links-heading">No links yet</h2>
      <p>Create a managed-domain link to organize its destination and status.</p>
      {canCreate ? (
        <button type="button" disabled aria-describedby="link-creation-status">
          Create your first link
        </button>
      ) : (
        <p>You have view-only access. Ask an owner or admin to change your role.</p>
      )}
      {canCreate && <p id="link-creation-status">Link creation is not available yet.</p>}
      <p>Link management is available. Redirects and analytics are not available yet.</p>
    </section>
  );
}
