package ai

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AnomalyLog struct {
	ID              uuid.UUID `json:"id"`
	LinkID          uuid.UUID `json:"link_id"`
	AnomalyType     string    `json:"anomaly_type"` // 'traffic_spike', 'traffic_drop', 'bot_surge'
	ConfidenceScore float64   `json:"confidence_score"`
	Summary         string    `json:"summary"`
	CreatedAt       time.Time `json:"created_at"`
}

type AnomalyDetectionResult struct {
	LinkID          uuid.UUID `json:"link_id"`
	IsAnomaly       bool      `json:"is_anomaly"`
	AnomalyType     string    `json:"anomaly_type,omitempty"`
	ZScore          float64   `json:"z_score"`
	ConfidenceScore float64   `json:"confidence_score"`
	Summary         string    `json:"summary"`
}

type CTRPredictionResult struct {
	LinkID        uuid.UUID `json:"link_id"`
	HistoricalCTR float64   `json:"historical_ctr"`
	PredictedCTR  float64   `json:"predicted_ctr"`
	Trend         string    `json:"trend"` // 'upward', 'downward', 'stable'
	Confidence    float64   `json:"confidence"`
}

type AnomalyDetector struct {
	ZScoreThreshold float64
}

func NewAnomalyDetector(threshold float64) *AnomalyDetector {
	if threshold <= 0 {
		threshold = 3.0
	}
	return &AnomalyDetector{ZScoreThreshold: threshold}
}

// DetectAnomaly calculates rolling Z-scores on traffic history and evaluates bot percentages.
func (d *AnomalyDetector) DetectAnomaly(linkID uuid.UUID, historicalCounts []float64, currentCount float64, botPercentage float64) AnomalyDetectionResult {
	if botPercentage > 0.50 {
		return AnomalyDetectionResult{
			LinkID:          linkID,
			IsAnomaly:       true,
			AnomalyType:     "bot_surge",
			ZScore:          0.0,
			ConfidenceScore: botPercentage,
			Summary:         fmt.Sprintf("High bot activity detected: %.1f%% of incoming traffic identified as automated bots.", botPercentage*100),
		}
	}

	if len(historicalCounts) == 0 {
		return AnomalyDetectionResult{
			LinkID:    linkID,
			IsAnomaly: false,
			Summary:   "Insufficient historical baseline traffic data.",
		}
	}

	mean := 0.0
	for _, val := range historicalCounts {
		mean += val
	}
	mean /= float64(len(historicalCounts))

	variance := 0.0
	for _, val := range historicalCounts {
		diff := val - mean
		variance += diff * diff
	}
	variance /= float64(len(historicalCounts))

	stdDev := math.Sqrt(variance)
	if stdDev == 0 {
		stdDev = 0.0001
	}

	zScore := (currentCount - mean) / stdDev
	absZ := math.Abs(zScore)

	if absZ >= d.ZScoreThreshold {
		anomalyType := "traffic_spike"
		if zScore < 0 {
			anomalyType = "traffic_drop"
		}

		confidence := math.Min(0.99, 0.50+(absZ*0.10))

		return AnomalyDetectionResult{
			LinkID:          linkID,
			IsAnomaly:       true,
			AnomalyType:     anomalyType,
			ZScore:          zScore,
			ConfidenceScore: confidence,
			Summary:         fmt.Sprintf("Traffic anomaly detected (%s): Z-score of %.2f deviates significantly from historical mean of %.1f.", anomalyType, zScore, mean),
		}
	}

	return AnomalyDetectionResult{
		LinkID:          linkID,
		IsAnomaly:       false,
		ZScore:          zScore,
		ConfidenceScore: 0.0,
		Summary:         "Traffic is within expected normal variance bounds.",
	}
}

type CTRPredictor struct{}

func NewCTRPredictor() *CTRPredictor {
	return &CTRPredictor{}
}

// PredictCTR evaluates historical impression & click trends using linear regression forecasting.
func (p *CTRPredictor) PredictCTR(linkID uuid.UUID, impressions []int64, clicks []int64) CTRPredictionResult {
	n := len(impressions)
	if n == 0 || n != len(clicks) {
		return CTRPredictionResult{
			LinkID:        linkID,
			HistoricalCTR: 0,
			PredictedCTR:  0,
			Trend:         "stable",
			Confidence:    0,
		}
	}

	totalImpressions := int64(0)
	totalClicks := int64(0)
	ctrs := make([]float64, n)

	for i := 0; i < n; i++ {
		totalImpressions += impressions[i]
		totalClicks += clicks[i]
		if impressions[i] > 0 {
			ctrs[i] = (float64(clicks[i]) / float64(impressions[i])) * 100.0
		}
	}

	avgCTR := 0.0
	if totalImpressions > 0 {
		avgCTR = (float64(totalClicks) / float64(totalImpressions)) * 100.0
	}

	// Linear regression slope on CTRs over time
	sumX, sumY, sumXY, sumXX := 0.0, 0.0, 0.0, 0.0
	for i := 0; i < n; i++ {
		x := float64(i)
		y := ctrs[i]
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	fn := float64(n)
	slope := (fn*sumXY - sumX*sumY) / (fn*sumXX - sumX*sumX)
	intercept := (sumY - slope*sumX) / fn

	predictedCTR := intercept + slope*float64(n)
	if predictedCTR < 0 {
		predictedCTR = 0
	}

	trend := "stable"
	if slope > 0.1 {
		trend = "upward"
	} else if slope < -0.1 {
		trend = "downward"
	}

	return CTRPredictionResult{
		LinkID:        linkID,
		HistoricalCTR: avgCTR,
		PredictedCTR:  predictedCTR,
		Trend:         trend,
		Confidence:    0.92,
	}
}

// GenerateSlugs provides AI-suggested short slug candidates based on URL or title text.
func GenerateSlugs(url string, pageTitle string) []string {
	baseText := pageTitle
	if baseText == "" {
		parts := strings.Split(url, "/")
		baseText = parts[len(parts)-1]
	}

	cleaned := strings.ToLower(baseText)
	cleaned = strings.ReplaceAll(cleaned, "https://", "")
	cleaned = strings.ReplaceAll(cleaned, "http://", "")
	cleaned = strings.ReplaceAll(cleaned, "www.", "")

	words := strings.FieldsFunc(cleaned, func(r rune) bool {
		return r < 'a' || r > 'z' && r < '0' || r > '9'
	})

	if len(words) == 0 {
		return []string{"link-v1", "go-now", "quick-link"}
	}

	slug1 := strings.Join(words, "-")
	if len(slug1) > 20 {
		slug1 = slug1[:20]
	}

	slug2 := words[0]
	if len(words) > 1 {
		slug2 += "-" + words[len(words)-1]
	}

	slug3 := "get-" + words[0]

	return []string{slug1, slug2, slug3}
}

// SummarizeCampaign generates executive summary briefs for campaign performance.
func SummarizeCampaign(campaignName string, spend float64, revenue float64, conversions int64, topPlatform string) string {
	roas := 0.0
	if spend > 0 {
		roas = revenue / spend
	}
	return fmt.Sprintf("Campaign '%s' generated %d conversions and $%.2f in total revenue from $%.2f spend (ROAS: %.2fx), driven primarily by %s traffic.",
		campaignName, conversions, revenue, spend, roas, topPlatform)
}
