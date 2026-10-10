// Package invitationcrypto shares the versioned private envelope format across roles.
package invitationcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"errors"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/google/uuid"
)

// Envelope is decrypted delivery material, never a queue, response or log DTO.
type Envelope struct {
	Message      *Message `json:"message,omitempty"`
	Token        string   `json:"token"`
	Email        string   `json:"email"`
	Sender       string   `json:"sender"`
	PublicOrigin string   `json:"publicOrigin"`
}

// Message freezes every provider request field before any external effect.
type Message struct {
	To             string `json:"to"`
	From           string `json:"from"`
	Subject        string `json:"subject"`
	HTML           string `json:"html"`
	IdempotencyKey string `json:"idempotencyKey"`
}

// Seal preserves the original AES-GCM/AAD format while adding encrypted snapshots.
//
//nolint:lll // Keep scoped SQL fences, bounded validation and exact fixture assertions together.
func Seal(cfg config.InvitationConfig, workspace, invitation, delivery uuid.UUID, keyID string, envelope Envelope) ([]byte, error) {
	keys, err := cfg.Keys()
	if err != nil {
		return nil, err
	}
	aead, err := Cipher(keys[keyID])
	if err != nil {
		return nil, err
	}
	plaintext, err := json.Marshal(envelope)
	if err != nil {
		return nil, errors.New("invitation envelope unavailable")
	}
	defer clear(plaintext)
	// #nosec G407 -- Standard NewGCMWithRandomNonce generates and prepends the nonce.
	return aead.Seal(nil, nil, plaintext, AAD(workspace, invitation, delivery, keyID)), nil
}

// AAD binds every tenant, invitation, intent and key identifier.
//
//nolint:lll // Keep scoped SQL fences, bounded validation and exact fixture assertions together.
func AAD(workspace, invitation, delivery uuid.UUID, keyID string) []byte {
	return []byte("flux.invitation.v1\n" + workspace.String() + "\n" + invitation.String() + "\n" + delivery.String() + "\n" + keyID)
}

// Cipher uses the standard random-nonce AES-256-GCM format.
func Cipher(key []byte) (cipher.AEAD, error) {
	const keyBytes = 32
	if len(key) != keyBytes {
		return nil, errors.New("invalid invitation encryption key")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.New("invitation encryption unavailable")
	}
	return cipher.NewGCMWithRandomNonce(block)
}

// Open fails closed and returns no plaintext on authentication or format failure.
//
//nolint:lll // Keep scoped SQL fences, bounded validation and exact fixture assertions together.
func Open(cfg config.InvitationConfig, workspace, invitation, delivery uuid.UUID, keyID string, ciphertext []byte) (Envelope, error) {
	invalid := errors.New("invitation envelope unavailable")
	keys, err := cfg.Keys()
	if err != nil {
		return Envelope{}, invalid
	}
	aead, err := Cipher(keys[keyID])
	if err != nil {
		return Envelope{}, invalid
	}
	plaintext, err := aead.Open(nil, nil, ciphertext, AAD(workspace, invitation, delivery, keyID))
	if err != nil {
		return Envelope{}, invalid
	}
	defer clear(plaintext)
	var envelope Envelope
	if json.Unmarshal(plaintext, &envelope) != nil {
		return Envelope{}, invalid
	}
	return envelope, nil
}
