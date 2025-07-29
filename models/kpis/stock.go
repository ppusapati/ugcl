// models/stock.go (your Stock struct already defined)
package kpis

// DTO for KPI response
type StockKPIs struct {
	TotalStockIn               int     `json:"totalStockIn"`
	TotalStockOut              int     `json:"totalStockOut"`
	CurrentStockLevel          int     `json:"currentStockLevel"`
	StockAgingDays             float64 `json:"stockAgingDays"`
	DefectiveMaterialPct       float64 `json:"defectiveMaterialPct"`
	TopContractors             []KVP   `json:"topContractors"`
	TopItemsPipeDiaUsed        []KVP   `json:"topItemsPipeDiaUsed"`
	DocumentationCompliancePct float64 `json:"documentationCompliancePct"`
	SpecialsVsRegularRatio     float64 `json:"specialsVsRegularRatio"`
	AvgDataEntryDelayDays      float64 `json:"avgDataEntryDelayDays"`
}

type KVP struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
	Extra float64 `json:"extra,omitempty"`
}
