// Package deeplink provides iOS Universal Links (AASA), Android App Links (AssetLinks), and deferred deep linking handlers.
package deeplink

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// DeepLinkConfig holds mobile deep linking metadata for links.
type DeepLinkConfig struct {
	ID                       uuid.UUID `json:"id" db:"id"`
	LinkID                   uuid.UUID `json:"link_id" db:"link_id"`
	IOSAppStoreID            string    `json:"ios_app_store_id" db:"ios_app_store_id"`
	IOSBundleID              string    `json:"ios_bundle_id" db:"ios_bundle_id"` // e.g. "TEAM123.com.acme.app"
	IOSCustomScheme          string    `json:"ios_custom_scheme" db:"ios_custom_scheme"`
	AndroidPackageName       string    `json:"android_package_name" db:"android_package_name"`
	AndroidSHA256Fingerprint string    `json:"android_sha256_fingerprint" db:"android_sha256_fingerprint"`
	AndroidCustomScheme      string    `json:"android_custom_scheme" db:"android_custom_scheme"`
	FallbackURL              string    `json:"fallback_url" db:"fallback_url"`
	CreatedAt                time.Time `json:"created_at" db:"created_at"`
}

// AASADetail represents an Apple App Site Association app details item.
type AASADetail struct {
	AppID string   `json:"appID"`
	Paths []string `json:"paths"`
}

// AASAAppLinks represents the root applinks object in AASA.
type AASAAppLinks struct {
	Apps    []string     `json:"apps"`
	Details []AASADetail `json:"details"`
}

// AASAResponse represents the complete /.well-known/apple-app-site-association JSON document.
type AASAResponse struct {
	AppLinks AASAAppLinks `json:"applinks"`
}

// AssetLinkTarget represents the Android target app specification in Digital Asset Links.
type AssetLinkTarget struct {
	Namespace              string   `json:"namespace"`
	PackageName            string   `json:"package_name"`
	SHA256CertFingerprints []string `json:"sha256_cert_fingerprints"`
}

// AssetLinkResponse represents an item in /.well-known/assetlinks.json array.
type AssetLinkResponse struct {
	Relation []string        `json:"relation"`
	Target   AssetLinkTarget `json:"target"`
}

// DeepLinkHandler handles mobile deep linking endpoints and fallback page rendering.
type DeepLinkHandler struct {
	configs []DeepLinkConfig
}

func NewDeepLinkHandler(configs []DeepLinkConfig) *DeepLinkHandler {
	if configs == nil {
		configs = make([]DeepLinkConfig, 0)
	}
	return &DeepLinkHandler{configs: configs}
}

// ServeAASA handles GET /.well-known/apple-app-site-association and returns valid AASA JSON.
func (h *DeepLinkHandler) ServeAASA(c echo.Context) error {
	details := make([]AASADetail, 0)

	for _, cfg := range h.configs {
		if cfg.IOSBundleID != "" {
			details = append(details, AASADetail{
				AppID: cfg.IOSBundleID,
				Paths: []string{"/*"},
			})
		}
	}

	// Default fallback detail if no specific config exists
	if len(details) == 0 {
		details = append(details, AASADetail{
			AppID: "DEFAULT_TEAM.com.flux.app",
			Paths: []string{"/*"},
		})
	}

	response := AASAResponse{
		AppLinks: AASAAppLinks{
			Apps:    []string{},
			Details: details,
		},
	}

	return c.JSON(http.StatusOK, response)
}

// ServeAssetLinks handles GET /.well-known/assetlinks.json and returns valid Digital Asset Links JSON.
func (h *DeepLinkHandler) ServeAssetLinks(c echo.Context) error {
	responses := make([]AssetLinkResponse, 0)

	for _, cfg := range h.configs {
		if cfg.AndroidPackageName != "" {
			fingerprints := []string{}
			if cfg.AndroidSHA256Fingerprint != "" {
				fingerprints = append(fingerprints, cfg.AndroidSHA256Fingerprint)
			}
			responses = append(responses, AssetLinkResponse{
				Relation: []string{"delegate_permission/common.handle_all_urls"},
				Target: AssetLinkTarget{
					Namespace:              "android_app",
					PackageName:            cfg.AndroidPackageName,
					SHA256CertFingerprints: fingerprints,
				},
			})
		}
	}

	// Default fallback item if no specific config exists
	if len(responses) == 0 {
		responses = append(responses, AssetLinkResponse{
			Relation: []string{"delegate_permission/common.handle_all_urls"},
			Target: AssetLinkTarget{
				Namespace:              "android_app",
				PackageName:            "com.flux.app",
				SHA256CertFingerprints: []string{"00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00"},
			},
		})
	}

	return c.JSON(http.StatusOK, responses)
}

// BuildFallbackHTML constructs a client-side intent wrapper HTML page with JavaScript fallback redirect when app is not installed.
func (h *DeepLinkHandler) BuildFallbackHTML(customScheme, storeURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Opening App...</title>
    <script type="text/javascript">
        window.location.href = "%s";
        setTimeout(function() {
            window.location.href = "%s";
        }, 1500);
    </script>
</head>
<body>
    <p>Redirecting to app... If not redirected, <a href="%s">click here</a>.</p>
</body>
</html>`, customScheme, storeURL, storeURL)
}

// ContainsSubstring checks if string s contains sub.
func ContainsSubstring(s, sub string) bool {
	return strings.Contains(s, sub)
}
