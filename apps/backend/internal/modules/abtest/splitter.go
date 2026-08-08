// Package abtest provides weighted traffic splitting algorithms, sticky visitor session routing, and A/B test winner evaluation.
package abtest

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ABVariant represents a single destination variant in an A/B test experiment.
type ABVariant struct {
	ID               uuid.UUID `json:"id,omitempty" db:"id"`
	LinkID           uuid.UUID `json:"link_id" db:"link_id"`
	Name             string    `json:"name" db:"name"`
	TargetURL        string    `json:"target_url" db:"target_url"`
	WeightPercentage int       `json:"weight_percentage" db:"weight_percentage"`
	ClicksCount      int64     `json:"clicks_count" db:"clicks_count"`
	ConversionsCount int64     `json:"conversions_count" db:"conversions_count"`
	IsWinner         bool      `json:"is_winner" db:"is_winner"`
	CreatedAt        time.Time `json:"created_at,omitempty" db:"created_at"`
}

// ValidateWeights ensures the sum of variant weight percentages equals 100%.
func ValidateWeights(variants []ABVariant) error {
	if len(variants) == 0 {
		return fmt.Errorf("variants list cannot be empty")
	}

	total := 0
	for _, v := range variants {
		if v.WeightPercentage < 0 || v.WeightPercentage > 100 {
			return fmt.Errorf("variant %q weight %d is out of range 0-100", v.Name, v.WeightPercentage)
		}
		total += v.WeightPercentage
	}

	if total != 100 {
		return fmt.Errorf("variant weights total %d%%; must sum to exactly 100%%", total)
	}
	return nil
}

// SelectVariant selects a variant using weighted distribution. If visitorID is provided, uses deterministic hashing for sticky sessions.
func SelectVariant(variants []ABVariant, visitorID string) (*ABVariant, error) {
	if len(variants) == 0 {
		return nil, fmt.Errorf("no variants available")
	}

	totalWeight := 0
	for _, v := range variants {
		totalWeight += v.WeightPercentage
	}

	if totalWeight <= 0 {
		return &variants[0], nil
	}

	var bucket int
	visitorID = strings.TrimSpace(visitorID)
	if visitorID != "" {
		// Deterministic FNV-1a hash bucket (1..totalWeight)
		hasher := fnv.New32a()
		_, _ = hasher.Write([]byte(visitorID))
		hashVal := hasher.Sum32()
		bucket = int(hashVal%uint32(totalWeight)) + 1
	} else {
		// Uniform random bucket (1..totalWeight)
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		bucket = r.Intn(totalWeight) + 1
	}

	cumulative := 0
	for i := range variants {
		cumulative += variants[i].WeightPercentage
		if bucket <= cumulative {
			return &variants[i], nil
		}
	}

	return &variants[len(variants)-1], nil
}

// EvaluateWinner determines the winning variant based on highest conversion rate once minimum sample size is met.
func EvaluateWinner(variants []ABVariant, minSampleSize int64) (*ABVariant, error) {
	if len(variants) == 0 {
		return nil, fmt.Errorf("no variants available for winner evaluation")
	}

	var bestVariant *ABVariant
	highestCTR := -1.0

	for i := range variants {
		v := &variants[i]
		if v.ClicksCount < minSampleSize {
			continue
		}

		ctr := float64(v.ConversionsCount) / float64(v.ClicksCount)
		if ctr > highestCTR {
			highestCTR = ctr
			bestVariant = v
		}
	}

	if bestVariant == nil {
		return nil, fmt.Errorf("no variant has reached the minimum sample size of %d clicks", minSampleSize)
	}

	bestVariant.IsWinner = true
	return bestVariant, nil
}
