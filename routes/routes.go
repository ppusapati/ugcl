package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"p9e.in/ugcl/handlers"
	"p9e.in/ugcl/middleware"
)

func RegisterRoutes() http.Handler {
	r := mux.NewRouter()
	// public
	r.HandleFunc("/api/v1/register", handlers.Register).Methods("POST")
	r.HandleFunc("/api/v1/login", handlers.Login).Methods("POST")
	r.HandleFunc("/api/v1/token", handlers.GetCurrentUser).Methods("GET")
	// a protected endpoint
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.JWTMiddleware)

	// anyone logged in can hit this
	api.HandleFunc("/api/profile", func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value("userID").(string)
		role := r.Context().Value("role").(string)
		json.NewEncoder(w).Encode(map[string]string{"userID": id, "role": role})
	}).Methods("GET")

	// only admins:
	admin := api.PathPrefix("/admin").Subrouter()
	admin.Use(func(h http.Handler) http.Handler {
		return middleware.RequireRole("admin", h)
	})
	admin.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("secret admin stats"))
	}).Methods("GET")

	admin.HandleFunc("/users", handlers.GetAllUsers).Methods("GET")
	api.HandleFunc("/dprsite", handlers.GetAllSiteEngineerReports).Methods("GET")
	api.HandleFunc("/dprsite", handlers.CreateSiteEngineerReport).Methods("POST")
	api.HandleFunc("/dprsite/{id}", handlers.GetSiteEngineerReport).Methods("GET")
	api.HandleFunc("/dprsite/{id}", handlers.UpdateSiteEngineerReport).Methods("PUT")
	api.HandleFunc("/dprsite/{id}", handlers.DeleteSiteEngineerReport).Methods("DELETE")
	api.HandleFunc("/dprsite/batch", handlers.BatchDprSites).Methods("POST")

	api.HandleFunc("/wrapping", handlers.GetAllWrappingReports).Methods("GET")
	api.HandleFunc("/wrapping", handlers.CreateWrappingReport).Methods("POST")
	api.HandleFunc("/wrapping/{id}", handlers.GetWrappingReport).Methods("GET")
	api.HandleFunc("/wrapping/{id}", handlers.UpdateWrappingReport).Methods("PUT")
	api.HandleFunc("/wrapping/{id}", handlers.DeleteWrappingReport).Methods("DELETE")
	api.HandleFunc("/wrapping/batch", handlers.BatchWrappings).Methods("POST")

	api.HandleFunc("/eway", handlers.GetAllEways).Methods("GET")
	api.HandleFunc("/eway", handlers.CreateEway).Methods("POST")
	api.HandleFunc("/eway/{id}", handlers.GetEway).Methods("GET")
	api.HandleFunc("/eway/{id}", handlers.UpdateEway).Methods("PUT")
	api.HandleFunc("/eway/{id}", handlers.DeleteEway).Methods("DELETE")
	api.HandleFunc("/eway/batch", handlers.BatchEwayss).Methods("POST")

	api.HandleFunc("/water", handlers.GetAllWaterTankerReports).Methods("GET")
	api.HandleFunc("/water", handlers.CreateWaterTankerReport).Methods("POST")
	api.HandleFunc("/water/{id}", handlers.GetWaterTankerReport).Methods("GET")
	api.HandleFunc("/water/{id}", handlers.UpdateWaterTankerReport).Methods("PUT")
	api.HandleFunc("/water/{id}", handlers.DeleteWaterTankerReport).Methods("DELETE")
	api.HandleFunc("/water/batch", handlers.BatchWaterReports).Methods("POST")

	api.HandleFunc("/stock", handlers.GetAllStockReports).Methods("GET")
	api.HandleFunc("/stock", handlers.CreateStockReport).Methods("POST")
	api.HandleFunc("/stock/{id}", handlers.GetStockReport).Methods("GET")
	api.HandleFunc("/stock/{id}", handlers.UpdateStockReport).Methods("PUT")
	api.HandleFunc("/stock/{id}", handlers.DeleteStockReport).Methods("DELETE")
	api.HandleFunc("/stock/batch", handlers.BatchStocks).Methods("POST")

	api.HandleFunc("/dairysite", handlers.GetAllDairySiteReports).Methods("GET")
	api.HandleFunc("/dairysite", handlers.CreateDairySiteReport).Methods("POST")
	api.HandleFunc("/dairysite/{id}", handlers.GetDairySiteReport).Methods("GET")
	api.HandleFunc("/dairysite/{id}", handlers.UpdateDairySiteReport).Methods("PUT")
	api.HandleFunc("/dairysite/{id}", handlers.DeleteDairySiteReport).Methods("DELETE")
	api.HandleFunc("/dairysite/batch", handlers.BatchDairySites).Methods("POST")
	api.HandleFunc("/dairysite/batch", handlers.BatchContractors).Methods("POST")

	api.HandleFunc("/payment", handlers.GetAllPayments).Methods("GET")
	api.HandleFunc("/payment", handlers.CreatePayment).Methods("POST")
	api.HandleFunc("/payment/{id}", handlers.GetPayment).Methods("GET")
	api.HandleFunc("/payment/{id}", handlers.UpdatePayment).Methods("PUT")
	api.HandleFunc("/payment/{id}", handlers.DeletePayment).Methods("DELETE")
	api.HandleFunc("/payment/batch", handlers.BatchPayments).Methods("POST")

	api.HandleFunc("/material", handlers.GetAllMaterials).Methods("GET")
	api.HandleFunc("/material", handlers.CreateMaterial).Methods("POST")
	api.HandleFunc("/material/{id}", handlers.GetMaterial).Methods("GET")
	api.HandleFunc("/material/{id}", handlers.UpdateMaterial).Methods("PUT")
	api.HandleFunc("/material/{id}", handlers.DeleteMaterial).Methods("DELETE")
	api.HandleFunc("/material/batch", handlers.BatchMaterials).Methods("POST")

	api.HandleFunc("/mnr", handlers.GetAllMNRReports).Methods("GET")
	api.HandleFunc("/mnr", handlers.CreateMNRReport).Methods("POST")
	api.HandleFunc("/mnr/{id}", handlers.GetMNRReport).Methods("GET")
	api.HandleFunc("/mnr/{id}", handlers.UpdateMNRReport).Methods("PUT")
	api.HandleFunc("/mnr/{id}", handlers.DeleteMNRReport).Methods("DELETE")
	api.HandleFunc("/mnr/batch", handlers.BatchMnrs).Methods("POST")

	api.HandleFunc("/contractor", handlers.GetAllContractorReports).Methods("GET")
	api.HandleFunc("/contractor", handlers.CreateContractorReport).Methods("POST")
	api.HandleFunc("/contractor/{id}", handlers.GetContractorReport).Methods("GET")
	api.HandleFunc("/contractor/{id}", handlers.UpdateContractorReport).Methods("PUT")
	api.HandleFunc("/contractor/{id}", handlers.DeleteContractorReport).Methods("DELETE")
	api.HandleFunc("/contractor/batch", handlers.BatchContractors).Methods("POST")

	api.HandleFunc("/painting", handlers.GetAllPaintingReports).Methods("GET")
	api.HandleFunc("/painting", handlers.CreatePaintingReport).Methods("POST")
	api.HandleFunc("/painting/{id}", handlers.GetPaintingReport).Methods("GET")
	api.HandleFunc("/painting/{id}", handlers.UpdatePaintingReport).Methods("PUT")
	api.HandleFunc("/painting/{id}", handlers.DeletePaintingReport).Methods("DELETE")
	api.HandleFunc("/painting/batch", handlers.BatchPaintings).Methods("POST")

	api.HandleFunc("/diesel", handlers.GetAllDieselReports).Methods("GET")
	api.HandleFunc("/diesel", handlers.CreateDieselReport).Methods("POST")
	api.HandleFunc("/diesel/{id}", handlers.GetDieselReport).Methods("GET")
	api.HandleFunc("/diesel/{id}", handlers.UpdateDieselReport).Methods("PUT")
	api.HandleFunc("/diesel/{id}", handlers.DeleteDieselReport).Methods("DELETE")
	api.HandleFunc("/diesel/batch", handlers.BatchDiesels).Methods("POST")

	api.HandleFunc("/tasks", handlers.GetAllTasks).Methods("GET")
	api.HandleFunc("/tasks", handlers.CreateTask).Methods("POST")
	api.HandleFunc("/tasks/{id}", handlers.GetTask).Methods("GET")
	api.HandleFunc("/tasks/{id}", handlers.UpdateTask).Methods("PUT")
	api.HandleFunc("/tasks/{id}", handlers.DeleteTask).Methods("DELETE")
	api.HandleFunc("/tasks/batch", handlers.BatchTasks).Methods("POST")

	api.HandleFunc("/files/upload", handlers.UploadFile).Methods("POST")
	return r
}
