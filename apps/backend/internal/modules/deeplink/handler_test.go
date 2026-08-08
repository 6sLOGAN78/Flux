package deeplink_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"flux/apps/backend/internal/modules/deeplink"

	"github.com/labstack/echo/v4"
)

func TestServeAASA_ReturnsValidJSON(t *testing.T) {
	configs := []deeplink.DeepLinkConfig{
		{
			IOSBundleID: "TEAM123.com.acme.app",
		},
	}

	h := deeplink.NewDeepLinkHandler(configs)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/.well-known/apple-app-site-association", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ServeAASA(c)
	if err != nil {
		t.Fatalf("expected ServeAASA to succeed, got error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status HTTP 200 OK, got: %d", rec.Code)
	}

	var res deeplink.AASAResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse AASA response JSON: %v", err)
	}

	if len(res.AppLinks.Details) == 0 {
		t.Fatalf("expected non-empty details array in AASA response")
	}

	if res.AppLinks.Details[0].AppID != "TEAM123.com.acme.app" {
		t.Errorf("expected appID 'TEAM123.com.acme.app', got '%s'", res.AppLinks.Details[0].AppID)
	}
}

func TestServeAssetLinks_ReturnsValidJSON(t *testing.T) {
	configs := []deeplink.DeepLinkConfig{
		{
			AndroidPackageName:        "com.acme.app",
			AndroidSHA256Fingerprint: "FA:C6:11:22:33:44",
		},
	}

	h := deeplink.NewDeepLinkHandler(configs)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/.well-known/assetlinks.json", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ServeAssetLinks(c)
	if err != nil {
		t.Fatalf("expected ServeAssetLinks to succeed, got error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status HTTP 200 OK, got: %d", rec.Code)
	}

	var res []deeplink.AssetLinkResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse AssetLinks response JSON: %v", err)
	}

	if len(res) == 0 {
		t.Fatalf("expected non-empty AssetLinks array")
	}

	if res[0].Target.PackageName != "com.acme.app" {
		t.Errorf("expected package_name 'com.acme.app', got '%s'", res[0].Target.PackageName)
	}
}

func TestBuildFallbackHTML(t *testing.T) {
	h := deeplink.NewDeepLinkHandler(nil)
	html := h.BuildFallbackHTML("acme://open?id=123", "https://apps.apple.com/app/id12345")

	if !containsSubstring(html, "acme://open?id=123") {
		t.Errorf("expected HTML to contain custom scheme URL")
	}
	if !containsSubstring(html, "https://apps.apple.com/app/id12345") {
		t.Errorf("expected HTML to contain fallback store URL")
	}
}

func containsSubstring(s, sub string) bool {
	return deeplink.ContainsSubstring(s, sub)
}
