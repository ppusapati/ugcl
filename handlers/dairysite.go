package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"gorm.io/gorm/clause"
	"p9e.in/ugcl/config"
	"p9e.in/ugcl/middleware"
	"p9e.in/ugcl/models"

	"github.com/gorilla/mux"
)

func GetAllDairySiteReports(w http.ResponseWriter, r *http.Request) {
	var items []models.DairySite
	config.DB.Find(&items)
	json.NewEncoder(w).Encode(items)
}

func CreateDairySiteReport(w http.ResponseWriter, r *http.Request) {
	var item models.DairySite
	json.NewDecoder(r.Body).Decode(&item)
	user := middleware.GetUser(r)
	item.SiteEngineerName = user.Name
	item.SiteEngineerPhone = user.Phone
	config.DB.Create(&item)
	json.NewEncoder(w).Encode(item)
}

func GetDairySiteReport(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var item models.DairySite
	config.DB.First(&item, id)
	json.NewEncoder(w).Encode(item)
}

func UpdateDairySiteReport(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var item models.DairySite
	config.DB.First(&item, id)
	json.NewDecoder(r.Body).Decode(&item)
	config.DB.Save(&item)
	json.NewEncoder(w).Encode(item)
}

func DeleteDairySiteReport(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	config.DB.Delete(&models.DairySite{}, id)
	w.WriteHeader(http.StatusNoContent)
}

// BatchContractorReports handles POST /api/v1/contractor/batch
func BatchDairySites(w http.ResponseWriter, r *http.Request) {
	var batch []models.DairySite
	// fmt.Println(json.NewDecoder(r.Body).Decode(&batch))
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	// who’s submitting?
	user := middleware.GetUser(r)
	for i := range batch {
		batch[i].SiteEngineerName = user.Name
		batch[i].SiteEngineerPhone = user.Phone
	}
	fmt.Println(&batch)
	if err := config.DB.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: true,
		}).
		Create(&batch).Error; err != nil {
		fmt.Println(err.Error())
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
