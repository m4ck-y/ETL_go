package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/m4ck-y/ETL_go/internal/domain/models"
)

// Configuración para reintentos HTTP
type retryConfig struct {
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

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

// retryHTTPRequest realiza petición HTTP con reintentos y backoff exponencial
func retryHTTPRequest(url string, config retryConfig) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= config.maxRetries; attempt++ {
		// Crear contexto con timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		// Crear petición HTTP
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("error creating request: %w", err)
		}

		// Ejecutar petición
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		cancel() // Liberar contexto

		// Verificar respuesta exitosa
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		// Manejar error
		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			resp.Body.Close()
		}

		// Calcular delay para siguiente intento
		if attempt < config.maxRetries {
			delay := time.Duration(float64(config.baseDelay) * math.Pow(2, float64(attempt)))
			if delay > config.maxDelay {
				delay = config.maxDelay
			}

			log.Printf("Intento %d/%d falló, reintentando en %v: %v",
				attempt+1, config.maxRetries+1, delay, lastErr)
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("falló después de %d intentos: %w", config.maxRetries+1, lastErr)
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
	log.Printf("Iniciando proceso ETL con filtro de fecha: %v", sinceDate)

	// Obtener datos de ambas fuentes
	ads, err := fetchAds(adsURL, sinceDate)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo datos de ads: %w", err)
	}

	crms, err := fetchCRM(crmURL, sinceDate)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo datos de crm: %w", err)
	}

	// Inicializar mapa de métricas
	metrics := make(map[models.UTMKey]models.AggregatedMetrics)

	// Procesar métricas de ads y crm
	processAdsMetrics(ads, sinceDate, metrics)
	processCRMMetrics(crms, sinceDate, metrics)

	log.Printf("ETL completado exitosamente. Procesados %d registros de ads y %d de crm",
		len(ads), len(crms))

	return metrics, nil
}

// fetchData realiza una petición HTTP con reintentos y parsea la respuesta JSON
func fetchData(url string, target interface{}, dataType string) error {
	config := retryConfig{
		maxRetries: 3,
		baseDelay:  1 * time.Second,
		maxDelay:   10 * time.Second,
	}

	resp, err := retryHTTPRequest(url, config)
	if err != nil {
		return fmt.Errorf("failed to fetch %s data: %w", dataType, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read %s response body: %w", dataType, err)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to parse %s JSON: %w", dataType, err)
	}

	return nil
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
	log.Printf("Fetched %d ads records from %s", len(records), url)
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
	log.Printf("Fetched %d crm records from %s", len(records), url)
	return records, nil
}
