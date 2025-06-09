package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"gorm.io/gorm/clause"
	"p9e.in/ugcl/config"
	"p9e.in/ugcl/middleware"
	"p9e.in/ugcl/models"

	"github.com/gorilla/mux"
)

func GetAllNmrVehicle(w http.ResponseWriter, r *http.Request) {
	var items []models.Nmr_Vehicle
	config.DB.Find(&items)
	json.NewEncoder(w).Encode(items)
}

func CreateNmrVehicle(w http.ResponseWriter, r *http.Request) {
	var item models.Nmr_Vehicle
	json.NewDecoder(r.Body).Decode(&item)
	user := middleware.GetUser(r)
	item.AttendanceTakenBy = user.Name
	item.AttendancePhone = user.Phone
	config.DB.Create(&item)
	json.NewEncoder(w).Encode(item)
}

func GetNmrVehicle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var item models.Nmr_Vehicle
	config.DB.First(&item, id)
	json.NewEncoder(w).Encode(item)
}

func UpdateNmrVehicle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var item models.Nmr_Vehicle
	config.DB.First(&item, id)
	json.NewDecoder(r.Body).Decode(&item)
	config.DB.Save(&item)
	json.NewEncoder(w).Encode(item)
}

func DeleteNmrVehicle(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	config.DB.Delete(&models.Nmr_Vehicle{}, id)
	w.WriteHeader(http.StatusNoContent)
}

func BatchNmrVehicle(w http.ResponseWriter, r *http.Request) {
	var batch []models.Nmr_Vehicle
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	user := middleware.GetUser(r)
	for i := range batch {
		batch[i].AttendanceTakenBy = user.Name
		batch[i].AttendancePhone = user.Phone
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
