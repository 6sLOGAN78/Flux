package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
)

// InvitationConfig is API-owned until the independent worker delivery slice.
// EncryptionKeys remains JSON text so the environment loader cannot split dots
// inside key IDs or reinterpret private values as hierarchical configuration.
type InvitationConfig struct {
	ActiveKeyID    string `koanf:"active_key_id"`
	EncryptionKeys string `koanf:"encryption_keys"`
	Sender         string `koanf:"sender"`
	PublicOrigin   string `koanf:"public_origin"`
}

// Keys validates the complete bounded key ring with safe diagnostics only.
func (cfg InvitationConfig) Keys() (map[string][]byte, error) {
	invalid := errors.New("invalid invitation encryption configuration")
	if len(cfg.EncryptionKeys) > 16384 || !validInvitationKeyID(cfg.ActiveKeyID) {
		return nil, invalid
	}
	var encoded map[string]string
	if err := json.Unmarshal([]byte(cfg.EncryptionKeys), &encoded); err != nil || len(encoded) == 0 || len(encoded) > 32 {
		return nil, invalid
	}
	keys := make(map[string][]byte, len(encoded))
	for id, value := range encoded {
		key, err := base64.StdEncoding.Strict().DecodeString(value)
		if !validInvitationKeyID(id) || err != nil || len(key) != 32 || base64.StdEncoding.EncodeToString(key) != value {
			return nil, invalid
		}
		keys[id] = key
	}
	if keys[cfg.ActiveKeyID] == nil {
		return nil, invalid
	}
	return keys, nil
}

func validInvitationKeyID(id string) bool {
	ok, err := regexp.MatchString(`^[A-Za-z0-9_-]{1,64}$`, id)
	return err == nil && ok
}

// Validate fails closed without including email, origin or key material.
func (cfg InvitationConfig) Validate() error {
	if _, err := cfg.Keys(); err != nil {
		return err
	}
	address, err := mail.ParseAddress(cfg.Sender)
	if err != nil || address.Address != cfg.Sender || len(cfg.Sender) > 254 {
		return errors.New("invalid invitation sender")
	}
	origin, err := url.Parse(cfg.PublicOrigin)
	if err != nil || origin.Scheme != "https" || origin.User != nil || origin.RawQuery != "" ||
		origin.Fragment != "" || origin.Path != "" || origin.Host == "" || origin.Opaque != "" ||
		strings.ContainsAny(cfg.PublicOrigin, "\r\n") {
		return errors.New("invalid invitation public origin")
	}
	if _, err = NormalizeLinkHost(origin.Host); err != nil {
		return errors.New("invalid invitation public origin")
	}
	return nil
}
