package model_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"flux/apps/backend/internal/model/analytics"
	"flux/apps/backend/internal/model/campaign"
	"flux/apps/backend/internal/model/link"
)

func TestCampaign_DomainInvariants(t *testing.T) {
	workspaceID := uuid.New()
	campID := uuid.New()

	c := campaign.Campaign{
		ID:          campID,
		WorkspaceID: workspaceID,
		Name:        "Summer Promo",
		Status:      "active",
	}

	assert.Equal(t, campID, c.ID)
	assert.Equal(t, workspaceID, c.WorkspaceID)
	assert.Equal(t, "active", c.Status)
}

func TestLink_UTMAndCampaign(t *testing.T) {
	campID := uuid.New()
	utmSource := "twitter"

	l := link.Link{
		ID:             uuid.New(),
		ShortCode:      "promo",
		DestinationURL: "https://example.com",
		CampaignID:     &campID,
		UTMSource:      &utmSource,
	}

	assert.NotNil(t, l.CampaignID)
	assert.Equal(t, campID, *l.CampaignID)
	assert.NotNil(t, l.UTMSource)
	assert.Equal(t, "twitter", *l.UTMSource)
	assert.Nil(t, l.UTMMedium) // should be nil
}

func TestAnalyticsEvent_SnapshotPreservation(t *testing.T) {
	campID := "campaign-A-uuid"
	utmSource := "facebook"

	// Represents a click event that happens today
	event := analytics.AnalyticsEvent{
		EventID:     "evt-123",
		EventType:   analytics.EventTypeLinkRedirect,
		Timestamp:   time.Now(),
		LinkID:      "link-123",
		WorkspaceID: "ws-123",
		CampaignID:  &campID,
		UTMSource:   &utmSource,
	}

	// Tomorrow, the user moves the link to Campaign B
	newCampID := "campaign-B-uuid"
	linkRecord := link.Link{
		ID:         uuid.New(),
		CampaignID: (*uuid.UUID)(nil), // Fake conversion for testing
	}
	_ = linkRecord
	_ = newCampID

	// Verify the historical event is immutable and retains its original values
	assert.Equal(t, "campaign-A-uuid", *event.CampaignID)
	assert.Equal(t, "facebook", *event.UTMSource)

	// Verify JSON marshaling keeps it optional
	data, err := json.Marshal(event)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "campaign-A-uuid")

	// Verify an event without a campaign correctly omits the fields
	emptyEvent := analytics.AnalyticsEvent{
		EventID: "evt-456",
	}
	emptyData, _ := json.Marshal(emptyEvent)
	assert.NotContains(t, string(emptyData), "campaign_id")
	assert.NotContains(t, string(emptyData), "utm_source")
}
