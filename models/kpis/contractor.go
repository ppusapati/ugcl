// models/kpis.go or similar
package kpis

type KeyValue struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
}

type ContractorKPIs struct {
	TotalMetersCompleted  float64      `json:"totalMetersCompleted"`
	TotalDieselTaken      float64      `json:"totalDieselTaken"`
	DieselPerMeter        float64      `json:"dieselPerMeter"`
	AverageMetersPerDay   float64      `json:"averageMetersPerDay"`
	ReportsWithPhotosPct  float64      `json:"reportsWithPhotosPct"`
	VehicleUtilization    []KeyValue   `json:"vehicleUtilization"` // [{key: type, value: count}]
	AverageWorkingHours   float64      `json:"averageWorkingHours"`
	CardNumberDieselDrawn []KeyValue   `json:"cardNumberDieselDrawn"` // [{key: cardNumber, value: sum}]
	GeoLocations          [][2]float64 `json:"geoLocations"`          // [[lat, lng], ...]
	ReportsByDateSite     []KeyValue   `json:"reportsByDateSite"`     // [{key: "2024-07-25|SiteA", value: count}]
	// You can add more groupings as needed
}
