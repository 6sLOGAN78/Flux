package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"flux/apps/backend/internal/handler"
	"flux/apps/backend/internal/model/analytics"
	"flux/apps/backend/internal/modules/attribution"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAttributionProvider struct {
	conversions []attribution.Conversion
	err         error
}

func (m *mockAttributionProvider) GetSummary(ctx context.Context, workspaceID string, from, to time.Time) (*analytics.AnalyticsSummary, error) { return nil, nil }
func (m *mockAttributionProvider) GetTimeseries(ctx context.Context, workspaceID string, from, to time.Time, interval string) (*analytics.TimeseriesResponse, error) { return nil, nil }
func (m *mockAttributionProvider) GetTopLinks(ctx context.Context, workspaceID string, from, to time.Time, limit int) (*analytics.TopLinksResponse, error) { return nil, nil }
func (m *mockAttributionProvider) GetReferrers(ctx context.Context, workspaceID string, from, to time.Time, limit int) (*analytics.ReferrersResponse, error) { return nil, nil }
func (m *mockAttributionProvider) GetCampaignPerformance(ctx context.Context, workspaceID string, from, to time.Time, limit int) (*analytics.CampaignPerformanceResponse, error) { return nil, nil }
func (m *mockAttributionProvider) GetUTMPerformance(ctx context.Context, workspaceID string, dimension string, from, to time.Time, limit int) (*analytics.UTMPerformanceResponse, error) { return nil, nil }
func (m *mockAttributionProvider) GetDomainPerformance(ctx context.Context, workspaceID string, from, to time.Time, limit int) (*analytics.DomainPerformanceResponse, error) { return nil, nil }

func (m *mockAttributionProvider) GetConversionsWithTouchpoints(ctx context.Context, workspaceID string, from, to time.Time) ([]attribution.Conversion, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.conversions, nil
}

func TestAnalyticsHandler_GetAttribution(t *testing.T) {
	workspaceID := uuid.New().String()
	campA := uuid.New()
	campB := uuid.New()

	conversions := []attribution.Conversion{
		{
			ID:               uuid.New(),
			VisitorSessionID: uuid.New(),
			Revenue:          100.0,
			ConvertedAt:      time.Now().UTC(),
			Touchpoints: []attribution.Touchpoint{
				{
					ID:           uuid.New(),
					CampaignID:   campA,
					CampaignName: "Campaign A",
					Timestamp:    time.Now().Add(-10 * time.Minute).UTC(),
				},
				{
					ID:           uuid.New(),
					CampaignID:   campB,
					CampaignName: "Campaign B",
					Timestamp:    time.Now().Add(-5 * time.Minute).UTC(),
				},
			},
		},
	}

	mockProvider := &mockAttributionProvider{
		conversions: conversions,
	}
	h := handler.NewAnalyticsHandler(mockProvider, nil)

	e := echo.New()

	tests := []struct {
		name           string
		model          string
		expectedStatus int
		verifyResult   func(t *testing.T, res *attribution.AttributionResult)
	}{
		{
			name:           "First Touch",
			model:          "first_touch",
			expectedStatus: http.StatusOK,
			verifyResult: func(t *testing.T, res *attribution.AttributionResult) {
				assert.Equal(t, attribution.ModelFirstTouch, res.Model)
				assert.Equal(t, 1, res.TotalConversions)
				assert.Equal(t, 100.0, res.TotalAttributedRevenue)
				require.Len(t, res.Campaigns, 1)
				assert.Equal(t, "Campaign A", res.Campaigns[0].CampaignName)
				assert.Equal(t, 100.0, res.Campaigns[0].AttributedRevenue)
			},
		},
		{
			name:           "Last Touch",
			model:          "last_touch",
			expectedStatus: http.StatusOK,
			verifyResult: func(t *testing.T, res *attribution.AttributionResult) {
				assert.Equal(t, attribution.ModelLastTouch, res.Model)
				assert.Equal(t, 1, res.TotalConversions)
				assert.Equal(t, 100.0, res.TotalAttributedRevenue)
				require.Len(t, res.Campaigns, 1)
				assert.Equal(t, "Campaign B", res.Campaigns[0].CampaignName)
				assert.Equal(t, 100.0, res.Campaigns[0].AttributedRevenue)
			},
		},
		{
			name:           "Linear",
			model:          "linear",
			expectedStatus: http.StatusOK,
			verifyResult: func(t *testing.T, res *attribution.AttributionResult) {
				assert.Equal(t, attribution.ModelLinear, res.Model)
				assert.Equal(t, 1, res.TotalConversions)
				assert.Equal(t, 100.0, res.TotalAttributedRevenue)
				require.Len(t, res.Campaigns, 2)
				assert.Equal(t, 50.0, res.Campaigns[0].AttributedRevenue)
				assert.Equal(t, 50.0, res.Campaigns[1].AttributedRevenue)
			},
		},
		{
			name:           "Default (No model specified)",
			model:          "",
			expectedStatus: http.StatusOK,
			verifyResult: func(t *testing.T, res *attribution.AttributionResult) {
				assert.Equal(t, attribution.ModelLinear, res.Model) // Linear is default
				assert.Equal(t, 1, res.TotalConversions)
			},
		},
		{
			name:           "Invalid Model",
			model:          "unknown_model",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/analytics/attribution?model="+tt.model, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			
			// Inject authenticated tenant context
			c.Set("tenant_id", uuid.MustParse(workspaceID))

			err := h.GetAttribution(c)
			statusCode := rec.Code
			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					statusCode = he.Code
				} else {
					t.Fatalf("expected HTTPError, got: %v", err)
				}
			}

			require.Equal(t, tt.expectedStatus, statusCode)

			if tt.verifyResult != nil {
				var res attribution.AttributionResult
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
				tt.verifyResult(t, &res)
			}
		})
	}
}
