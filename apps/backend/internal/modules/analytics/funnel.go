// Package analytics provides ClickHouse time-series click ingestion and multi-step conversion funnel drop-off evaluation.
package analytics

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
)

type FunnelStepInput struct {
	StepOrder int       `json:"step_order"`
	Name      string    `json:"name"`
	LinkID    uuid.UUID `json:"link_id"`
}

type FunnelQueryPayload struct {
	FunnelName string            `json:"funnel_name"`
	Steps      []FunnelStepInput `json:"steps"`
	From       time.Time         `json:"from"`
	To         time.Time         `json:"to"`
}

type FunnelStepResult struct {
	StepOrder         int     `json:"step_order"`
	Name              string  `json:"name"`
	LinkID            string  `json:"link_id"`
	Visitors          int64   `json:"visitors"`
	OverallConversion float64 `json:"overall_conversion_pct"`
	StepConversion    float64 `json:"step_conversion_pct"`
	DropOffCount      int64   `json:"drop_off_count"`
	DropOffPercentage float64 `json:"drop_off_pct"`
}

type FunnelAnalysisResult struct {
	FunnelName      string             `json:"funnel_name"`
	TotalStarted    int64              `json:"total_started"`
	TotalConverted  int64              `json:"total_converted"`
	FinalConversion float64            `json:"final_conversion_pct"`
	Steps           []FunnelStepResult `json:"steps"`
}

type FunnelEvaluator struct{}

func NewFunnelEvaluator() *FunnelEvaluator {
	return &FunnelEvaluator{}
}

// Evaluate analyzes visitor click event streams to calculate step-by-step conversion rates and visitor drop-off percentages.
func (e *FunnelEvaluator) Evaluate(events []ClickEvent, payload FunnelQueryPayload) (*FunnelAnalysisResult, error) {
	if len(payload.Steps) == 0 {
		return nil, fmt.Errorf("funnel payload must contain at least one step")
	}

	// Sort steps by StepOrder ASC
	sortedSteps := make([]FunnelStepInput, len(payload.Steps))
	copy(sortedSteps, payload.Steps)
	sort.Slice(sortedSteps, func(i, j int) bool {
		return sortedSteps[i].StepOrder < sortedSteps[j].StepOrder
	})

	// Group events by Visitor identifier (UserID or IPAddress)
	visitorEvents := make(map[string][]ClickEvent)
	for _, evt := range events {
		key := evt.UserID.String()
		if evt.UserID == uuid.Nil {
			key = evt.IPAddress
		}
		if key == "" || key == uuid.Nil.String() {
			continue
		}
		if !payload.From.IsZero() && evt.Timestamp.Before(payload.From) {
			continue
		}
		if !payload.To.IsZero() && evt.Timestamp.After(payload.To) {
			continue
		}
		visitorEvents[key] = append(visitorEvents[key], evt)
	}

	// For each visitor, sort their events chronologically and trace how far down the funnel sequence they progressed
	stepCounts := make([]int64, len(sortedSteps))

	for _, vEvents := range visitorEvents {
		sort.Slice(vEvents, func(i, j int) bool {
			return vEvents[i].Timestamp.Before(vEvents[j].Timestamp)
		})

		currentStepIdx := 0
		for _, evt := range vEvents {
			if currentStepIdx < len(sortedSteps) && evt.LinkID == sortedSteps[currentStepIdx].LinkID {
				stepCounts[currentStepIdx]++
				currentStepIdx++
			}
		}
	}

	totalStarted := stepCounts[0]
	totalConverted := int64(0)
	if len(stepCounts) > 0 {
		totalConverted = stepCounts[len(stepCounts)-1]
	}

	finalConversion := 0.0
	if totalStarted > 0 {
		finalConversion = (float64(totalConverted) / float64(totalStarted)) * 100.0
	}

	stepResults := make([]FunnelStepResult, len(sortedSteps))
	for i, step := range sortedSteps {
		count := stepCounts[i]

		overallPct := 0.0
		if totalStarted > 0 {
			overallPct = (float64(count) / float64(totalStarted)) * 100.0
		}

		stepPct := 100.0
		dropOffCount := int64(0)
		dropOffPct := 0.0

		if i > 0 {
			prevCount := stepCounts[i-1]
			if prevCount > 0 {
				stepPct = (float64(count) / float64(prevCount)) * 100.0
				dropOffCount = prevCount - count
				dropOffPct = (float64(dropOffCount) / float64(prevCount)) * 100.0
			} else {
				stepPct = 0.0
			}
		}

		stepResults[i] = FunnelStepResult{
			StepOrder:         step.StepOrder,
			Name:              step.Name,
			LinkID:            step.LinkID.String(),
			Visitors:          count,
			OverallConversion: overallPct,
			StepConversion:    stepPct,
			DropOffCount:      dropOffCount,
			DropOffPercentage: dropOffPct,
		}
	}

	return &FunnelAnalysisResult{
		FunnelName:      payload.FunnelName,
		TotalStarted:    totalStarted,
		TotalConverted:  totalConverted,
		FinalConversion: finalConversion,
		Steps:           stepResults,
	}, nil
}
