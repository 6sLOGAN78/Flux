package abtest_test

import (
	"testing"

	"flux/apps/backend/internal/modules/abtest"

	"github.com/google/uuid"
)

func TestSelectVariant_WeightedDistribution(t *testing.T) {
	linkID := uuid.New()
	varA := abtest.ABVariant{ID: uuid.New(), LinkID: linkID, Name: "Control A", TargetURL: "https://acme.com/v1", WeightPercentage: 80}
	varB := abtest.ABVariant{ID: uuid.New(), LinkID: linkID, Name: "Challenger B", TargetURL: "https://acme.com/v2", WeightPercentage: 20}

	variants := []abtest.ABVariant{varA, varB}

	// 1. Sticky visitor check (same visitor ID must ALWAYS select the exact same variant)
	visitorID := "user_session_abc123"
	selected1, err := abtest.SelectVariant(variants, visitorID)
	if err != nil {
		t.Fatalf("unexpected error selecting variant: %v", err)
	}

	selected2, err := abtest.SelectVariant(variants, visitorID)
	if err != nil {
		t.Fatalf("unexpected error selecting variant on second lookup: %v", err)
	}

	if selected1.ID != selected2.ID {
		t.Errorf("expected sticky session selection to match, got %v and %v", selected1.ID, selected2.ID)
	}

	// 2. Random selection statistical distribution across 10,000 iterations
	counts := make(map[string]int)
	iterations := 10000

	for i := 0; i < iterations; i++ {
		sel, err := abtest.SelectVariant(variants, "")
		if err != nil {
			t.Fatalf("unexpected error in random selection: %v", err)
		}
		counts[sel.Name]++
	}

	// Variant A (80%) should be ~8000 (e.g. 7000..9000), Variant B (20%) should be ~2000 (e.g. 1000..3000)
	if counts["Control A"] < 7000 || counts["Control A"] > 9000 {
		t.Errorf("expected Control A count around 8000, got %d", counts["Control A"])
	}
	if counts["Challenger B"] < 1000 || counts["Challenger B"] > 3000 {
		t.Errorf("expected Challenger B count around 2000, got %d", counts["Challenger B"])
	}
}

func TestEvaluateWinner_HighestConversionRate(t *testing.T) {
	linkID := uuid.New()
	varA := abtest.ABVariant{ID: uuid.New(), LinkID: linkID, Name: "Page A", ClicksCount: 500, ConversionsCount: 25} // 5% CTR
	varB := abtest.ABVariant{ID: uuid.New(), LinkID: linkID, Name: "Page B", ClicksCount: 500, ConversionsCount: 75} // 15% CTR

	variants := []abtest.ABVariant{varA, varB}

	winner, err := abtest.EvaluateWinner(variants, 100)
	if err != nil {
		t.Fatalf("unexpected error evaluating winner: %v", err)
	}

	if winner.Name != "Page B" {
		t.Errorf("expected winner to be 'Page B', got %q", winner.Name)
	}
	if !winner.IsWinner {
		t.Error("expected IsWinner to be true on selected winner")
	}
}

func TestValidateWeights(t *testing.T) {
	invalidVariants := []abtest.ABVariant{
		{WeightPercentage: 40},
		{WeightPercentage: 40},
	}

	if err := abtest.ValidateWeights(invalidVariants); err == nil {
		t.Error("expected error when weights sum to 80 instead of 100")
	}

	validVariants := []abtest.ABVariant{
		{WeightPercentage: 60},
		{WeightPercentage: 40},
	}

	if err := abtest.ValidateWeights(validVariants); err != nil {
		t.Errorf("expected valid weights to pass, got: %v", err)
	}
}
