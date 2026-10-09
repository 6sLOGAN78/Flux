package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/repository"
	"github.com/google/uuid"
)

const maxCursorBytes = 2048

// CursorFilters binds the exact effective collection query, independently of
// page size. Search is normalized before signing; lifecycle is a closed value.
type CursorFilters struct {
	Search string `json:"search"`
	State  string `json:"state"`
}

type linkCursor struct {
	CreatedAt   time.Time `json:"createdAt"`
	Fingerprint string    `json:"fingerprint"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	ID          uuid.UUID `json:"id"`
	Version     int       `json:"version"`
}

func cursorInvalid() error {
	return &errs.HTTPError{Code: "CURSOR_INVALID", Status: http.StatusBadRequest,
		Message: "This page is no longer available. Return to the first page."}
}

func cursorFingerprint(filters CursorFilters) string {
	// A fixed struct avoids map ordering or ambiguous concatenation.
	payload, _ := json.Marshal(filters) // Strings always have a JSON representation.
	hash := sha256.Sum256(payload)
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func encodeCursor(key []byte, workspace uuid.UUID, filters CursorFilters,
	position repository.LinkPosition,
) (string, error) {
	if len(key) != sha256.Size || workspace == uuid.Nil || position.ID == uuid.Nil || position.CreatedAt.IsZero() {
		return "", errors.New("cursor signing unavailable")
	}
	payload, err := json.Marshal(linkCursor{Version: 1, WorkspaceID: workspace,
		Fingerprint: cursorFingerprint(filters), CreatedAt: position.CreatedAt.UTC(), ID: position.ID})
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	signature := hmac.New(sha256.New, key)
	_, _ = signature.Write([]byte(encoded))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(signature.Sum(nil)), nil
}

func decodeCursor(key []byte, token string, workspace uuid.UUID,
	filters CursorFilters,
) (*repository.LinkPosition, error) {
	if len(key) != sha256.Size || len(token) == 0 || len(token) > maxCursorBytes {
		return nil, cursorInvalid()
	}
	encoded, signed, found := strings.Cut(token, ".")
	if !found || strings.Contains(signed, ".") {
		return nil, cursorInvalid()
	}
	decode := base64.RawURLEncoding.Strict()
	signature, err := decode.DecodeString(signed)
	if err != nil || len(signature) != sha256.Size || decode.EncodeToString(signature) != signed {
		return nil, cursorInvalid()
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(encoded))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return nil, cursorInvalid()
	}
	payload, err := decode.DecodeString(encoded)
	if err != nil || decode.EncodeToString(payload) != encoded {
		return nil, cursorInvalid()
	}
	var cursor linkCursor
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&cursor); err != nil {
		return nil, cursorInvalid()
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) || cursor.Version != 1 || cursor.WorkspaceID != workspace ||
		workspace == uuid.Nil || cursor.Fingerprint != cursorFingerprint(filters) || cursor.ID == uuid.Nil ||
		cursor.CreatedAt.IsZero() || cursor.CreatedAt.Year() < 1 || cursor.CreatedAt.Year() > 9999 {
		return nil, cursorInvalid()
	}
	return &repository.LinkPosition{CreatedAt: cursor.CreatedAt, ID: cursor.ID}, nil
}
