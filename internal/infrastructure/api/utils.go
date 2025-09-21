package api

import (
	"crypto/md5"
	"fmt"
	"os"
	"time"
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
