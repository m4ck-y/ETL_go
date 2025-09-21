package application

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/m4ck-y/ETL_go/internal/domain/models"
)

// normalizeUTM normaliza y asigna valores por defecto a UTMs
func normalizeUTM(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return strings.ToLower(strings.TrimSpace(value))
}

func BuildUTMKey(campaign, source, medium string) models.UTMKey {
	return models.UTMKey{
		Campaign: normalizeUTM(campaign, "unknown_campaign"),
		Source:   normalizeUTM(source, "unknown_source"),
		Medium:   normalizeUTM(medium, "unknown_medium"),
	}
}

// parseRecordDate intenta parsear diferentes formatos de fecha
func parseRecordDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}

	// Formatos más comunes primero
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported date format: %s", dateStr)
}

// isRecordInDateRange verifica si el registro está dentro del rango de fechas
func isRecordInDateRange(recordDateStr string, filterDate *time.Time) bool {
	// Si no hay filtro de fecha, incluir todos los registros
	if filterDate == nil || filterDate.IsZero() {
		return true
	}

	recordDate, err := parseRecordDate(recordDateStr)
	if err != nil {
		log.Printf("Error parsing date %s: %v", recordDateStr, err)
		return false
	}

	// Incluir registros desde la fecha filtro inclusive
	return recordDate.After(*filterDate) || recordDate.Equal(*filterDate)
}
