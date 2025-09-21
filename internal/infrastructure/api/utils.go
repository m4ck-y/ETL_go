package api

import (
	"crypto/md5"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/m4ck-y/ETL_go/internal/domain/models"
)

// generateBatchID crea un identificador único para lotes ETL
func generateBatchID(adsURL, crmURL, sinceParam string) string {
	// Incluir timestamp diario para granularidad por día
	input := fmt.Sprintf("%s|%s|%s|%d", adsURL, crmURL, sinceParam, time.Now().Unix()/86400)
	hash := md5.Sum([]byte(input))
	return fmt.Sprintf("%x", hash)[:16]
}

// parseSinceDate parsea el parámetro opcional de fecha
func parseSinceDate(sinceParam string) (*time.Time, error) {
	if sinceParam == "" {
		return nil, nil
	}

	parsedDate, err := time.Parse("2006-01-02", sinceParam)
	if err != nil {
		return nil, fmt.Errorf("formato de fecha inválido. Use YYYY-MM-DD")
	}

	return &parsedDate, nil
}

// validateEnvironment valida que las variables de entorno necesarias estén configuradas
func validateEnvironment() error {
	adsURL := os.Getenv("ADS_API_URL")
	crmURL := os.Getenv("CRM_API_URL")

	if adsURL == "" || crmURL == "" {
		return fmt.Errorf("ADS_API_URL y CRM_API_URL deben estar configuradas")
	}

	return nil
}

func parseDateRange(fromParam, toParam string) (*time.Time, *time.Time, error) {
	var fromDate, toDate *time.Time

	if fromParam != "" {
		parsed, err := time.Parse("2006-01-02", fromParam)
		if err != nil {
			return nil, nil, fmt.Errorf("fecha 'from' inválida")
		}
		fromDate = &parsed
	}

	if toParam != "" {
		parsed, err := time.Parse("2006-01-02", toParam)
		if err != nil {
			return nil, nil, fmt.Errorf("fecha 'to' inválida")
		}
		toDate = &parsed
	}

	return fromDate, toDate, nil
}

func filterMetricsByChannel(metrics []models.MetricResponse, channel string) []models.MetricResponse {
	if channel == "" {
		return metrics
	}

	var filtered []models.MetricResponse
	for _, m := range metrics {
		// TODO:
		// Aquí necesitaríamos tener el campo channel en los datos
		// Por ahora, filtramos por UTM source que podría indicar el canal
		if strings.Contains(strings.ToLower(m.UTMSource), strings.ToLower(channel)) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func filterMetricsByCampaign(metrics []models.MetricResponse, campaign string) []models.MetricResponse {
	if campaign == "" {
		return metrics
	}

	var filtered []models.MetricResponse
	for _, m := range metrics {
		if strings.Contains(strings.ToLower(m.UTMCampaign), strings.ToLower(campaign)) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func paginateMetrics(metrics []models.MetricResponse, limit, offset int) []models.MetricResponse {
	if offset >= len(metrics) {
		return []models.MetricResponse{}
	}

	end := offset + limit
	if end > len(metrics) {
		end = len(metrics)
	}

	return metrics[offset:end]
}
