package application

import (
	"fmt"
	"strings"
	"time"

	"github.com/m4ck-y/ETL_go/internal/domain/models"
	"github.com/m4ck-y/ETL_go/internal/pkg/logger"
)

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

func processCRMMetrics(crms []models.CRMRecord, sinceDate *time.Time, metrics map[models.UTMKey]models.AggregatedMetrics) {
	for _, crm := range crms {
		if !isRecordInDateRange(crm.CreatedAt, sinceDate) {
			continue
		}

		key := BuildUTMKey(crm.UTMCampaign, crm.UTMSource, crm.UTMMedium)
		m := metrics[key]

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
	logger.GlobalLogger.Info("Iniciando proceso ETL", "system", map[string]interface{}{
		"ads_url":    adsURL,
		"crm_url":    crmURL,
		"since_date": sinceDate,
	})

	ads, err := fetchAds(adsURL, sinceDate)
	if err != nil {
		logger.GlobalLogger.Error("Error obteniendo datos de ads", "system", map[string]interface{}{
			"ads_url": adsURL,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("error obteniendo datos de ads: %w", err)
	}

	crms, err := fetchCRM(crmURL, sinceDate)
	if err != nil {
		logger.GlobalLogger.Error("Error obteniendo datos de crm", "system", map[string]interface{}{
			"crm_url": crmURL,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("error obteniendo datos de crm: %w", err)
	}

	metrics := make(map[models.UTMKey]models.AggregatedMetrics)

	processAdsMetrics(ads, sinceDate, metrics)
	processCRMMetrics(crms, sinceDate, metrics)

	logger.GlobalLogger.Info("ETL completado exitosamente", "system", map[string]interface{}{
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
	return records, nil
}
