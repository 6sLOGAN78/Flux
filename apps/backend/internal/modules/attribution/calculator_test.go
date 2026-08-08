package attribution

import (
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCalculateAttribution_FirstTouch(t *testing.T) {
	calc := NewCalculator()

	campA := uuid.New()
	campB := uuid.New()
	sessionID := uuid.New()
	now := time.Now()

	conversions := []Conversion{
		{
			ID:               uuid.New(),
			VisitorSessionID: sessionID,
			Revenue:          100.0,
			ConvertedAt:      now,
			Touchpoints: []Touchpoint{
				{ID: uuid.New(), VisitorSessionID: sessionID, CampaignID: campA, CampaignName: "Campaign A", Timestamp: now.Add(-2 * time.Hour)},
				{ID: uuid.New(), VisitorSessionID: sessionID, CampaignID: campB, CampaignName: "Campaign B", Timestamp: now.Add(-1 * time.Hour)},
			},
		},
	}

	res, err := calc.Calculate(conversions, ModelFirstTouch, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Model != ModelFirstTouch {
		t.Errorf("expected model %s, got %s", ModelFirstTouch, res.Model)
	}
	if res.TotalConversions != 1 {
		t.Errorf("expected 1 total conversion, got %d", res.TotalConversions)
	}
	if res.TotalAttributedRevenue != 100.0 {
		t.Errorf("expected 100.0 attributed revenue, got %f", res.TotalAttributedRevenue)
	}

	if len(res.Campaigns) != 1 {
		t.Fatalf("expected 1 campaign in result, got %d", len(res.Campaigns))
	}

	if res.Campaigns[0].CampaignID != campA {
		t.Errorf("expected First-Touch credit to go to Campaign A (%s), got %s", campA, res.Campaigns[0].CampaignID)
	}
	if res.Campaigns[0].AttributedRevenue != 100.0 {
		t.Errorf("expected 100.0 attributed revenue for Campaign A, got %f", res.Campaigns[0].AttributedRevenue)
	}
}

func TestCalculateAttribution_LastTouch(t *testing.T) {
	calc := NewCalculator()

	campA := uuid.New()
	campB := uuid.New()
	sessionID := uuid.New()
	now := time.Now()

	conversions := []Conversion{
		{
			ID:               uuid.New(),
			VisitorSessionID: sessionID,
			Revenue:          200.0,
			ConvertedAt:      now,
			Touchpoints: []Touchpoint{
				{ID: uuid.New(), VisitorSessionID: sessionID, CampaignID: campA, CampaignName: "Campaign A", Timestamp: now.Add(-2 * time.Hour)},
				{ID: uuid.New(), VisitorSessionID: sessionID, CampaignID: campB, CampaignName: "Campaign B", Timestamp: now.Add(-1 * time.Hour)},
			},
		},
	}

	res, err := calc.Calculate(conversions, ModelLastTouch, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Campaigns) != 1 {
		t.Fatalf("expected 1 campaign in result, got %d", len(res.Campaigns))
	}
	if res.Campaigns[0].CampaignID != campB {
		t.Errorf("expected Last-Touch credit to go to Campaign B (%s), got %s", campB, res.Campaigns[0].CampaignID)
	}
	if res.Campaigns[0].AttributedRevenue != 200.0 {
		t.Errorf("expected 200.0 attributed revenue for Campaign B, got %f", res.Campaigns[0].AttributedRevenue)
	}
}

func TestCalculateAttribution_Linear(t *testing.T) {
	calc := NewCalculator()

	campA := uuid.New()
	campB := uuid.New()
	sessionID := uuid.New()
	now := time.Now()

	conversions := []Conversion{
		{
			ID:               uuid.New(),
			VisitorSessionID: sessionID,
			Revenue:          100.0,
			ConvertedAt:      now,
			Touchpoints: []Touchpoint{
				{ID: uuid.New(), VisitorSessionID: sessionID, CampaignID: campA, CampaignName: "Campaign A", Timestamp: now.Add(-2 * time.Hour)},
				{ID: uuid.New(), VisitorSessionID: sessionID, CampaignID: campB, CampaignName: "Campaign B", Timestamp: now.Add(-1 * time.Hour)},
			},
		},
	}

	res, err := calc.Calculate(conversions, ModelLinear, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Campaigns) != 2 {
		t.Fatalf("expected 2 campaigns in result, got %d", len(res.Campaigns))
	}

	for _, c := range res.Campaigns {
		if math.Abs(c.AttributedConversions-0.5) > 0.001 {
			t.Errorf("expected 0.5 attributed conversions for %s, got %f", c.CampaignName, c.AttributedConversions)
		}
		if math.Abs(c.AttributedRevenue-50.0) > 0.001 {
			t.Errorf("expected 50.0 attributed revenue for %s, got %f", c.CampaignName, c.AttributedRevenue)
		}
	}
}

func TestCalculateAttribution_TimeDecay(t *testing.T) {
	calc := NewCalculator()

	campA := uuid.New()
	campB := uuid.New()
	sessionID := uuid.New()
	now := time.Now()

	halfLife := 7 * 24 * time.Hour

	conversions := []Conversion{
		{
			ID:               uuid.New(),
			VisitorSessionID: sessionID,
			Revenue:          100.0,
			ConvertedAt:      now,
			Touchpoints: []Touchpoint{
				{ID: uuid.New(), VisitorSessionID: sessionID, CampaignID: campA, CampaignName: "Campaign A", Timestamp: now.Add(-7 * 24 * time.Hour)},
				{ID: uuid.New(), VisitorSessionID: sessionID, CampaignID: campB, CampaignName: "Campaign B", Timestamp: now},
			},
		},
	}

	res, err := calc.Calculate(conversions, ModelTimeDecay, halfLife)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Campaigns) != 2 {
		t.Fatalf("expected 2 campaigns in result, got %d", len(res.Campaigns))
	}

	var campARev, campBRev float64
	for _, c := range res.Campaigns {
		if c.CampaignID == campA {
			campARev = c.AttributedRevenue
		} else if c.CampaignID == campB {
			campBRev = c.AttributedRevenue
		}
	}

	if math.Abs(campARev-33.333) > 0.1 {
		t.Errorf("expected Campaign A revenue ~33.33, got %f", campARev)
	}
	if math.Abs(campBRev-66.667) > 0.1 {
		t.Errorf("expected Campaign B revenue ~66.67, got %f", campBRev)
	}
}

func TestCalculateAttribution_PositionBased(t *testing.T) {
	calc := NewCalculator()

	campFirst := uuid.New()
	campMid1 := uuid.New()
	campMid2 := uuid.New()
	campLast := uuid.New()
	sessionID := uuid.New()
	now := time.Now()

	conversions := []Conversion{
		{
			ID:               uuid.New(),
			VisitorSessionID: sessionID,
			Revenue:          100.0,
			ConvertedAt:      now,
			Touchpoints: []Touchpoint{
				{ID: uuid.New(), VisitorSessionID: sessionID, CampaignID: campFirst, CampaignName: "First", Timestamp: now.Add(-4 * time.Hour)},
				{ID: uuid.New(), VisitorSessionID: sessionID, CampaignID: campMid1, CampaignName: "Mid 1", Timestamp: now.Add(-3 * time.Hour)},
				{ID: uuid.New(), VisitorSessionID: sessionID, CampaignID: campMid2, CampaignName: "Mid 2", Timestamp: now.Add(-2 * time.Hour)},
				{ID: uuid.New(), VisitorSessionID: sessionID, CampaignID: campLast, CampaignName: "Last", Timestamp: now.Add(-1 * time.Hour)},
			},
		},
	}

	res, err := calc.Calculate(conversions, ModelPositionBased, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Campaigns) != 4 {
		t.Fatalf("expected 4 campaigns in result, got %d", len(res.Campaigns))
	}

	revMap := make(map[uuid.UUID]float64)
	for _, c := range res.Campaigns {
		revMap[c.CampaignID] = c.AttributedRevenue
	}

	if math.Abs(revMap[campFirst]-40.0) > 0.001 {
		t.Errorf("expected First campaign revenue 40.0, got %f", revMap[campFirst])
	}
	if math.Abs(revMap[campLast]-40.0) > 0.001 {
		t.Errorf("expected Last campaign revenue 40.0, got %f", revMap[campLast])
	}
	if math.Abs(revMap[campMid1]-10.0) > 0.001 {
		t.Errorf("expected Mid1 campaign revenue 10.0, got %f", revMap[campMid1])
	}
	if math.Abs(revMap[campMid2]-10.0) > 0.001 {
		t.Errorf("expected Mid2 campaign revenue 10.0, got %f", revMap[campMid2])
	}
}

func TestCalculateAttribution_InvalidModel(t *testing.T) {
	calc := NewCalculator()
	_, err := calc.Calculate(nil, "unknown_model", 0)
	if err == nil {
		t.Errorf("expected error for invalid attribution model")
	}
}
