package ai

import (
	"testing"

	"github.com/google/uuid"
)

func TestAnomalyDetector_TrafficSpike(t *testing.T) {
	detector := NewAnomalyDetector(3.0)
	linkID := uuid.New()

	// Baseline hourly clicks around 100 with small variance
	historical := []float64{95, 102, 98, 105, 100, 97, 101, 99, 104, 96}
	current := 500.0 // Massive traffic spike

	res := detector.DetectAnomaly(linkID, historical, current, 0.05)

	if !res.IsAnomaly {
		t.Errorf("expected anomaly detection to be true for spike of 500 clicks")
	}

	if res.AnomalyType != "traffic_spike" {
		t.Errorf("expected anomaly type 'traffic_spike', got '%s'", res.AnomalyType)
	}

	if res.ZScore <= 3.0 {
		t.Errorf("expected Z-score > 3.0, got %f", res.ZScore)
	}
}

func TestAnomalyDetector_TrafficDrop(t *testing.T) {
	detector := NewAnomalyDetector(3.0)
	linkID := uuid.New()

	historical := []float64{100, 105, 95, 98, 102, 100, 101, 99}
	current := 2.0 // Sudden traffic drop

	res := detector.DetectAnomaly(linkID, historical, current, 0.02)

	if !res.IsAnomaly {
		t.Errorf("expected anomaly detection for traffic drop")
	}

	if res.AnomalyType != "traffic_drop" {
		t.Errorf("expected anomaly type 'traffic_drop', got '%s'", res.AnomalyType)
	}

	if res.ZScore >= -3.0 {
		t.Errorf("expected Z-score < -3.0, got %f", res.ZScore)
	}
}

func TestAnomalyDetector_BotSurge(t *testing.T) {
	detector := NewAnomalyDetector(3.0)
	linkID := uuid.New()

	historical := []float64{100, 102, 98, 100}
	current := 105.0 // Normal volume, but 80% bot traffic
	botPct := 0.80

	res := detector.DetectAnomaly(linkID, historical, current, botPct)

	if !res.IsAnomaly {
		t.Errorf("expected anomaly detection for bot surge")
	}

	if res.AnomalyType != "bot_surge" {
		t.Errorf("expected anomaly type 'bot_surge', got '%s'", res.AnomalyType)
	}
}

func TestCTRPredictionEngine_Predict(t *testing.T) {
	engine := NewCTRPredictor()
	linkID := uuid.New()

	impressions := []int64{1000, 1000, 1000, 1000, 1000}
	clicks := []int64{20, 30, 40, 50, 60} // Upward trend CTR: 2%, 3%, 4%, 5%, 6%

	res := engine.PredictCTR(linkID, impressions, clicks)

	if res.Trend != "upward" {
		t.Errorf("expected trend 'upward', got '%s'", res.Trend)
	}

	if res.PredictedCTR <= res.HistoricalCTR {
		t.Errorf("expected predicted CTR (%f) > historical average CTR (%f)", res.PredictedCTR, res.HistoricalCTR)
	}
}
