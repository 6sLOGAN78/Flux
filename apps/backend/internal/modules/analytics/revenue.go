package analytics

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

type AdSpend struct {
	ID           uuid.UUID `json:"id"`
	CampaignID   uuid.UUID `json:"campaign_id"`
	CampaignName string    `json:"campaign_name,omitempty"`
	Date         time.Time `json:"date"`
	AmountSpent  float64   `json:"amount_spent"`
	Platform     string    `json:"platform"`
}

type CustomerConversion struct {
	CustomerID  uuid.UUID `json:"customer_id"`
	CampaignID  uuid.UUID `json:"campaign_id"`
	Revenue     float64   `json:"revenue"`
	ConvertedAt time.Time `json:"converted_at"`
}

type CampaignRevenueMetrics struct {
	CampaignID        uuid.UUID `json:"campaign_id"`
	CampaignName      string    `json:"campaign_name,omitempty"`
	TotalSpend        float64   `json:"spend"`
	TotalRevenue      float64   `json:"revenue"`
	CustomersAcquired int64     `json:"customers_acquired"`
	CAC               float64   `json:"cac"`
	ROAS              float64   `json:"roas"`
	ROIPercentage     float64   `json:"roi_pct"`
	LTV               float64   `json:"ltv"`
	LTVtoCACRatio     float64   `json:"ltv_to_cac_ratio"`
}

type RevenueSummaryResult struct {
	TotalSpend     float64                  `json:"total_spend"`
	TotalRevenue   float64                  `json:"total_revenue"`
	TotalCustomers int64                    `json:"total_customers"`
	OverallCAC     float64                  `json:"overall_cac"`
	OverallROAS    float64                  `json:"overall_roas"`
	OverallROI     float64                  `json:"overall_roi_pct"`
	OverallLTV     float64                  `json:"overall_ltv"`
	Campaigns      []CampaignRevenueMetrics `json:"campaigns"`
}

type RevenueCalculator struct{}

func NewRevenueCalculator() *RevenueCalculator {
	return &RevenueCalculator{}
}

// Calculate aggregates spend and conversion revenue to compute LTV, ROAS, CAC, ROI, and LTV:CAC ratios per campaign and overall.
func (c *RevenueCalculator) Calculate(spends []AdSpend, conversions []CustomerConversion) (*RevenueSummaryResult, error) {
	campaignSpends := make(map[uuid.UUID]float64)
	campaignNames := make(map[uuid.UUID]string)
	totalSpend := 0.0

	for _, s := range spends {
		campaignSpends[s.CampaignID] += s.AmountSpent
		if s.CampaignName != "" {
			campaignNames[s.CampaignID] = s.CampaignName
		}
		totalSpend += s.AmountSpent
	}

	campaignRevenues := make(map[uuid.UUID]float64)
	campaignCustomers := make(map[uuid.UUID]map[uuid.UUID]bool)
	allCustomers := make(map[uuid.UUID]bool)
	totalRevenue := 0.0

	for _, conv := range conversions {
		campaignRevenues[conv.CampaignID] += conv.Revenue
		totalRevenue += conv.Revenue

		if conv.CustomerID != uuid.Nil {
			allCustomers[conv.CustomerID] = true
			if campaignCustomers[conv.CampaignID] == nil {
				campaignCustomers[conv.CampaignID] = make(map[uuid.UUID]bool)
			}
			campaignCustomers[conv.CampaignID][conv.CustomerID] = true
		}
	}

	allCampaignIDs := make(map[uuid.UUID]bool)
	for id := range campaignSpends {
		allCampaignIDs[id] = true
	}
	for id := range campaignRevenues {
		allCampaignIDs[id] = true
	}

	campaignMetrics := make([]CampaignRevenueMetrics, 0, len(allCampaignIDs))

	for id := range allCampaignIDs {
		spend := campaignSpends[id]
		rev := campaignRevenues[id]
		custCount := int64(len(campaignCustomers[id]))

		cac := 0.0
		if custCount > 0 {
			cac = spend / float64(custCount)
		}

		roas := 0.0
		roi := 0.0
		if spend > 0 {
			roas = rev / spend
			roi = ((rev - spend) / spend) * 100.0
		}

		ltv := 0.0
		if custCount > 0 {
			ltv = rev / float64(custCount)
		}

		ltvCac := 0.0
		if cac > 0 {
			ltvCac = ltv / cac
		}

		campaignMetrics = append(campaignMetrics, CampaignRevenueMetrics{
			CampaignID:        id,
			CampaignName:      campaignNames[id],
			TotalSpend:        spend,
			TotalRevenue:      rev,
			CustomersAcquired: custCount,
			CAC:               cac,
			ROAS:              roas,
			ROIPercentage:     roi,
			LTV:               ltv,
			LTVtoCACRatio:     ltvCac,
		})
	}

	sort.Slice(campaignMetrics, func(i, j int) bool {
		return campaignMetrics[i].TotalRevenue > campaignMetrics[j].TotalRevenue
	})

	totalCustCount := int64(len(allCustomers))

	overallCAC := 0.0
	if totalCustCount > 0 {
		overallCAC = totalSpend / float64(totalCustCount)
	}

	overallROAS := 0.0
	overallROI := 0.0
	if totalSpend > 0 {
		overallROAS = totalRevenue / totalSpend
		overallROI = ((totalRevenue - totalSpend) / totalSpend) * 100.0
	}

	overallLTV := 0.0
	if totalCustCount > 0 {
		overallLTV = totalRevenue / float64(totalCustCount)
	}

	return &RevenueSummaryResult{
		TotalSpend:     totalSpend,
		TotalRevenue:   totalRevenue,
		TotalCustomers: totalCustCount,
		OverallCAC:     overallCAC,
		OverallROAS:    overallROAS,
		OverallROI:     overallROI,
		OverallLTV:     overallLTV,
		Campaigns:      campaignMetrics,
	}, nil
}
