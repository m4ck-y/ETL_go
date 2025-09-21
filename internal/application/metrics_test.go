package application

import (
	"testing"

	"github.com/m4ck-y/ETL_go/internal/domain/models"
)

func TestSafeDivide(t *testing.T) {
	tests := []struct {
		name        string
		numerator   float64
		denominator float64
		expected    float64
	}{
		{
			name:        "División normal",
			numerator:   10.0,
			denominator: 2.0,
			expected:    5.0,
		},
		{
			name:        "División por cero",
			numerator:   10.0,
			denominator: 0.0,
			expected:    0.0,
		},
		{
			name:        "División por cero (denominador cero)",
			numerator:   10.0,
			denominator: 0,
			expected:    0.0,
		},
		{
			name:        "Numerador cero",
			numerator:   0.0,
			denominator: 5.0,
			expected:    0.0,
		},
		{
			name:        "Ambos cero",
			numerator:   0.0,
			denominator: 0.0,
			expected:    0.0,
		},
		{
			name:        "Números decimales",
			numerator:   7.5,
			denominator: 2.5,
			expected:    3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeDivide(tt.numerator, tt.denominator)

			if result != tt.expected {
				t.Errorf("safeDivide(%v, %v) = %v, want %v", tt.numerator, tt.denominator, result, tt.expected)
			}
		})
	}
}

func TestCalculateDerivedMetrics(t *testing.T) {
	tests := []struct {
		name                 string
		metrics              models.AggregatedMetrics
		expectedCPC          float64
		expectedCPA          float64
		expectedCVRLeadToOpp float64
		expectedCVROppToWon  float64
		expectedROAS         float64
	}{
		{
			name: "Métricas normales",
			metrics: models.AggregatedMetrics{
				Clicks:        1000,
				Cost:          500.0,
				Leads:         50,
				Opportunities: 30,
				ClosedWon:     15,
				Revenue:       7500.0,
			},
			expectedCPC:          0.5,  // 500 / 1000
			expectedCPA:          10.0, // 500 / 50
			expectedCVRLeadToOpp: 0.6,  // 30 / 50
			expectedCVROppToWon:  0.5,  // 15 / 30
			expectedROAS:         15.0, // 7500 / 500
		},
		{
			name: "Sin clics (CPC = 0)",
			metrics: models.AggregatedMetrics{
				Clicks:        0,
				Cost:          100.0,
				Leads:         10,
				Opportunities: 5,
				ClosedWon:     2,
				Revenue:       500.0,
			},
			expectedCPC:          0.0,  // 100 / 0 = 0 (protegido)
			expectedCPA:          10.0, // 100 / 10
			expectedCVRLeadToOpp: 0.5,  // 5 / 10
			expectedCVROppToWon:  0.4,  // 2 / 5
			expectedROAS:         5.0,  // 500 / 100
		},
		{
			name: "Sin leads (CPA = 0, CVR = 0)",
			metrics: models.AggregatedMetrics{
				Clicks:        500,
				Cost:          250.0,
				Leads:         0,
				Opportunities: 0,
				ClosedWon:     0,
				Revenue:       0.0,
			},
			expectedCPC:          0.5, // 250 / 500
			expectedCPA:          0.0, // 250 / 0 = 0 (protegido)
			expectedCVRLeadToOpp: 0.0, // 0 / 0 = 0 (protegido)
			expectedCVROppToWon:  0.0, // 0 / 0 = 0 (protegido)
			expectedROAS:         0.0, // 0 / 250 = 0 (protegido)
		},
		{
			name: "Sin oportunidades (CVR = 0)",
			metrics: models.AggregatedMetrics{
				Clicks:        200,
				Cost:          100.0,
				Leads:         20,
				Opportunities: 0,
				ClosedWon:     0,
				Revenue:       0.0,
			},
			expectedCPC:          0.5, // 100 / 200
			expectedCPA:          5.0, // 100 / 20
			expectedCVRLeadToOpp: 0.0, // 0 / 20 = 0 (protegido)
			expectedCVROppToWon:  0.0, // 0 / 0 = 0 (protegido)
			expectedROAS:         0.0, // 0 / 100 = 0 (protegido)
		},
		{
			name: "Sin costo (ROAS = 0, CPC = 0, CPA = 0)",
			metrics: models.AggregatedMetrics{
				Clicks:        100,
				Cost:          0.0,
				Leads:         10,
				Opportunities: 5,
				ClosedWon:     2,
				Revenue:       1000.0,
			},
			expectedCPC:          0.0, // 0 / 100 = 0 (protegido)
			expectedCPA:          0.0, // 0 / 10 = 0 (protegido)
			expectedCVRLeadToOpp: 0.5, // 5 / 10
			expectedCVROppToWon:  0.4, // 2 / 5
			expectedROAS:         0.0, // 1000 / 0 = 0 (protegido)
		},
		{
			name: "Caso realista de marketing",
			metrics: models.AggregatedMetrics{
				Clicks:        5000,
				Cost:          1250.0,
				Leads:         125,
				Opportunities: 75,
				ClosedWon:     30,
				Revenue:       15000.0,
			},
			expectedCPC:          0.25, // 1250 / 5000
			expectedCPA:          10.0, // 1250 / 125
			expectedCVRLeadToOpp: 0.6,  // 75 / 125
			expectedCVROppToWon:  0.4,  // 30 / 75
			expectedROAS:         12.0, // 15000 / 1250
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpc, cpa, cvrLeadToOpp, cvrOppToWon, roas := CalculateDerivedMetrics(tt.metrics)

			if cpc != tt.expectedCPC {
				t.Errorf("CPC = %v, want %v", cpc, tt.expectedCPC)
			}
			if cpa != tt.expectedCPA {
				t.Errorf("CPA = %v, want %v", cpa, tt.expectedCPA)
			}
			if cvrLeadToOpp != tt.expectedCVRLeadToOpp {
				t.Errorf("CVR Lead to Opp = %v, want %v", cvrLeadToOpp, tt.expectedCVRLeadToOpp)
			}
			if cvrOppToWon != tt.expectedCVROppToWon {
				t.Errorf("CVR Opp to Won = %v, want %v", cvrOppToWon, tt.expectedCVROppToWon)
			}
			if roas != tt.expectedROAS {
				t.Errorf("ROAS = %v, want %v", roas, tt.expectedROAS)
			}
		})
	}
}

func TestCalculateDerivedMetricsEdgeCases(t *testing.T) {
	t.Run("Métricas completamente vacías", func(t *testing.T) {
		metrics := models.AggregatedMetrics{}
		cpc, cpa, cvrLeadToOpp, cvrOppToWon, roas := CalculateDerivedMetrics(metrics)

		if cpc != 0.0 || cpa != 0.0 || cvrLeadToOpp != 0.0 || cvrOppToWon != 0.0 || roas != 0.0 {
			t.Errorf("Expected all metrics to be 0.0, got CPC=%v, CPA=%v, CVR1=%v, CVR2=%v, ROAS=%v",
				cpc, cpa, cvrLeadToOpp, cvrOppToWon, roas)
		}
	})

	t.Run("Solo clics sin costo", func(t *testing.T) {
		metrics := models.AggregatedMetrics{
			Clicks: 1000,
			Cost:   0.0,
		}
		cpc, _, _, _, _ := CalculateDerivedMetrics(metrics)

		if cpc != 0.0 {
			t.Errorf("Expected CPC to be 0.0 when cost is 0, got %v", cpc)
		}
	})

	t.Run("Revenue sin costo", func(t *testing.T) {
		metrics := models.AggregatedMetrics{
			Cost:    0.0,
			Revenue: 10000.0,
		}
		_, _, _, _, roas := CalculateDerivedMetrics(metrics)

		if roas != 0.0 {
			t.Errorf("Expected ROAS to be 0.0 when cost is 0, got %v", roas)
		}
	})
}
