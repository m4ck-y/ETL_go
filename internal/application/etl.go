package application

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/m4ck-y/ETL_go/internal/domain/models"
)

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
