package kpis

type DieselKPIs struct {
	TotalDieselConsumed     float64     `json:"totalDieselConsumed"`
	TotalAmountPaid         float64     `json:"totalAmountPaid"`
	AvgDieselPerLiter       float64     `json:"avgDieselPerLiter"`
	DieselByContractor      []KVP       `json:"dieselByContractor"` // sum by contractor
	DieselByVehicle         []KVP       `json:"dieselByVehicle"`    // sum by vehicle
	CardNumberUsage         []KVP       `json:"cardNumberUsage"`    // sum diesel/amount by card
	EntriesWithPhotosPct    float64     `json:"entriesWithPhotosPct"`
	GeoPoints               [][]float64 `json:"geoPoints"` // [lat, lng] pairs
	EntriesWithRemarks      int         `json:"entriesWithRemarks"`
	EntriesSubmittedPerSite []KVP       `json:"entriesPerSite"` // count by site
	EntriesSubmittedPerDate []KVP       `json:"entriesPerDate"` // count by date
	TotalEntries            int         `json:"totalEntries"`
}
