// Package attribution implements multi-touch attribution algorithms: First-Touch, Last-Touch, Linear, Time-Decay, and U-Shaped.
package attribution

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
)

type AttributionModel string

const (
	ModelFirstTouch     AttributionModel = "first_touch"
	ModelLastTouch      AttributionModel = "last_touch"
	ModelLinear         AttributionModel = "linear"
	ModelTimeDecay      AttributionModel = "time_decay"
	ModelPositionBased AttributionModel = "position_based" // U-Shaped
)

type Touchpoint struct {
	ID               uuid.UUID `json:"id"`
	VisitorSessionID uuid.UUID `json:"visitor_session_id"`
	LinkID           uuid.UUID `json:"link_id"`
	CampaignID       uuid.UUID `json:"campaign_id"`
	CampaignName     string    `json:"campaign_name"`
	Timestamp        time.Time `json:"timestamp"`
	ReferrerDomain   string    `json:"referrer_domain,omitempty"`
	UTMSource        string    `json:"utm_source,omitempty"`
	UTMMedium        string    `json:"utm_medium,omitempty"`
	UTMCampaign      string    `json:"utm_campaign,omitempty"`
}

type Conversion struct {
	ID               uuid.UUID    `json:"id"`
	VisitorSessionID uuid.UUID    `json:"visitor_session_id"`
	Revenue          float64      `json:"revenue"`
	ConvertedAt      time.Time    `json:"converted_at"`
	Touchpoints      []Touchpoint `json:"touchpoints"`
}

type CampaignAttribution struct {
	CampaignID            uuid.UUID `json:"campaign_id"`
	CampaignName          string    `json:"campaign_name"`
	AttributedConversions float64   `json:"attributed_conversions"`
	AttributedRevenue     float64   `json:"attributed_revenue"`
}

type AttributionResult struct {
	Model                  AttributionModel      `json:"model"`
	TotalConversions       int                   `json:"total_conversions"`
	TotalAttributedRevenue float64               `json:"total_attributed_revenue"`
	Campaigns              []CampaignAttribution `json:"campaigns"`
}

type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

// Calculate computes campaign attribution credit across conversions based on the selected model.
func (c *Calculator) Calculate(conversions []Conversion, model AttributionModel, halfLife time.Duration) (*AttributionResult, error) {
	switch model {
	case ModelFirstTouch, ModelLastTouch, ModelLinear, ModelTimeDecay, ModelPositionBased:
	default:
		return nil, fmt.Errorf("unsupported attribution model: %s", model)
	}

	if halfLife <= 0 {
		halfLife = 7 * 24 * time.Hour
	}

	campaignMap := make(map[uuid.UUID]*CampaignAttribution)
	var totalRevenue float64
	var totalConversions int

	for _, conv := range conversions {
		if len(conv.Touchpoints) == 0 {
			continue
		}
		totalConversions++
		totalRevenue += conv.Revenue

		// Sort touchpoints chronologically
		sortedTPs := make([]Touchpoint, len(conv.Touchpoints))
		copy(sortedTPs, conv.Touchpoints)
		sort.Slice(sortedTPs, func(i, j int) bool {
			return sortedTPs[i].Timestamp.Before(sortedTPs[j].Timestamp)
		})

		weights := calculateWeights(sortedTPs, conv.ConvertedAt, model, halfLife)

		for i, tp := range sortedTPs {
			w := weights[i]
			if w == 0 {
				continue
			}

			attr, exists := campaignMap[tp.CampaignID]
			if !exists {
				name := tp.CampaignName
				if name == "" {
					name = tp.CampaignID.String()
				}
				attr = &CampaignAttribution{
					CampaignID:   tp.CampaignID,
					CampaignName: name,
				}
				campaignMap[tp.CampaignID] = attr
			}

			attr.AttributedConversions += w * 1.0
			attr.AttributedRevenue += w * conv.Revenue
		}
	}

	campaigns := make([]CampaignAttribution, 0, len(campaignMap))
	for _, attr := range campaignMap {
		campaigns = append(campaigns, *attr)
	}

	// Sort campaigns by AttributedRevenue DESC then CampaignName ASC
	sort.Slice(campaigns, func(i, j int) bool {
		if math.Abs(campaigns[i].AttributedRevenue-campaigns[j].AttributedRevenue) > 0.0001 {
			return campaigns[i].AttributedRevenue > campaigns[j].AttributedRevenue
		}
		return campaigns[i].CampaignName < campaigns[j].CampaignName
	})

	return &AttributionResult{
		Model:                  model,
		TotalConversions:       totalConversions,
		TotalAttributedRevenue: totalRevenue,
		Campaigns:              campaigns,
	}, nil
}

func calculateWeights(tps []Touchpoint, convertedAt time.Time, model AttributionModel, halfLife time.Duration) []float64 {
	n := len(tps)
	weights := make([]float64, n)
	if n == 0 {
		return weights
	}

	switch model {
	case ModelFirstTouch:
		weights[0] = 1.0

	case ModelLastTouch:
		weights[n-1] = 1.0

	case ModelLinear:
		share := 1.0 / float64(n)
		for i := range weights {
			weights[i] = share
		}

	case ModelTimeDecay:
		var sum float64
		halfLifeSec := halfLife.Seconds()
		if halfLifeSec <= 0 {
			halfLifeSec = (7 * 24 * time.Hour).Seconds()
		}

		rawWeights := make([]float64, n)
		for i, tp := range tps {
			diff := convertedAt.Sub(tp.Timestamp).Seconds()
			if diff < 0 {
				diff = 0
			}
			w := math.Pow(2, -diff/halfLifeSec)
			rawWeights[i] = w
			sum += w
		}

		if sum > 0 {
			for i := range weights {
				weights[i] = rawWeights[i] / sum
			}
		} else {
			weights[n-1] = 1.0
		}

	case ModelPositionBased:
		if n == 1 {
			weights[0] = 1.0
		} else if n == 2 {
			weights[0] = 0.5
			weights[1] = 0.5
		} else {
			weights[0] = 0.4
			weights[n-1] = 0.4
			midShare := 0.2 / float64(n-2)
			for i := 1; i < n-1; i++ {
				weights[i] = midShare
			}
		}
	}

	return weights
}
