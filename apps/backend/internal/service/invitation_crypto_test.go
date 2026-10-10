//nolint:testpackage // Tests private crypto primitives without exporting decryption to the API.
package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/google/uuid"
)

func TestInvitationEnvelopeIntegrity(t *testing.T) {
	cfg := config.InvitationConfig{ActiveKeyID: "fixture",
		EncryptionKeys: `{"fixture":"` + base64.StdEncoding.EncodeToString(make([]byte, 32)) + `"}`,
		Sender:         "invites@example.test", PublicOrigin: "https://app.flux.test"}
	scope := repository.Scope{WorkspaceID: uuid.New(), ActorID: uuid.New()}
	invite, delivery := uuid.New(), uuid.New()
	intent, digest, err := sealInvitation(cfg, scope, invite, delivery, "future+tag@example.test")
	if err != nil {
		t.Fatal("encryption failed")
	}
	keys, err := cfg.Keys()
	if err != nil {
		t.Fatal("key ring failed")
	}
	aead, err := invitationCipher(keys[cfg.ActiveKeyID])
	if err != nil {
		t.Fatal("cipher failed")
	}
	plaintext, err := aead.Open(nil, nil, intent.Ciphertext,
		invitationAAD(scope.WorkspaceID, invite, delivery, intent.KeyID))
	if err != nil {
		t.Fatal("authenticated decryption failed")
	}
	defer clear(plaintext)
	var envelope InvitationEnvelope
	if err = json.Unmarshal(plaintext, &envelope); err != nil {
		t.Fatal("envelope invalid")
	}
	raw, err := base64.RawURLEncoding.DecodeString(envelope.Token)
	if err != nil || len(raw) != 32 || envelope.Email != "future+tag@example.test" ||
		envelope.Sender != cfg.Sender || envelope.PublicOrigin != cfg.PublicOrigin {
		t.Fatal("private envelope invalid")
	}
	hash := sha256.Sum256([]byte(envelope.Token))
	if string(digest) != string(hash[:]) || strings.Contains(string(intent.Ciphertext), envelope.Token) {
		t.Fatal("acceptance digest or encryption invalid")
	}
	second, _, err := sealInvitation(cfg, scope, invite, delivery, envelope.Email)
	if err != nil || string(second.Ciphertext) == string(intent.Ciphertext) {
		t.Fatal("random token and nonce repeated")
	}
	for _, aad := range [][]byte{
		invitationAAD(uuid.New(), invite, delivery, intent.KeyID),
		invitationAAD(scope.WorkspaceID, uuid.New(), delivery, intent.KeyID),
		invitationAAD(scope.WorkspaceID, invite, uuid.New(), intent.KeyID),
		invitationAAD(scope.WorkspaceID, invite, delivery, "other"),
	} {
		if _, err = aead.Open(nil, nil, intent.Ciphertext, aad); err == nil {
			t.Fatal("foreign envelope authenticated")
		}
	}
	tampered := append([]byte(nil), intent.Ciphertext...)
	tampered[len(tampered)-1] ^= 1
	if _, err = aead.Open(nil, nil, tampered,

		invitationAAD(scope.WorkspaceID, invite, delivery, intent.KeyID)); err == nil {
		t.Fatal("tampered envelope authenticated")
	}
}

func TestInvitationRotationAndFailClosedDecryption(t *testing.T) {
	oldKey, newKey := make([]byte, 32), bytes.Repeat([]byte{1}, 32)
	cfg := config.InvitationConfig{ActiveKeyID: "old",
		EncryptionKeys: `{"old":"` + base64.StdEncoding.EncodeToString(oldKey) + `","new":"` + base64.StdEncoding.EncodeToString(newKey) + `"}`,
		Sender:         "invites@example.test", PublicOrigin: "https://app.flux.test"}
	scope := repository.Scope{WorkspaceID: uuid.New(), ActorID: uuid.New()}
	invite, delivery := uuid.New(), uuid.New()
	intent, _, err := sealInvitation(cfg, scope, invite, delivery, "rotation@example.test")
	if err != nil {
		t.Fatal("seal failed")
	}
	cfg.ActiveKeyID = "new"
	rotated, _, err := sealInvitation(cfg, scope, invite, uuid.New(), "rotation@example.test")
	if err != nil || rotated.KeyID != "new" || intent.KeyID != "old" {
		t.Fatal("active rotation failed")
	}
	keys, err := cfg.Keys()
	if err != nil {
		t.Fatal("retained ring failed")
	}
	aead, err := invitationCipher(keys[intent.KeyID])
	if err != nil {
		t.Fatal("old key unavailable")
	}
	aad := invitationAAD(scope.WorkspaceID, invite, delivery, intent.KeyID)
	plaintext, err := aead.Open(nil, nil, intent.Ciphertext, aad)
	if err != nil || !bytes.Contains(plaintext, []byte("rotation@example.test")) {
		t.Fatal("retained key cannot decrypt pending intent")
	}
	clear(plaintext)
	for length := range 28 {
		out, openErr := aead.Open(nil, nil, intent.Ciphertext[:length], aad)
		if openErr == nil || len(out) != 0 {
			t.Fatal("truncated envelope released plaintext")
		}
	}
	wrong, err := invitationCipher(keys[cfg.ActiveKeyID])
	if err != nil {
		t.Fatal("new cipher unavailable")
	}
	if out, openErr := wrong.Open(nil, nil, intent.Ciphertext, aad); openErr == nil || len(out) != 0 {
		t.Fatal("wrong key released plaintext")
	}
	for _, key := range [][]byte{keys["missing"], make([]byte, 16), make([]byte, 31), make([]byte, 33)} {
		if aead, cipherErr := invitationCipher(key); cipherErr == nil || aead != nil {
			t.Fatal("missing or malformed key accepted")
		}
	}
	cfg.EncryptionKeys = `{"new":"` + base64.StdEncoding.EncodeToString(newKey) + `"}`
	keys, err = cfg.Keys()
	if err != nil {
		t.Fatal("new ring failed")
	}
	if aead, cipherErr := invitationCipher(keys[intent.KeyID]); cipherErr == nil || aead != nil {
		t.Fatal("retired key silently substituted")
	}
}

func TestInvitationInputAndRolePolicy(t *testing.T) {
	for _, email := range []string{"", "a@b", "a@localhost", "a..b@example.test", ".a@example.test",
		strings.Repeat("a", 255) + "@example.test", "name <a@example.test>", "a@example.test\r\nprivate"} {
		if _, err := normalizeInvitationEmail(email); err == nil {
			t.Fatal("invalid email accepted")
		}
	}
	if email, err := normalizeInvitationEmail("  A.B+Tag@Example.Test  "); err != nil || email != "a.b+tag@example.test" {
		t.Fatal("email alias was rewritten")
	}
	for _, actor := range []string{"owner", "admin", "member", "viewer", "forged"} {
		for _, role := range []string{"owner", "admin", "member", "viewer", "forged"} {
			expected := (actor == "owner" && (role == "admin" || role == "member" || role == "viewer")) ||
				(actor == "admin" && (role == "member" || role == "viewer"))
			if AllowsInvitation(actor, role) != expected {
				t.Fatal("role matrix violated")
			}
		}
	}
	var unavailable *InvitationService
	if _, err := unavailable.Create(context.Background(), repository.Scope{},
		"a@example.test", "member", "valid-invite-key-01"); err == nil {
		t.Fatal("missing service accepted")
	}
}
