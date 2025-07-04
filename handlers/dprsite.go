package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/gorm/clause"
	"p9e.in/ugcl/config"
	"p9e.in/ugcl/middleware"
	"p9e.in/ugcl/models"
)

func GetAllSiteEngineerReports(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 10

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}
	offset := (page - 1) * limit
	var items []models.DprSite
	if err := config.DB.
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		http.Error(w, "DB fetch error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var total int64
	if err := config.DB.
		Model(&models.DprSite{}).
		Count(&total).Error; err != nil {
		http.Error(w, "DB count error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"total": total,
		"page":  page,
		"limit": limit,
		"data":  items,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CreateSiteEngineerReport(w http.ResponseWriter, r *http.Request) {
	var report models.DprSite
	json.NewDecoder(r.Body).Decode(&report)
	user := middleware.GetUser(r)
	report.InformationEnteredBy = user.Name
	report.PhoneNumberOfInformationEnteredPerson = user.Phone

	config.DB.Create(&report)
	json.NewEncoder(w).Encode(report)
}

func GetSiteEngineerReport(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var report models.DprSite
	config.DB.First(&report, id)
	json.NewEncoder(w).Encode(report)
}

func UpdateSiteEngineerReport(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var report models.DprSite
	config.DB.First(&report, id)
	json.NewDecoder(r.Body).Decode(&report)
	config.DB.Save(&report)
	json.NewEncoder(w).Encode(report)
}

func DeleteSiteEngineerReport(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	config.DB.Delete(&models.DprSite{}, id)
	w.WriteHeader(http.StatusNoContent)
}

// BatchContractorReports handles POST /api/v1/contractor/batch
func BatchDprSites(w http.ResponseWriter, r *http.Request) {
	var batch []models.DprSite
	// fmt.Println(json.NewDecoder(r.Body).Decode(&batch))
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	user := middleware.GetUser(r)
	for i := range batch {
		batch[i].InformationEnteredBy = user.Name
		batch[i].PhoneNumberOfInformationEnteredPerson = user.Phone
	}
	if err := config.DB.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: true,
		}).
		Create(&batch).Error; err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
