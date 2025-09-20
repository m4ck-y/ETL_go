package application

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/m4ck-y/ETL_go/internal/domain/models"
)

func BuildUTMKey(campaign, source, medium string) models.UTMKey {
	return models.UTMKey{
		Campaign: strings.ToLower(strings.TrimSpace(campaign)),
		Source:   strings.ToLower(strings.TrimSpace(source)),
		Medium:   strings.ToLower(strings.TrimSpace(medium)),
	}
}

func RunETL(adsURL, crmURL string) (map[models.UTMKey]models.AggregatedMetrics, error) {
	ads, err := fetchAds(adsURL)
	if err != nil {
		return nil, err
	}

	crms, err := fetchCRM(crmURL)
	if err != nil {
		return nil, err
	}

	metrics := make(map[models.UTMKey]models.AggregatedMetrics)

	// Ads
	for _, ad := range ads {
		key := BuildUTMKey(ad.UTMCampaign, ad.UTMSource, ad.UTMMedium)
		m := metrics[key]
		m.Clicks += ad.Clicks
		m.Cost += ad.Cost
		metrics[key] = m
	}

	// CRM
	for _, crm := range crms {
		key := BuildUTMKey(crm.UTMCampaign, crm.UTMSource, crm.UTMMedium)
		m := metrics[key]
		if strings.ToLower(crm.Stage) == "lead" {
			m.Leads++
		}
		m.Opportunities++
		if strings.ToLower(crm.Stage) == "closed_won" {
			m.ClosedWon++
			m.Revenue += crm.Amount
		}
		metrics[key] = m
	}

	return metrics, nil
}

func fetchAds(url string) ([]models.AdRecord, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var parsed struct {
		External struct {
			Ads struct {
				Performance []models.AdRecord `json:"performance"`
			} `json:"ads"`
		} `json:"external"`
	}
	err = json.Unmarshal(body, &parsed)
	return parsed.External.Ads.Performance, err
}

func fetchCRM(url string) ([]models.CRMRecord, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var parsed struct {
		External struct {
			CRM struct {
				Opportunities []models.CRMRecord `json:"opportunities"`
			} `json:"crm"`
		} `json:"external"`
	}
	err = json.Unmarshal(body, &parsed)
	return parsed.External.CRM.Opportunities, err
}
