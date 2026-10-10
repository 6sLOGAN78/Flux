//nolint:testpackage // Verify role-owned environment validation and safe private causes.
package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestInvitationConfigFailClosedAndRoleOwnership(t *testing.T) {
	key := base64.StdEncoding.EncodeToString(make([]byte, 32))
	for _, test := range []struct {
		field string
		value string
	}{
		{"active_key_id", ""}, {"active_key_id", "unsafe.private"}, {"active_key_id", "missing"},
		{"encryption_keys", "PRIVATE-INVALID-KEY"}, {"encryption_keys", `{}`}, {"encryption_keys", `{"fixture":"short"}`},
		{"encryption_keys", `{"fixture":"` + key + `","bad.id":"` + key + `"}`},
		{"sender", ""}, {"sender", "Name <a@example.test>"}, {"public_origin", "http://app.flux.test"},
		{"public_origin", "https://private:secret@app.flux.test"}, {"public_origin", "https://app.flux.test/path"},
	} {
		values := configValues()
		values["invitations"].(map[string]any)[test.field] = test.value
		for _, role := range []Role{RoleAPI, RoleWorker} {
			_, err := loadConfigForRole(configProvider{values: values}, role)
			if err == nil {
				t.Fatal("invalid owned invitation setting accepted")
			}
			if test.value != "" && (strings.Contains(fmt.Sprint(err), test.value) ||
				strings.Contains(fmt.Sprint(errors.Unwrap(err)), test.value)) {
				t.Fatal("private setting disclosed")
			}
		}
		for _, role := range []Role{RoleRedirector, RoleMigrator} {
			if _, err := loadConfigForRole(configProvider{values: values}, role); err != nil {
				t.Fatal("unowned invitation setting blocked role")
			}
		}
	}
}
