package controllers

import (
	"encoding/json"
	"math"
	"testing"
)

func TestExtractSemesterStatistics_EmptyStrings(t *testing.T) {
	statsResp := map[string]interface{}{
		"result": map[string]interface{}{
			"averageScore": "",
			"creditTotal":  " ",
		},
	}

	avg, credit := extractSemesterStatistics(statsResp)
	if avg != nil {
		t.Fatalf("expected avg nil, got %v", *avg)
	}
	if credit != nil {
		t.Fatalf("expected credit nil, got %v", *credit)
	}
}

func TestExtractSemesterStatistics_NormalValues(t *testing.T) {
	statsResp := map[string]interface{}{
		"result": map[string]interface{}{
			"avgScore":     "85.126",
			"totalCredit":  json.Number("12.5"),
			"averageScore": "",
		},
	}

	avg, credit := extractSemesterStatistics(statsResp)
	if avg == nil {
		t.Fatalf("expected avg not nil")
	}
	if credit == nil {
		t.Fatalf("expected credit not nil")
	}

	if math.Abs(*avg-85.13) > 1e-9 {
		t.Fatalf("expected avg 85.13, got %v", *avg)
	}
	if math.Abs(*credit-12.5) > 1e-9 {
		t.Fatalf("expected credit 12.5, got %v", *credit)
	}
}

func TestExtractSemesterStatistics_BoundaryZeroValues(t *testing.T) {
	statsResp := map[string]interface{}{
		"result": map[string]interface{}{
			"averageScore": 0,
			"creditTotal":  "0",
		},
	}

	avg, credit := extractSemesterStatistics(statsResp)
	if avg == nil || credit == nil {
		t.Fatalf("expected avg and credit not nil, got avg=%v credit=%v", avg, credit)
	}
	if *avg != 0 {
		t.Fatalf("expected avg 0, got %v", *avg)
	}
	if *credit != 0 {
		t.Fatalf("expected credit 0, got %v", *credit)
	}
}

