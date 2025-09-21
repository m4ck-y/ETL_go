package application

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/m4ck-y/ETL_go/internal/domain/models"
)

// Logger local para el package application
var etlLogger = &Logger{
	serviceName: "etl-go-service",
	minLevel:    INFO,
}

// Logger simple para evitar dependencias circulares
type Logger struct {
	serviceName string
	minLevel    LogLevel
}

type LogLevel string

const (
	INFO  LogLevel = "INFO"
	ERROR LogLevel = "ERROR"
	WARN  LogLevel = "WARN"
)

func (l *Logger) Info(message string, requestID string, extra map[string]interface{}) {
	l.log(INFO, message, requestID, extra)
}

func (l *Logger) Error(message string, requestID string, extra map[string]interface{}) {
	l.log(ERROR, message, requestID, extra)
}

func (l *Logger) Warn(message string, requestID string, extra map[string]interface{}) {
	l.log(WARN, message, requestID, extra)
}

func (l *Logger) log(level LogLevel, message string, requestID string, extra map[string]interface{}) {
	// Salida simple para evitar complejidad
	log.Printf("[%s] %s - %s (request_id: %s)", level, l.serviceName, message, requestID)
	if extra != nil && len(extra) > 0 {
		log.Printf("Extra: %+v", extra)
	}
}

// processAdsMetrics procesa métricas de anuncios
func processAdsMetrics(ads []models.AdRecord, sinceDate *time.Time, metrics map[models.UTMKey]models.AggregatedMetrics) {
	for _, ad := range ads {
		if !isRecordInDateRange(ad.Date, sinceDate) {
			continue
		}

		key := BuildUTMKey(ad.UTMCampaign, ad.UTMSource, ad.UTMMedium)
		m := metrics[key]
		m.Clicks += ad.Clicks
		m.Cost += ad.Cost
		metrics[key] = m
	}
}

// processCRMMetrics procesa métricas de CRM
func processCRMMetrics(crms []models.CRMRecord, sinceDate *time.Time, metrics map[models.UTMKey]models.AggregatedMetrics) {
	for _, crm := range crms {
		if !isRecordInDateRange(crm.CreatedAt, sinceDate) {
			continue
		}

		key := BuildUTMKey(crm.UTMCampaign, crm.UTMSource, crm.UTMMedium)
		m := metrics[key]

		// Contar por stage
		stage := strings.ToLower(crm.Stage)
		switch stage {
		case "lead":
			m.Leads++
		case "closed_won":
			m.ClosedWon++
			m.Revenue += crm.Amount
		}

		m.Opportunities++
		metrics[key] = m
	}
}

func RunETL(adsURL, crmURL string, sinceDate *time.Time) (map[models.UTMKey]models.AggregatedMetrics, error) {
	// Logging estructurado del inicio del proceso
	etlLogger.Info("Iniciando proceso ETL", "system", map[string]interface{}{
		"ads_url":    adsURL,
		"crm_url":    crmURL,
		"since_date": sinceDate,
	})

	// Obtener datos de ambas fuentes
	ads, err := fetchAds(adsURL, sinceDate)
	if err != nil {
		etlLogger.Error("Error obteniendo datos de ads", "system", map[string]interface{}{
			"ads_url": adsURL,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("error obteniendo datos de ads: %w", err)
	}

	crms, err := fetchCRM(crmURL, sinceDate)
	if err != nil {
		etlLogger.Error("Error obteniendo datos de crm", "system", map[string]interface{}{
			"crm_url": crmURL,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("error obteniendo datos de crm: %w", err)
	}

	// Inicializar mapa de métricas
	metrics := make(map[models.UTMKey]models.AggregatedMetrics)

	// Procesar métricas de ads y crm
	processAdsMetrics(ads, sinceDate, metrics)
	processCRMMetrics(crms, sinceDate, metrics)

	// Logging estructurado del resultado
	etlLogger.Info("ETL completado exitosamente", "system", map[string]interface{}{
		"ads_records":        len(ads),
		"crm_records":        len(crms),
		"total_combinations": len(metrics),
	})

	return metrics, nil
}

func fetchAds(url string, sinceDate *time.Time) ([]models.AdRecord, error) {
	var response struct {
		External struct {
			Ads struct {
				Performance []models.AdRecord `json:"performance"`
			} `json:"ads"`
		} `json:"external"`
	}

	if err := fetchData(url, &response, "ads"); err != nil {
		return nil, err
	}

	records := response.External.Ads.Performance
	etlLogger.Info("Datos de ads obtenidos", "system", map[string]interface{}{
		"url":     url,
		"records": len(records),
	})
	return records, nil
}

func fetchCRM(url string, sinceDate *time.Time) ([]models.CRMRecord, error) {
	var response struct {
		External struct {
			CRM struct {
				Opportunities []models.CRMRecord `json:"opportunities"`
			} `json:"crm"`
		} `json:"external"`
	}

	if err := fetchData(url, &response, "crm"); err != nil {
		return nil, err
	}

	records := response.External.CRM.Opportunities
	etlLogger.Info("Datos de crm obtenidos", "system", map[string]interface{}{
		"url":     url,
		"records": len(records),
	})
	return records, nil
}
