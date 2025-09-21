package application

import (
	"github.com/m4ck-y/ETL_go/internal/domain/models"
)

// safeDivide realiza división segura protegiendo contra división por cero
func safeDivide(numerator, denominator float64) float64 {
	if denominator == 0 || denominator == 0.0 {
		return 0.0
	}
	return numerator / denominator
}

// CalculateDerivedMetrics calcula las métricas derivadas requeridas
func CalculateDerivedMetrics(agg models.AggregatedMetrics) (cpc, cpa, cvrLeadToOpp, cvrOppToWon, roas float64) {
	// CPC = Cost Per Click = cost / clicks
	cpc = safeDivide(agg.Cost, float64(agg.Clicks))

	// CPA = Cost Per Acquisition = cost / leads
	cpa = safeDivide(agg.Cost, float64(agg.Leads))

	// CVR Lead to Opportunity = opportunities / leads
	cvrLeadToOpp = safeDivide(float64(agg.Opportunities), float64(agg.Leads))

	// CVR Opportunity to Won = won / opportunities
	cvrOppToWon = safeDivide(float64(agg.ClosedWon), float64(agg.Opportunities))

	// ROAS = Return on Ad Spend = revenue / cost
	roas = safeDivide(agg.Revenue, agg.Cost)

	return cpc, cpa, cvrLeadToOpp, cvrOppToWon, roas
}
