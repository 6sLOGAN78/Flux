package repository

import (
	"context"
	"fmt"
	"time"

	"flux/apps/backend/internal/model/analytics"

		"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ClickHouseAnalyticsRepository struct {
	conn driver.Conn
}

func NewClickHouseAnalyticsRepository(conn driver.Conn) *ClickHouseAnalyticsRepository {
	return &ClickHouseAnalyticsRepository{conn: conn}
}

func (r *ClickHouseAnalyticsRepository) GetSummary(ctx context.Context, workspaceID string, from, to time.Time) (*analytics.AnalyticsSummary, error) {
	query := `
		SELECT 
			uniqExact(event_id) as total_clicks,
			uniqExact(ip_hash) as unique_visitors
		FROM analytics_events
		WHERE workspace_id = @workspace_id
		  AND timestamp >= @from 
		  AND timestamp <= @to
	`

	row := r.conn.QueryRow(ctx, query,
		clickhouse.Named("workspace_id", workspaceID),
		clickhouse.Named("from", from),
		clickhouse.Named("to", to),
	)

	var summary analytics.AnalyticsSummary
	if err := row.Scan(&summary.TotalClicks, &summary.UniqueVisitors); err != nil {
		return nil, fmt.Errorf("clickhouse get summary error: %w", err)
	}

	return &summary, nil
}

func (r *ClickHouseAnalyticsRepository) GetTimeseries(ctx context.Context, workspaceID string, from, to time.Time, interval string) (*analytics.TimeseriesResponse, error) {
	intervalFunc := "toStartOfDay(timestamp)"
	if interval == "hour" {
		intervalFunc = "toStartOfHour(timestamp)"
	}

	query := fmt.Sprintf(`
		SELECT 
			%s as ts,
			uniqExact(event_id) as clicks,
			uniqExact(ip_hash) as unique_visitors
		FROM analytics_events
		WHERE workspace_id = @workspace_id
		  AND timestamp >= @from 
		  AND timestamp <= @to
		GROUP BY ts
		ORDER BY ts ASC
	`, intervalFunc)

	rows, err := r.conn.Query(ctx, query,
		clickhouse.Named("workspace_id", workspaceID),
		clickhouse.Named("from", from),
		clickhouse.Named("to", to),
	)
	if err != nil {
		return nil, fmt.Errorf("clickhouse timeseries query error: %w", err)
	}
	defer rows.Close()

	var data []analytics.TimeseriesDataPoint
	for rows.Next() {
		var ts time.Time
		var pt analytics.TimeseriesDataPoint
		if err := rows.Scan(&ts, &pt.Clicks, &pt.UniqueVisitors); err != nil {
			return nil, fmt.Errorf("clickhouse scan timeseries error: %w", err)
		}
		pt.Timestamp = ts.Format(time.RFC3339)
		data = append(data, pt)
	}
	
	if data == nil {
		data = []analytics.TimeseriesDataPoint{} // Return empty array instead of null
	}

	return &analytics.TimeseriesResponse{Data: data}, nil
}

func (r *ClickHouseAnalyticsRepository) GetTopLinks(ctx context.Context, workspaceID string, from, to time.Time, limit int) (*analytics.TopLinksResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	query := `
		SELECT 
			link_id,
			any(short_code) as short_code,
			uniqExact(event_id) as clicks
		FROM analytics_events
		WHERE workspace_id = @workspace_id
		  AND timestamp >= @from 
		  AND timestamp <= @to
		GROUP BY link_id
		ORDER BY clicks DESC
		LIMIT @limit
	`

	rows, err := r.conn.Query(ctx, query,
		clickhouse.Named("workspace_id", workspaceID),
		clickhouse.Named("from", from),
		clickhouse.Named("to", to),
		clickhouse.Named("limit", limit),
	)
	if err != nil {
		return nil, fmt.Errorf("clickhouse toplinks query error: %w", err)
	}
	defer rows.Close()

	var data []analytics.TopLink
	for rows.Next() {
		var link analytics.TopLink
		if err := rows.Scan(&link.LinkID, &link.ShortCode, &link.Clicks); err != nil {
			return nil, fmt.Errorf("clickhouse scan toplinks error: %w", err)
		}
		data = append(data, link)
	}
	
	if data == nil {
		data = []analytics.TopLink{}
	}

	return &analytics.TopLinksResponse{Data: data}, nil
}

func (r *ClickHouseAnalyticsRepository) GetReferrers(ctx context.Context, workspaceID string, from, to time.Time, limit int) (*analytics.ReferrersResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	query := `
		SELECT 
			referrer,
			uniqExact(event_id) as clicks
		FROM analytics_events
		WHERE workspace_id = @workspace_id
		  AND timestamp >= @from 
		  AND timestamp <= @to
		  AND referrer != ''
		GROUP BY referrer
		ORDER BY clicks DESC
		LIMIT @limit
	`

	rows, err := r.conn.Query(ctx, query,
		clickhouse.Named("workspace_id", workspaceID),
		clickhouse.Named("from", from),
		clickhouse.Named("to", to),
		clickhouse.Named("limit", limit),
	)
	if err != nil {
		return nil, fmt.Errorf("clickhouse referrers query error: %w", err)
	}
	defer rows.Close()

	var data []analytics.ReferrerStat
	for rows.Next() {
		var ref analytics.ReferrerStat
		if err := rows.Scan(&ref.Referrer, &ref.Clicks); err != nil {
			return nil, fmt.Errorf("clickhouse scan referrers error: %w", err)
		}
		data = append(data, ref)
	}
	
	if data == nil {
		data = []analytics.ReferrerStat{}
	}

	return &analytics.ReferrersResponse{Data: data}, nil
}

func (r *ClickHouseAnalyticsRepository) GetCampaignPerformance(ctx context.Context, workspaceID string, from, to time.Time, limit int) (*analytics.CampaignPerformanceResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT 
			campaign_id,
			uniqExact(event_id) as clicks,
			uniqExact(ip_hash) as unique_visitors
		FROM analytics_events
		WHERE workspace_id = @workspace_id
		  AND timestamp >= @from 
		  AND timestamp <= @to
		GROUP BY campaign_id
		ORDER BY clicks DESC
		LIMIT @limit
	`

	rows, err := r.conn.Query(ctx, query,
		clickhouse.Named("workspace_id", workspaceID),
		clickhouse.Named("from", from),
		clickhouse.Named("to", to),
		clickhouse.Named("limit", limit),
	)
	if err != nil {
		return nil, fmt.Errorf("clickhouse campaign_performance query error: %w", err)
	}
	defer rows.Close()

	var data []analytics.CampaignPerformance
	for rows.Next() {
		var p analytics.CampaignPerformance
		if err := rows.Scan(&p.CampaignID, &p.Clicks, &p.UniqueVisitors); err != nil {
			return nil, fmt.Errorf("clickhouse scan campaign_performance error: %w", err)
		}
		data = append(data, p)
	}
	
	if data == nil {
		data = []analytics.CampaignPerformance{}
	}

	return &analytics.CampaignPerformanceResponse{Data: data}, nil
}

func (r *ClickHouseAnalyticsRepository) GetUTMPerformance(ctx context.Context, workspaceID string, dimension string, from, to time.Time, limit int) (*analytics.UTMPerformanceResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var dimCol string
	switch dimension {
	case "utm_source":
		dimCol = "utm_source"
	case "utm_medium":
		dimCol = "utm_medium"
	case "utm_campaign":
		dimCol = "utm_campaign"
	case "utm_term":
		dimCol = "utm_term"
	case "utm_content":
		dimCol = "utm_content"
	default:
		dimCol = "utm_source"
	}

	query := fmt.Sprintf(`
		SELECT 
			%s as utm_value,
			uniqExact(event_id) as clicks,
			uniqExact(ip_hash) as unique_visitors
		FROM analytics_events
		WHERE workspace_id = @workspace_id
		  AND timestamp >= @from 
		  AND timestamp <= @to
		  AND %s IS NOT NULL
		  AND %s != ''
		GROUP BY utm_value
		ORDER BY clicks DESC
		LIMIT @limit
	`, dimCol, dimCol, dimCol)

	rows, err := r.conn.Query(ctx, query,
		clickhouse.Named("workspace_id", workspaceID),
		clickhouse.Named("from", from),
		clickhouse.Named("to", to),
		clickhouse.Named("limit", limit),
	)
	if err != nil {
		return nil, fmt.Errorf("clickhouse utm_performance query error: %w", err)
	}
	defer rows.Close()

	var data []analytics.UTMPerformance
	for rows.Next() {
		var p analytics.UTMPerformance
		// utm_value could be read as string pointer? No, WHERE IS NOT NULL filters it out, but just in case, read as *string
		var utmVal *string
		if err := rows.Scan(&utmVal, &p.Clicks, &p.UniqueVisitors); err != nil {
			return nil, fmt.Errorf("clickhouse scan utm_performance error: %w", err)
		}
		if utmVal != nil {
			p.UTMValue = *utmVal
		}
		data = append(data, p)
	}
	
	if data == nil {
		data = []analytics.UTMPerformance{}
	}

	return &analytics.UTMPerformanceResponse{Dimension: dimCol, Data: data}, nil
}

func (r *ClickHouseAnalyticsRepository) GetDomainPerformance(ctx context.Context, workspaceID string, from, to time.Time, limit int) (*analytics.DomainPerformanceResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	query := `
		SELECT 
			if(isNull(hostname) OR hostname = '', 'platform', hostname) as host,
			count(event_id) as clicks,
			uniqExact(ip_hash) as unique_visitors
		FROM analytics_events
		WHERE workspace_id = @workspace_id
		  AND timestamp >= @from
		  AND timestamp <= @to
		GROUP BY host
		ORDER BY clicks DESC
		LIMIT @limit
	`
	
	rows, err := r.conn.Query(ctx, query,
		clickhouse.Named("workspace_id", workspaceID),
		clickhouse.Named("from", from),
		clickhouse.Named("to", to),
		clickhouse.Named("limit", limit),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query domain performance: %w", err)
	}
	defer rows.Close()

	var data []analytics.DomainPerformance
	for rows.Next() {
		var host string
		var clicks, uniqueVisitors uint64
		if err := rows.Scan(&host, &clicks, &uniqueVisitors); err != nil {
			return nil, fmt.Errorf("failed to scan domain performance row: %w", err)
		}
		data = append(data, analytics.DomainPerformance{
			Hostname:       host,
			Clicks:         clicks,
			UniqueVisitors: uniqueVisitors,
		})
	}
	
	return &analytics.DomainPerformanceResponse{Data: data}, nil
}
