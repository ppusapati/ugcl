package kpis

type DairyKPI struct {
	TotalReports           int            `json:"totalReports"`
	ReportsPerSite         map[string]int `json:"reportsPerSite"`
	ReportsPerDay          map[string]int `json:"reportsPerDay"`
	ReportsPerEngineer     map[string]int `json:"reportsPerEngineer"`
	UniqueSites            int            `json:"uniqueSites"`
	UniqueEngineers        int            `json:"uniqueEngineers"`
	WorkLogs               []WorkLogEntry `json:"workLogs"`
	GeoPoints              []GeoPoint     `json:"geoPoints"`
	ReportingCompliancePct float64        `json:"reportingCompliancePct"`
}

type WorkLogEntry struct {
	NameOfSite string `json:"nameOfSite"`
	Date       string `json:"date"`
	TodaysWork string `json:"todaysWork"`
}

type GeoPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
