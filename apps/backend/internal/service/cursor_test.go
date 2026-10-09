//nolint:testpackage // Verify private signing and strict decoding boundaries.
package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCursorScopeAndIntegrity(t *testing.T) {
	key := make([]byte, 32) // Test-only signer, never a production fallback.
	workspace := uuid.New()
	filters := CursorFilters{State: "nondeleted", Search: "quoted \" search"}
	position := repository.LinkPosition{CreatedAt: time.Date(2026, 1, 1, 12, 0, 0, 123456000, time.UTC), ID: uuid.New()}
	token, err := encodeCursor(key, workspace, filters, position)
	require.NoError(t, err)
	require.LessOrEqual(t, len(token), maxCursorBytes)
	decoded, err := decodeCursor(key, token, workspace, filters)
	require.NoError(t, err)
	require.Equal(t, position, *decoded)
	for _, invalid := range []string{
		"", "malformed", token + ".extra", token + "=", token[:len(token)-1] + "!", strings.Repeat("x", 2049),
	} {
		_, err = decodeCursor(key, invalid, workspace, filters)
		require.Error(t, err)
		require.Equal(t, cursorInvalid(), err)
	}
	_, err = decodeCursor(key, token, uuid.New(), filters)
	require.Equal(t, cursorInvalid(), err)
	for _, changed := range []CursorFilters{
		{State: "active", Search: filters.Search}, {State: filters.State, Search: "changed"},
	} {
		_, err = decodeCursor(key, token, workspace, changed)
		require.Equal(t, cursorInvalid(), err)
	}
	otherKey := append([]byte(nil), key...)
	otherKey[0] = 1
	_, err = decodeCursor(otherKey, token, workspace, filters)
	require.Equal(t, cursorInvalid(), err)
}

func TestCursorAuthenticatedPayloadValidation(t *testing.T) {
	key := make([]byte, 32)
	workspace := uuid.New()
	filters := CursorFilters{State: "nondeleted"}
	valid := linkCursor{Version: 1, WorkspaceID: workspace, Fingerprint: cursorFingerprint(filters),
		CreatedAt: time.Now().UTC(), ID: uuid.New()}
	for _, mutation := range []func(*linkCursor){
		func(c *linkCursor) { c.Version = 2 },
		func(c *linkCursor) { c.ID = uuid.Nil },
		func(c *linkCursor) { c.CreatedAt = time.Time{} },
		func(c *linkCursor) { c.WorkspaceID = uuid.Nil },
		func(c *linkCursor) { c.Fingerprint = "different" },
	} {
		cursor := valid
		mutation(&cursor)
		payload, err := json.Marshal(cursor)
		require.NoError(t, err)
		_, err = decodeCursor(key, signTestCursor(key, payload), workspace, filters)
		require.Equal(t, cursorInvalid(), err)
	}
	for _, payload := range []string{`{"createdAt":"not-a-time"}`, `{"id":"not-a-uuid"}`, `{"unknown":true}`, `{} {}`} {
		_, err := decodeCursor(key, signTestCursor(key, []byte(payload)), workspace, filters)
		require.Equal(t, cursorInvalid(), err)
	}
	_, err := encodeCursor(key[:31], workspace, filters, repository.LinkPosition{CreatedAt: valid.CreatedAt, ID: valid.ID})
	require.Error(t, err)
}

func signTestCursor(key, payload []byte) string {
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(encoded))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
