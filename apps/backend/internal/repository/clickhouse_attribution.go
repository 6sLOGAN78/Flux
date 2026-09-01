package repository

import (
	"context"
	"fmt"
	"time"

	"flux/apps/backend/internal/modules/attribution"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
)

// StringToUUID safely converts any string to a UUID.
func StringToUUID(s string) uuid.UUID {
	if s == "" {
		return uuid.Nil
	}
	if u, err := uuid.Parse(s); err == nil {
		return u
	}
	return uuid.NewMD5(uuid.NameSpaceOID, []byte(s))
}

func (r *ClickHouseAnalyticsRepository) GetConversionsWithTouchpoints(ctx context.Context, workspaceID string, from, to time.Time) ([]attribution.Conversion, error) {
	query := `
SELECT
    c.conversion_id AS conversion_id,
    c.visitor_id AS visitor_id,
    c.timestamp AS converted_at,
    c.revenue AS revenue,
    e.event_id AS event_id,
    e.timestamp AS touch_time,
    e.link_id AS link_id,
    e.campaign_id AS campaign_id,
    e.referrer AS referrer,
    e.utm_source AS utm_source,
    e.utm_medium AS utm_medium,
    e.utm_campaign AS utm_campaign
FROM (
    SELECT conversion_id, visitor_id, timestamp, revenue, cid
    FROM (
        SELECT conversion_id, visitor_id, timestamp, revenue, click_ids
        FROM conversions
        WHERE workspace_id = @workspace_id
          AND timestamp >= @from AND timestamp <= @to
        ORDER BY timestamp DESC
        LIMIT 1 BY conversion_id
    ) ARRAY JOIN click_ids AS cid
) AS c
INNER JOIN (
    SELECT event_id, timestamp, link_id, campaign_id, referrer, utm_source, utm_medium, utm_campaign
    FROM analytics_events
    WHERE workspace_id = @workspace_id
) AS e
ON e.event_id = c.cid
ORDER BY c.timestamp DESC, e.timestamp ASC
	`

	rows, err := r.conn.Query(ctx, query,
		clickhouse.Named("workspace_id", workspaceID),
		clickhouse.Named("from", from),
		clickhouse.Named("to", to),
	)
	if err != nil {
		return nil, fmt.Errorf("clickhouse attribution query error: %w", err)
	}
	defer rows.Close()

	convMap := make(map[string]*attribution.Conversion)
	var orderedConversions []string

	for rows.Next() {
		var (
			convID       string
			visitorID    string
			convertedAt  time.Time
			revenue      float64
			eventID      string
			touchTime    time.Time
			linkID       string
			campaignID   *string
			referrer     string
			utmSource    *string
			utmMedium    *string
			utmCampaign  *string
		)

		if err := rows.Scan(
			&convID,
			&visitorID,
			&convertedAt,
			&revenue,
			&eventID,
			&touchTime,
			&linkID,
			&campaignID,
			&referrer,
			&utmSource,
			&utmMedium,
			&utmCampaign,
		); err != nil {
			return nil, fmt.Errorf("failed to scan attribution row: %w", err)
		}

		conv, exists := convMap[convID]
		if !exists {
			conv = &attribution.Conversion{
				ID:               StringToUUID(convID),
				VisitorSessionID: StringToUUID(visitorID),
				Revenue:          revenue,
				ConvertedAt:      convertedAt,
				Touchpoints:      []attribution.Touchpoint{},
			}
			convMap[convID] = conv
			orderedConversions = append(orderedConversions, convID)
		}

		cID := uuid.Nil
		cName := "Unassigned"
		if campaignID != nil && *campaignID != "" {
			cID = StringToUUID(*campaignID)
			cName = *campaignID // Best effort for name since we don't join PostgreSQL
		} else {
			// If there's no explicit campaign, we treat the link itself as the attribution target
			cID = StringToUUID(linkID)
			cName = linkID
		}

		tp := attribution.Touchpoint{
			ID:               StringToUUID(eventID),
			VisitorSessionID: StringToUUID(visitorID), // using conversion's visitor_id for simplicity, since it's joined
			LinkID:           StringToUUID(linkID),
			CampaignID:       cID,
			CampaignName:     cName,
			Timestamp:        touchTime,
			ReferrerDomain:   referrer,
		}
		
		if utmSource != nil { tp.UTMSource = *utmSource }
		if utmMedium != nil { tp.UTMMedium = *utmMedium }
		if utmCampaign != nil { tp.UTMCampaign = *utmCampaign }

		conv.Touchpoints = append(conv.Touchpoints, tp)
	}

	result := make([]attribution.Conversion, 0, len(orderedConversions))
	for _, id := range orderedConversions {
		result = append(result, *convMap[id])
	}

	return result, nil
}
