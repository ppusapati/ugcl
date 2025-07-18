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

func GetAllStockReports(w http.ResponseWriter, r *http.Request) {
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
	var items []models.Stock
	if err := config.DB.
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		http.Error(w, "DB fetch error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var total int64
	if err := config.DB.
		Model(&models.Stock{}).Count(&total).Error; err != nil {
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

func CreateStockReport(w http.ResponseWriter, r *http.Request) {
	var item models.Stock
	json.NewDecoder(r.Body).Decode(&item)
	user := middleware.GetUser(r)
	item.YardInchargeName = user.Name
	item.YardInchargePhone = user.Phone
	config.DB.Create(&item)
	json.NewEncoder(w).Encode(item)
}

func GetStockReport(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var item models.Stock
	config.DB.First(&item, id)
	json.NewEncoder(w).Encode(item)
}

func UpdateStockReport(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var item models.Stock
	config.DB.First(&item, id)
	json.NewDecoder(r.Body).Decode(&item)
	config.DB.Save(&item)
	json.NewEncoder(w).Encode(item)
}

func DeleteStockReport(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	config.DB.Delete(&models.Stock{}, id)
	w.WriteHeader(http.StatusNoContent)
}

func BatchStocks(w http.ResponseWriter, r *http.Request) {
	var batch []models.Stock
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	user := middleware.GetUser(r)
	for i := range batch {
		batch[i].YardInchargeName = user.Name
		batch[i].YardInchargePhone = user.Phone
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
