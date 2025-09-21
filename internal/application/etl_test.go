package application

import (
	"testing"
	"time"

	"github.com/m4ck-y/ETL_go/internal/domain/models"
)

func TestBuildUTMKey(t *testing.T) {
	tests := []struct {
		name        string
		campaign    string
		source      string
		medium      string
		expectedKey models.UTMKey
	}{
		{
			name:     "UTMs válidos",
			campaign: "summer_sale",
			source:   "google",
			medium:   "cpc",
			expectedKey: models.UTMKey{
				Campaign: "summer_sale",
				Source:   "google",
				Medium:   "cpc",
			},
		},
		{
			name:     "UTMs vacíos se normalizan",
			campaign: "",
			source:   "",
			medium:   "",
			expectedKey: models.UTMKey{
				Campaign: "unknown_campaign",
				Source:   "unknown_source",
				Medium:   "unknown_medium",
			},
		},
		{
			name:     "UTMs con espacios se limpian",
			campaign: " summer sale ",
			source:   " GOOGLE ",
			medium:   " CPC ",
			expectedKey: models.UTMKey{
				Campaign: "summer sale",
				Source:   "google",
				Medium:   "cpc",
			},
		},
		{
			name:     "UTMs con mayúsculas se convierten a minúsculas",
			campaign: "SUMMER_SALE",
			source:   "GOOGLE",
			medium:   "CPC",
			expectedKey: models.UTMKey{
				Campaign: "summer_sale",
				Source:   "google",
				Medium:   "cpc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildUTMKey(tt.campaign, tt.source, tt.medium)

			if result != tt.expectedKey {
				t.Errorf("BuildUTMKey() = %v, want %v", result, tt.expectedKey)
			}
		})
	}
}

func TestParseRecordDate(t *testing.T) {
	tests := []struct {
		name        string
		dateStr     string
		expectError bool
		expected    time.Time
	}{
		{
			name:        "Fecha vacía",
			dateStr:     "",
			expectError: true,
		},
		{
			name:        "Fecha YYYY-MM-DD válida",
			dateStr:     "2025-01-15",
			expectError: false,
			expected:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:        "Fecha con tiempo",
			dateStr:     "2025-01-15T10:30:00Z",
			expectError: false,
			expected:    time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			name:        "Fecha RFC3339",
			dateStr:     "2025-01-15T10:30:00.123456Z",
			expectError: false,
			expected:    time.Date(2025, 1, 15, 10, 30, 0, 123456000, time.UTC),
		},
		{
			name:        "Fecha inválida",
			dateStr:     "invalid-date",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseRecordDate(tt.dateStr)

			if tt.expectError {
				if err == nil {
					t.Errorf("parseRecordDate() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("parseRecordDate() unexpected error: %v", err)
				return
			}

			if !result.Equal(tt.expected) {
				t.Errorf("parseRecordDate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsRecordInDateRange(t *testing.T) {
	// Fecha de filtro: 2025-01-15
	filterDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		recordDate string
		filterDate *time.Time
		expected   bool
	}{
		{
			name:       "Sin filtro de fecha",
			recordDate: "2025-01-10",
			filterDate: nil,
			expected:   true,
		},
		{
			name:       "Fecha anterior al filtro",
			recordDate: "2025-01-10",
			filterDate: &filterDate,
			expected:   false,
		},
		{
			name:       "Fecha igual al filtro",
			recordDate: "2025-01-15",
			filterDate: &filterDate,
			expected:   true,
		},
		{
			name:       "Fecha posterior al filtro",
			recordDate: "2025-01-20",
			filterDate: &filterDate,
			expected:   true,
		},
		{
			name:       "Fecha inválida",
			recordDate: "invalid-date",
			filterDate: &filterDate,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRecordInDateRange(tt.recordDate, tt.filterDate)

			if result != tt.expected {
				t.Errorf("isRecordInDateRange() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestProcessAdsMetrics(t *testing.T) {
	// Fecha de filtro: 2025-01-15
	filterDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	ads := []models.AdRecord{
		{Date: "2025-01-10", CampaignID: "C1", UTMSource: "google", UTMCampaign: "sale", UTMMedium: "cpc", Clicks: 100, Cost: 50.0},
		{Date: "2025-01-15", CampaignID: "C2", UTMSource: "google", UTMCampaign: "sale", UTMMedium: "cpc", Clicks: 200, Cost: 100.0},
		{Date: "2025-01-20", CampaignID: "C3", UTMSource: "google", UTMCampaign: "sale", UTMMedium: "cpc", Clicks: 150, Cost: 75.0},
	}

	metrics := make(map[models.UTMKey]models.AggregatedMetrics)

	processAdsMetrics(ads, &filterDate, metrics)

	// Debería incluir solo los registros del 15 y 20 (2 registros)
	expectedClicks := 200 + 150  // 350
	expectedCost := 100.0 + 75.0 // 175.0

	if len(metrics) != 1 {
		t.Errorf("Expected 1 UTM key, got %d", len(metrics))
	}

	for _, m := range metrics {
		if m.Clicks != expectedClicks {
			t.Errorf("Expected %d clicks, got %d", expectedClicks, m.Clicks)
		}
		if m.Cost != expectedCost {
			t.Errorf("Expected %.2f cost, got %.2f", expectedCost, m.Cost)
		}
	}
}

func TestProcessCRMMetrics(t *testing.T) {
	// Fecha de filtro: 2025-01-15
	filterDate := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	crms := []models.CRMRecord{
		{CreatedAt: "2025-01-10", Stage: "lead", UTMCampaign: "sale", UTMSource: "google", UTMMedium: "cpc", Amount: 1000.0},
		{CreatedAt: "2025-01-15", Stage: "lead", UTMCampaign: "sale", UTMSource: "google", UTMMedium: "cpc", Amount: 0.0},
		{CreatedAt: "2025-01-15", Stage: "closed_won", UTMCampaign: "sale", UTMSource: "google", UTMMedium: "cpc", Amount: 2000.0},
		{CreatedAt: "2025-01-20", Stage: "opportunity", UTMCampaign: "sale", UTMSource: "google", UTMMedium: "cpc", Amount: 0.0},
	}

	metrics := make(map[models.UTMKey]models.AggregatedMetrics)

	processCRMMetrics(crms, &filterDate, metrics)

	// Debería procesar solo los registros del 15 y 20
	if len(metrics) != 1 {
		t.Errorf("Expected 1 UTM key, got %d", len(metrics))
	}

	for _, m := range metrics {
		if m.Leads != 1 {
			t.Errorf("Expected 1 lead, got %d", m.Leads)
		}
		if m.Opportunities != 3 { // lead + closed_won + opportunity
			t.Errorf("Expected 3 opportunities, got %d", m.Opportunities)
		}
		if m.ClosedWon != 1 {
			t.Errorf("Expected 1 closed won, got %d", m.ClosedWon)
		}
		if m.Revenue != 2000.0 {
			t.Errorf("Expected 2000.0 revenue, got %.2f", m.Revenue)
		}
	}
}
