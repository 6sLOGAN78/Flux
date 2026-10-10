package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/google/uuid"
)

// InvitationEnvelope is private delivery material, never a response or log DTO.
type InvitationEnvelope struct {
	Token        string `json:"token"`
	Email        string `json:"email"`
	Sender       string `json:"sender"`
	PublicOrigin string `json:"publicOrigin"`
}

func invitationAAD(workspace, invitation, delivery uuid.UUID, keyID string) []byte {
	return []byte("flux.invitation.v1\n" + workspace.String() + "\n" +
		invitation.String() + "\n" + delivery.String() + "\n" + keyID)
}

func invitationCipher(key []byte) (cipher.AEAD, error) {
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

func sealInvitation(cfg config.InvitationConfig, scope repository.Scope, invite, delivery uuid.UUID,
	email string,
) (repository.DeliveryIntent, []byte, error) {
	keys, err := cfg.Keys()
	if err != nil {
		return repository.DeliveryIntent{}, nil, err
	}
	aead, err := invitationCipher(keys[cfg.ActiveKeyID])
	if err != nil {
		return repository.DeliveryIntent{}, nil, err
	}
	var raw [32]byte
	if _, err = rand.Read(raw[:]); err != nil {
		return repository.DeliveryIntent{}, nil, errors.New("invitation randomness unavailable")
	}
	token := base64.RawURLEncoding.EncodeToString(raw[:])
	digest := sha256.Sum256([]byte(token))
	plaintext, err := json.Marshal(InvitationEnvelope{
		Token: token, Email: email, Sender: cfg.Sender, PublicOrigin: cfg.PublicOrigin,
	})
	if err != nil {
		return repository.DeliveryIntent{}, nil, errors.New("invitation encryption unavailable")
	}
	defer clear(plaintext)
	// #nosec G407 -- NewGCMWithRandomNonce creates and prepends a random nonce; its API requires nil here.
	ciphertext := aead.Seal(nil, nil, plaintext, invitationAAD(scope.WorkspaceID, invite, delivery, cfg.ActiveKeyID))
	return repository.DeliveryIntent{
		ID: delivery, InvitationID: invite, KeyID: cfg.ActiveKeyID, Ciphertext: ciphertext,
	}, digest[:], nil
}
