// Presentation hints follow the closed Flux role matrix. Go reauthorizes effects.
export const workspaceCapabilities = (role: string) => {
  const valid = ["owner", "admin", "member", "viewer"].includes(role);
  return {
    read: valid,
    create: valid && role !== "viewer",
    team: role === "owner" || role === "admin",
  };
};
