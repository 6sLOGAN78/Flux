package analytics

import (
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRevenueCalculator_CalculateCampaignMetrics(t *testing.T) {
	calc := NewRevenueCalculator()

	campID := uuid.New()
	cust1 := uuid.New()
	cust2 := uuid.New()

	spends := []AdSpend{
		{ID: uuid.New(), CampaignID: campID, CampaignName: "Search Ads", AmountSpent: 1000.0, Platform: "google", Date: time.Now()},
		{ID: uuid.New(), CampaignID: campID, CampaignName: "Search Ads", AmountSpent: 500.0, Platform: "facebook", Date: time.Now()},
	}

	conversions := []CustomerConversion{
		{CustomerID: cust1, CampaignID: campID, Revenue: 3000.0, ConvertedAt: time.Now()},
		{CustomerID: cust2, CampaignID: campID, Revenue: 1500.0, ConvertedAt: time.Now()},
		{CustomerID: cust1, CampaignID: campID, Revenue: 500.0, ConvertedAt: time.Now()}, // Repeat purchase for cust1
	}

	res, err := calc.Calculate(spends, conversions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TotalSpend != 1500.0 {
		t.Errorf("expected total spend 1500.0, got %f", res.TotalSpend)
	}

	if res.TotalRevenue != 5000.0 {
		t.Errorf("expected total revenue 5000.0, got %f", res.TotalRevenue)
	}

	if res.TotalCustomers != 2 {
		t.Errorf("expected 2 total unique customers, got %d", res.TotalCustomers)
	}

	// Overall CAC = Spend (1500) / Customers (2) = 750.0
	if math.Abs(res.OverallCAC-750.0) > 0.001 {
		t.Errorf("expected overall CAC 750.0, got %f", res.OverallCAC)
	}

	// Overall ROAS = Revenue (5000) / Spend (1500) = 3.3333...
	if math.Abs(res.OverallROAS-3.3333) > 0.01 {
		t.Errorf("expected overall ROAS ~3.33, got %f", res.OverallROAS)
	}

	// Overall ROI = (5000 - 1500) / 1500 * 100 = 233.333...%
	if math.Abs(res.OverallROI-233.333) > 0.1 {
		t.Errorf("expected overall ROI ~233.33%%, got %f", res.OverallROI)
	}

	// Overall LTV = Revenue (5000) / Customers (2) = 2500.0
	if math.Abs(res.OverallLTV-2500.0) > 0.001 {
		t.Errorf("expected overall LTV 2500.0, got %f", res.OverallLTV)
	}

	if len(res.Campaigns) != 1 {
		t.Fatalf("expected 1 campaign metric, got %d", len(res.Campaigns))
	}

	c := res.Campaigns[0]
	if c.CampaignID != campID {
		t.Errorf("expected campaign ID %s, got %s", campID, c.CampaignID)
	}

	if math.Abs(c.LTVtoCACRatio-3.3333) > 0.01 { // LTV (2500) / CAC (750) = 3.333...
		t.Errorf("expected LTV:CAC ratio ~3.33, got %f", c.LTVtoCACRatio)
	}
}

func TestRevenueCalculator_ZeroSpendHandling(t *testing.T) {
	calc := NewRevenueCalculator()

	campID := uuid.New()
	cust1 := uuid.New()

	spends := []AdSpend{}
	conversions := []CustomerConversion{
		{CustomerID: cust1, CampaignID: campID, Revenue: 100.0, ConvertedAt: time.Now()},
	}

	res, err := calc.Calculate(spends, conversions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.OverallROAS != 0.0 {
		t.Errorf("expected 0.0 ROAS for zero spend, got %f", res.OverallROAS)
	}
	if res.OverallCAC != 0.0 {
		t.Errorf("expected 0.0 CAC for zero spend, got %f", res.OverallCAC)
	}
}
