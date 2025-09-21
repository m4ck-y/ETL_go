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

type MetricResponse struct {
	UTMCampaign   string  `json:"utm_campaign"`
	UTMSource     string  `json:"utm_source"`
	UTMMedium     string  `json:"utm_medium"`
	Clicks        int     `json:"clicks"`
	Cost          float64 `json:"cost"`
	Leads         int     `json:"leads"`
	Opportunities int     `json:"opportunities"`
	ClosedWon     int     `json:"closed_won"`
	Revenue       float64 `json:"revenue"`
	// Métricas adicionales calculadas automáticamente a partir de los datos principales
	CPC          float64 `json:"cpc"`             // Cost por click = cost / clicks
	CPA          float64 `json:"cpa"`             // Cost por adquisición = cost / leads
	CVRLeadToOpp float64 `json:"cvr_lead_to_opp"` // Tasa de conversión de Lead a Opportunity
	CVROppToWon  float64 `json:"cvr_opp_to_won"`  // Tasa de conversión de Opportunity a ClosedWon
	ROAS         float64 `json:"roas"`            // Retorno de inversión publicitaria = revenue / cost
}
