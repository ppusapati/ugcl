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

// GET /api/v1/tasks
func GetAllTasks(w http.ResponseWriter, r *http.Request) {
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
	var items []models.Task
	if err := config.DB.
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		http.Error(w, "DB fetch error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var total int64
	if err := config.DB.
		Model(&models.Task{}).Count(&total).Error; err != nil {
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

// POST /api/v1/tasks
func CreateTask(w http.ResponseWriter, r *http.Request) {
	var item models.Task
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	user := middleware.GetUser(r)
	item.SiteEngineerName = user.Name
	item.SiteEngineerPhone = user.Phone
	config.DB.Create(&item)
	json.NewEncoder(w).Encode(item)
}

// GET /api/v1/tasks/{id}
func GetTask(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	var item models.Task
	if result := config.DB.First(&item, "id = ?", id); result.Error != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(item)
}

// PUT /api/v1/tasks/{id}
func UpdateTask(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	// id := params["id"]
	id, _ := strconv.Atoi(params["id"])
	var item models.Task
	if result := config.DB.First(&item, "id = ?", id); result.Error != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	// item.ID, _ = models.ParseUUID(id) // ensure correct id
	config.DB.Save(&item)
	json.NewEncoder(w).Encode(item)
}

// DELETE /api/v1/tasks/{id}
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	config.DB.Delete(&models.Task{}, "id = ?", id)
	w.WriteHeader(http.StatusNoContent)
}

// POST /api/v1/tasks/batch
func BatchTasks(w http.ResponseWriter, r *http.Request) {
	var batch []models.Task
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	user := middleware.GetUser(r)
	for i := range batch {
		batch[i].SiteEngineerName = user.Name
		batch[i].SiteEngineerPhone = user.Phone
	}
	// for i := range batch {
	// 	if user != nil && batch[i].WorkAssignedBy == nil {
	// 		assigned := user.Name
	// 		batch[i].WorkAssignedBy = &assigned
	// 	}
	// }
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
