package tenant_test

import (
	"testing"

	"flux/apps/backend/internal/modules/tenant"
)

func TestRoleHierarchyAndPermissions(t *testing.T) {
	// Owner permissions
	if !tenant.HasPermission(tenant.RoleOwner, tenant.PermissionManageBilling) {
		t.Error("expected owner to have manage_billing permission")
	}
	if !tenant.HasPermission(tenant.RoleOwner, tenant.PermissionDeleteWorkspace) {
		t.Error("expected owner to have delete_workspace permission")
	}

	// Admin permissions
	if !tenant.HasPermission(tenant.RoleAdmin, tenant.PermissionAdmin) {
		t.Error("expected admin to have admin permission")
	}
	if tenant.HasPermission(tenant.RoleAdmin, tenant.PermissionManageBilling) {
		t.Error("expected admin NOT to have manage_billing permission")
	}

	// Editor permissions
	if !tenant.HasPermission(tenant.RoleEditor, tenant.PermissionWrite) {
		t.Error("expected editor to have write permission")
	}
	if tenant.HasPermission(tenant.RoleEditor, tenant.PermissionAdmin) {
		t.Error("expected editor NOT to have admin permission")
	}

	// Viewer permissions
	if !tenant.HasPermission(tenant.RoleViewer, tenant.PermissionRead) {
		t.Error("expected viewer to have read permission")
	}
	if tenant.HasPermission(tenant.RoleViewer, tenant.PermissionWrite) {
		t.Error("expected viewer NOT to have write permission")
	}
}

func TestCustomPermissionOverrides(t *testing.T) {
	customPerms := map[string]bool{
		"manage_custom_domains": true,
	}

	// Viewer normally doesn't have custom domain permissions
	if tenant.HasPermissionWithCustom(tenant.RoleViewer, "manage_custom_domains", customPerms) != true {
		t.Error("expected custom permission override to grant manage_custom_domains to viewer")
	}
}

func TestValidateRole(t *testing.T) {
	validRoles := []string{"owner", "admin", "editor", "viewer"}
	for _, r := range validRoles {
		if !tenant.IsValidRole(r) {
			t.Errorf("expected role %q to be valid", r)
		}
	}

	if tenant.IsValidRole("super_god_mode") {
		t.Error("expected invalid role to fail validation")
	}
}
