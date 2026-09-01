package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"flux/apps/backend/internal/handler"
	"flux/apps/backend/internal/model/analytics"
	"flux/apps/backend/internal/model/user"
	"flux/apps/backend/internal/repository"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWorkspaceRepo struct {
	workspace *user.Workspace
	err       error
}

func (m *mockWorkspaceRepo) GetWorkspaceByTrackingClientID(ctx context.Context, clientID string) (*user.Workspace, error) {
	if m.err != nil {
		return nil, m.err
	}
	if clientID == "valid-client-id" {
		return m.workspace, nil
	}
	return nil, repository.ErrNotFound
}

type mockConversionPublisher struct {
	event *analytics.ConversionEvent
	err   error
}

func (m *mockConversionPublisher) PublishConversion(ctx context.Context, event *analytics.ConversionEvent) error {
	m.event = event
	return m.err
}

func TestTrackingHandler_TrackConversion(t *testing.T) {
	workspaceID := uuid.New()
	validWorkspace := &user.Workspace{
		ID:               workspaceID,
		Name:             "Test Workspace",
		TrackingClientID: uuid.New(),
	}

	tests := []struct {
		name           string
		clientID       string
		payload        analytics.TrackConversionRequest
		mockRepoErr    error
		mockPubErr     error
		expectedStatus int
		verifyEvent    func(t *testing.T, event *analytics.ConversionEvent)
	}{
		{
			name:     "Valid Request",
			clientID: "valid-client-id",
			payload: analytics.TrackConversionRequest{
				ConversionID:   "conv-123",
				ConversionName: "signup",
				Revenue:        10.50,
				Currency:       "USD",
				ClickIDs:       []string{"click-1", "click-2"},
			},
			expectedStatus: http.StatusAccepted,
			verifyEvent: func(t *testing.T, event *analytics.ConversionEvent) {
				assert.NotNil(t, event)
				assert.Equal(t, "conv-123", event.ConversionID)
				assert.Equal(t, workspaceID.String(), event.WorkspaceID)
				assert.Equal(t, "signup", event.ConversionName)
				assert.Equal(t, 10.50, event.Revenue)
				assert.Equal(t, []string{"click-1", "click-2"}, event.ClickIDs)
				assert.False(t, event.Timestamp.IsZero())
			},
		},
		{
			name:     "Invalid Client ID",
			clientID: "invalid-client-id",
			payload: analytics.TrackConversionRequest{
				ConversionID:   "conv-123",
				ConversionName: "signup",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:     "Missing Client ID",
			clientID: "",
			payload: analytics.TrackConversionRequest{
				ConversionID:   "conv-123",
				ConversionName: "signup",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:     "Missing Conversion ID",
			clientID: "valid-client-id",
			payload: analytics.TrackConversionRequest{
				ConversionName: "signup",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Missing Conversion Name",
			clientID: "valid-client-id",
			payload: analytics.TrackConversionRequest{
				ConversionID: "conv-123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Too Many Click IDs",
			clientID: "valid-client-id",
			payload: analytics.TrackConversionRequest{
				ConversionID:   "conv-123",
				ConversionName: "signup",
				ClickIDs:       make([]string, 51),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Redis Publisher Failure",
			clientID: "valid-client-id",
			payload: analytics.TrackConversionRequest{
				ConversionID:   "conv-123",
				ConversionName: "signup",
			},
			mockPubErr:     errors.New("redis down"),
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/track?client_id="+tt.clientID, bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockRepo := &mockWorkspaceRepo{workspace: validWorkspace, err: tt.mockRepoErr}
			mockPub := &mockConversionPublisher{err: tt.mockPubErr}
			h := handler.NewTrackingHandler(mockRepo, mockPub)

			err := h.TrackConversion(c)
			
			// Extract status code from either the direct return or the HTTPError
			statusCode := rec.Code
			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					statusCode = he.Code
				} else {
					t.Fatalf("expected HTTPError, got: %v", err)
				}
			}

			require.Equal(t, tt.expectedStatus, statusCode)

			if tt.verifyEvent != nil {
				tt.verifyEvent(t, mockPub.event)
			}
		})
	}
}
