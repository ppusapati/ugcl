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

// GetAllContractorReports godoc
// @Summary      Get all contractor reports
// @Description  Retrieves paginated list of contractor reports
// @Tags         contractor
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        page   query     int  false  "Page number"
// @Param        limit  query     int  false  "Items per page"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /api/v1/admin/contractor [get]
func GetAllContractorReports(w http.ResponseWriter, r *http.Request) {
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

	// Fetch data with pagination
	var items []models.Contractor
	if err := config.DB.
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		http.Error(w, "DB fetch error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get total count
	var total int64
	if err := config.DB.
		Model(&models.Contractor{}).
		Count(&total).Error; err != nil {
		http.Error(w, "DB count error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Build response
	response := map[string]interface{}{
		"total": total,
		"page":  page,
		"limit": limit,
		"data":  items,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateContractorReport godoc
// @Summary      Create contractor report
// @Description  Creates a new contractor report entry
// @Tags         contractor
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        contractor  body      models.Contractor  true  "Contractor Report Data"
// @Success      200         {object}  models.Contractor
// @Failure      400         {object}  map[string]string
// @Router       /api/v1/contractor [post]
func CreateContractorReport(w http.ResponseWriter, r *http.Request) {
	var item models.Contractor
	json.NewDecoder(r.Body).Decode(&item)
	user := middleware.GetUser(r)
	item.SiteEngineerName = user.Name
	item.SiteEngineerPhone = user.Phone
	config.DB.Create(&item)
	json.NewEncoder(w).Encode(item)
}

// GetContractorReport godoc
// @Summary      Get contractor report by ID
// @Description  Retrieves a contractor report by ID
// @Tags         contractor
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Contractor Report ID"
// @Success      200  {object}  models.Contractor
// @Failure      404  {object}  map[string]string
// @Router       /api/v1/admin/contractor/{id} [get]
func GetContractorReport(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var item models.Contractor
	config.DB.First(&item, id)
	json.NewEncoder(w).Encode(item)
}

// UpdateContractorReport godoc
// @Summary      Update contractor report
// @Description  Updates a contractor report by ID
// @Tags         contractor
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id          path      int                true  "Contractor Report ID"
// @Param        contractor  body      models.Contractor  true  "Updated contractor report data"
// @Success      200         {object}  models.Contractor
// @Failure      400         {object}  map[string]string
// @Router       /api/v1/admin/contractor/{id} [put]
func UpdateContractorReport(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	var item models.Contractor
	config.DB.First(&item, id)
	json.NewDecoder(r.Body).Decode(&item)
	config.DB.Save(&item)
	json.NewEncoder(w).Encode(item)
}

// DeleteContractorReport godoc
// @Summary      Delete contractor report
// @Description  Deletes a contractor report by ID
// @Tags         contractor
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path  int  true  "Contractor Report ID"
// @Success      204  {string}  string  "No Content"
// @Failure      400  {object}  map[string]string
// @Router       /api/v1/admin/contractor/{id} [delete]
func DeleteContractorReport(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	config.DB.Delete(&models.Contractor{}, id)
	w.WriteHeader(http.StatusNoContent)
}

// BatchContractors godoc
// @Summary      Batch create contractor reports
// @Description  Bulk creates contractor reports. Existing IDs are ignored.
// @Tags         contractor
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        batch  body      []models.Contractor  true  "Array of contractor reports"
// @Success      200    {string}  string  "OK"
// @Failure      400    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /api/v1/contractor/batch [post]
func BatchContractors(w http.ResponseWriter, r *http.Request) {
	var batch []models.Contractor
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	user := middleware.GetUser(r)
	for i := range batch {
		batch[i].SiteEngineerName = user.Name
		batch[i].SiteEngineerPhone = user.Phone
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
