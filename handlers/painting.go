package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"p9e/ugcl/config"
	"p9e/ugcl/core/helper"
	"p9e/ugcl/core/services"
	"p9e/ugcl/middleware"
	"p9e/ugcl/models"

	"github.com/gorilla/mux"
	"gorm.io/gorm/clause"
)

func GetAllPaintingReports(w http.ResponseWriter, r *http.Request) {
	params, err := helper.ParseQueryParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := params.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	service := services.NewQueryService(config.DB, models.Painting{})
	response, err := service.GetQuery(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func CreatePaintingReport(w http.ResponseWriter, r *http.Request) {
	var item models.Painting
	json.NewDecoder(r.Body).Decode(&item)
	user := middleware.GetUser(r)
	item.SiteEngineerName = user.Name
	item.PhoneNumber = user.Phone
	config.DB.Create(&item)
	json.NewEncoder(w).Encode(item)
}

func GetPaintingReport(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var item models.Painting
	config.DB.First(&item, id)
	json.NewEncoder(w).Encode(item)
}

func UpdatePaintingReport(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var item models.Painting
	config.DB.First(&item, id)
	json.NewDecoder(r.Body).Decode(&item)
	config.DB.Save(&item)
	json.NewEncoder(w).Encode(item)
}

func DeletePaintingReport(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	config.DB.Delete(&models.Painting{}, id)
	w.WriteHeader(http.StatusNoContent)
}

func BatchPaintings(w http.ResponseWriter, r *http.Request) {
	var batch []models.Painting
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	user := middleware.GetUser(r)
	for i := range batch {
		batch[i].SiteEngineerName = user.Name
		batch[i].PhoneNumber = user.Phone
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
