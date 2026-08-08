package analytics

import (
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestFunnelEvaluator_MultiStepDropOff(t *testing.T) {
	eval := NewFunnelEvaluator()

	linkStep1 := uuid.New()
	linkStep2 := uuid.New()
	linkStep3 := uuid.New()

	userID1 := uuid.New()
	userID2 := uuid.New()
	userID3 := uuid.New()

	now := time.Now()

	events := []ClickEvent{
		// Visitor 1: Completes Step 1 -> Step 2 -> Step 3
		{ID: uuid.New(), UserID: userID1, LinkID: linkStep1, Timestamp: now.Add(-30 * time.Minute)},
		{ID: uuid.New(), UserID: userID1, LinkID: linkStep2, Timestamp: now.Add(-20 * time.Minute)},
		{ID: uuid.New(), UserID: userID1, LinkID: linkStep3, Timestamp: now.Add(-10 * time.Minute)},

		// Visitor 2: Completes Step 1 -> Step 2 (drops off before Step 3)
		{ID: uuid.New(), UserID: userID2, LinkID: linkStep1, Timestamp: now.Add(-25 * time.Minute)},
		{ID: uuid.New(), UserID: userID2, LinkID: linkStep2, Timestamp: now.Add(-15 * time.Minute)},

		// Visitor 3: Completes Step 1 only (drops off before Step 2)
		{ID: uuid.New(), UserID: userID3, LinkID: linkStep1, Timestamp: now.Add(-5 * time.Minute)},
	}

	payload := FunnelQueryPayload{
		FunnelName: "Checkout Funnel",
		Steps: []FunnelStepInput{
			{StepOrder: 1, Name: "Landing Page", LinkID: linkStep1},
			{StepOrder: 2, Name: "Signup Page", LinkID: linkStep2},
			{StepOrder: 3, Name: "Purchase Confirmation", LinkID: linkStep3},
		},
		From: now.Add(-1 * time.Hour),
		To:   now.Add(1 * time.Hour),
	}

	res, err := eval.Evaluate(events, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.FunnelName != "Checkout Funnel" {
		t.Errorf("expected funnel name 'Checkout Funnel', got '%s'", res.FunnelName)
	}

	if res.TotalStarted != 3 {
		t.Errorf("expected TotalStarted 3, got %d", res.TotalStarted)
	}

	if res.TotalConverted != 1 {
		t.Errorf("expected TotalConverted 1, got %d", res.TotalConverted)
	}

	if math.Abs(res.FinalConversion-33.333) > 0.1 {
		t.Errorf("expected final conversion ~33.33%%, got %f", res.FinalConversion)
	}

	if len(res.Steps) != 3 {
		t.Fatalf("expected 3 step results, got %d", len(res.Steps))
	}

	// Step 1: 3 visitors, 100% overall, 100% step, 0 drop-off
	s1 := res.Steps[0]
	if s1.Visitors != 3 {
		t.Errorf("step 1 expected 3 visitors, got %d", s1.Visitors)
	}
	if s1.DropOffCount != 0 {
		t.Errorf("step 1 expected 0 drop-off, got %d", s1.DropOffCount)
	}

	// Step 2: 2 visitors, 66.67% overall, 66.67% step, 1 drop-off (33.33%)
	s2 := res.Steps[1]
	if s2.Visitors != 2 {
		t.Errorf("step 2 expected 2 visitors, got %d", s2.Visitors)
	}
	if s2.DropOffCount != 1 {
		t.Errorf("step 2 expected 1 drop-off count, got %d", s2.DropOffCount)
	}
	if math.Abs(s2.DropOffPercentage-33.333) > 0.1 {
		t.Errorf("step 2 expected ~33.33%% drop-off, got %f", s2.DropOffPercentage)
	}

	// Step 3: 1 visitor, 33.33% overall, 50% step, 1 drop-off (50%)
	s3 := res.Steps[2]
	if s3.Visitors != 1 {
		t.Errorf("step 3 expected 1 visitor, got %d", s3.Visitors)
	}
	if s3.DropOffCount != 1 {
		t.Errorf("step 3 expected 1 drop-off count, got %d", s3.DropOffCount)
	}
	if math.Abs(s3.StepConversion-50.0) > 0.001 {
		t.Errorf("step 3 expected 50%% step conversion, got %f", s3.StepConversion)
	}
}

func TestFunnelEvaluator_EmptyStepsPayload(t *testing.T) {
	eval := NewFunnelEvaluator()
	payload := FunnelQueryPayload{
		FunnelName: "Empty Funnel",
		Steps:      []FunnelStepInput{},
	}

	_, err := eval.Evaluate(nil, payload)
	if err == nil {
		t.Errorf("expected error for empty steps payload")
	}
}
