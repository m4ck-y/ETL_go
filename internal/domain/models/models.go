package models

type AdRecord struct {
	Date        string  `json:"date"`
	CampaignID  string  `json:"campaign_id"`
	Channel     string  `json:"channel"`
	Clicks      int     `json:"clicks"`
	Impressions int     `json:"impressions"`
	Cost        float64 `json:"cost"`
	UTMCampaign string  `json:"utm_campaign"`
	UTMSource   string  `json:"utm_source"`
	UTMMedium   string  `json:"utm_medium"`
}

type CRMRecord struct {
	OpportunityID string  `json:"opportunity_id"`
	ContactEmail  string  `json:"contact_email"`
	Stage         string  `json:"stage"`
	Amount        float64 `json:"amount"`
	CreatedAt     string  `json:"created_at"`
	UTMCampaign   string  `json:"utm_campaign"`
	UTMSource     string  `json:"utm_source"`
	UTMMedium     string  `json:"utm_medium"`
}

type UTMKey struct {
	Campaign string
	Source   string
	Medium   string
}

type AggregatedMetrics struct {
	Clicks        int
	Cost          float64
	Leads         int
	Opportunities int
	ClosedWon     int
	Revenue       float64
}
