package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	_ "p9e.in/ugcl/docs"
	"p9e.in/ugcl/handlers"
	"p9e.in/ugcl/middleware"
)

func RegisterRoutes() http.Handler {
	r := mux.NewRouter()
	// public
	r.HandleFunc("/api/v1/register", handlers.Register).Methods("POST")
	r.HandleFunc("/api/v1/login", handlers.Login).Methods("POST")
	r.HandleFunc("/api/v1/token", handlers.GetCurrentUser).Methods("GET")
	r.PathPrefix("/uploads/").Handler(
		http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))),
	)
	// a protected endpoint
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.SecurityMiddleware)
	api.Use(middleware.JWTMiddleware)
	// anyone logged in can hit this
	api.HandleFunc("/api/profile", func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value("userID").(string)
		role := r.Context().Value("role").(string)
		json.NewEncoder(w).Encode(map[string]string{"userID": id, "role": role})
	}).Methods("GET")

	// only admins:
	admin := api.PathPrefix("/admin").Subrouter()

	// admin.Use(middleware.SecurityMiddleware) // ⬅️ Enforce API key + IP
	admin.Use(func(h http.Handler) http.Handler {
		return middleware.RequireRole("admin", h) // ⬅️ Enforce admin role
	})

	admin.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("secret admin stats"))
	}).Methods("GET")

	admin.HandleFunc("/users", handlers.GetAllUsers).Methods("GET")
	admin.HandleFunc("/dprsite", handlers.GetAllSiteEngineerReports).Methods("GET")
	api.HandleFunc("/dprsite", handlers.CreateSiteEngineerReport).Methods("POST")
	admin.HandleFunc("/dprsite/{id}", handlers.GetSiteEngineerReport).Methods("GET")
	admin.HandleFunc("/dprsite/{id}", handlers.UpdateSiteEngineerReport).Methods("PUT")
	admin.HandleFunc("/dprsite/{id}", handlers.DeleteSiteEngineerReport).Methods("DELETE")
	api.HandleFunc("/dprsite/batch", handlers.BatchDprSites).Methods("POST")

	admin.HandleFunc("/wrapping", handlers.GetAllWrappingReports).Methods("GET")
	api.HandleFunc("/wrapping", handlers.CreateWrappingReport).Methods("POST")
	admin.HandleFunc("/wrapping/{id}", handlers.GetWrappingReport).Methods("GET")
	admin.HandleFunc("/wrapping/{id}", handlers.UpdateWrappingReport).Methods("PUT")
	admin.HandleFunc("/wrapping/{id}", handlers.DeleteWrappingReport).Methods("DELETE")
	api.HandleFunc("/wrapping/batch", handlers.BatchWrappings).Methods("POST")

	admin.HandleFunc("/eway", handlers.GetAllEways).Methods("GET")
	api.HandleFunc("/eway", handlers.CreateEway).Methods("POST")
	admin.HandleFunc("/eway/{id}", handlers.GetEway).Methods("GET")
	admin.HandleFunc("/eway/{id}", handlers.UpdateEway).Methods("PUT")
	admin.HandleFunc("/eway/{id}", handlers.DeleteEway).Methods("DELETE")
	api.HandleFunc("/eway/batch", handlers.BatchEwayss).Methods("POST")

	admin.HandleFunc("/water", handlers.GetAllWaterTankerReports).Methods("GET")
	api.HandleFunc("/water", handlers.CreateWaterTankerReport).Methods("POST")
	admin.HandleFunc("/water/{id}", handlers.GetWaterTankerReport).Methods("GET")
	admin.HandleFunc("/water/{id}", handlers.UpdateWaterTankerReport).Methods("PUT")
	admin.HandleFunc("/water/{id}", handlers.DeleteWaterTankerReport).Methods("DELETE")
	api.HandleFunc("/water/batch", handlers.BatchWaterReports).Methods("POST")

	admin.HandleFunc("/stock", handlers.GetAllStockReports).Methods("GET")
	api.HandleFunc("/stock", handlers.CreateStockReport).Methods("POST")
	admin.HandleFunc("/stock/{id}", handlers.GetStockReport).Methods("GET")
	admin.HandleFunc("/stock/{id}", handlers.UpdateStockReport).Methods("PUT")
	admin.HandleFunc("/stock/{id}", handlers.DeleteStockReport).Methods("DELETE")
	api.HandleFunc("/stock/batch", handlers.BatchStocks).Methods("POST")

	admin.HandleFunc("/dairysite", handlers.GetAllDairySiteReports).Methods("GET")
	api.HandleFunc("/dairysite", handlers.CreateDairySiteReport).Methods("POST")
	admin.HandleFunc("/dairysite/{id}", handlers.GetDairySiteReport).Methods("GET")
	admin.HandleFunc("/dairysite/{id}", handlers.UpdateDairySiteReport).Methods("PUT")
	admin.HandleFunc("/dairysite/{id}", handlers.DeleteDairySiteReport).Methods("DELETE")
	api.HandleFunc("/dairysite/batch", handlers.BatchDairySites).Methods("POST")
	api.HandleFunc("/dairysite/batch", handlers.BatchContractors).Methods("POST")

	admin.HandleFunc("/payment", handlers.GetAllPayments).Methods("GET")
	api.HandleFunc("/payment", handlers.CreatePayment).Methods("POST")
	admin.HandleFunc("/payment/{id}", handlers.GetPayment).Methods("GET")
	admin.HandleFunc("/payment/{id}", handlers.UpdatePayment).Methods("PUT")
	admin.HandleFunc("/payment/{id}", handlers.DeletePayment).Methods("DELETE")
	api.HandleFunc("/payment/batch", handlers.BatchPayments).Methods("POST")

	admin.HandleFunc("/material", handlers.GetAllMaterials).Methods("GET")
	api.HandleFunc("/material", handlers.CreateMaterial).Methods("POST")
	admin.HandleFunc("/material/{id}", handlers.GetMaterial).Methods("GET")
	admin.HandleFunc("/material/{id}", handlers.UpdateMaterial).Methods("PUT")
	admin.HandleFunc("/material/{id}", handlers.DeleteMaterial).Methods("DELETE")
	api.HandleFunc("/material/batch", handlers.BatchMaterials).Methods("POST")

	admin.HandleFunc("/mnr", handlers.GetAllMNRReports).Methods("GET")
	api.HandleFunc("/mnr", handlers.CreateMNRReport).Methods("POST")
	admin.HandleFunc("/mnr/{id}", handlers.GetMNRReport).Methods("GET")
	admin.HandleFunc("/mnr/{id}", handlers.UpdateMNRReport).Methods("PUT")
	admin.HandleFunc("/mnr/{id}", handlers.DeleteMNRReport).Methods("DELETE")
	api.HandleFunc("/mnr/batch", handlers.BatchMnrs).Methods("POST")

	admin.HandleFunc("/nmr_vehicle", handlers.GetAllNmrVehicle).Methods("GET")
	api.HandleFunc("/nmr_vehicle", handlers.CreateNmrVehicle).Methods("POST")
	admin.HandleFunc("/nmr_vehicle/{id}", handlers.GetNmrVehicle).Methods("GET")
	admin.HandleFunc("/nmr_vehicle/{id}", handlers.UpdateNmrVehicle).Methods("PUT")
	admin.HandleFunc("/nmr_vehicle/{id}", handlers.DeleteNmrVehicle).Methods("DELETE")
	api.HandleFunc("/nmr_vehicle/batch", handlers.BatchNmrVehicle).Methods("POST")

	admin.HandleFunc("/contractor", handlers.GetAllContractorReports).Methods("GET")
	api.HandleFunc("/contractor", handlers.CreateContractorReport).Methods("POST")
	admin.HandleFunc("/contractor/{id}", handlers.GetContractorReport).Methods("GET")
	admin.HandleFunc("/contractor/{id}", handlers.UpdateContractorReport).Methods("PUT")
	admin.HandleFunc("/contractor/{id}", handlers.DeleteContractorReport).Methods("DELETE")
	api.HandleFunc("/contractor/batch", handlers.BatchContractors).Methods("POST")

	admin.HandleFunc("/painting", handlers.GetAllPaintingReports).Methods("GET")
	api.HandleFunc("/painting", handlers.CreatePaintingReport).Methods("POST")
	admin.HandleFunc("/painting/{id}", handlers.GetPaintingReport).Methods("GET")
	admin.HandleFunc("/painting/{id}", handlers.UpdatePaintingReport).Methods("PUT")
	admin.HandleFunc("/painting/{id}", handlers.DeletePaintingReport).Methods("DELETE")
	api.HandleFunc("/painting/batch", handlers.BatchPaintings).Methods("POST")

	admin.HandleFunc("/diesel", handlers.GetAllDieselReports).Methods("GET")
	api.HandleFunc("/diesel", handlers.CreateDieselReport).Methods("POST")
	admin.HandleFunc("/diesel/{id}", handlers.GetDieselReport).Methods("GET")
	admin.HandleFunc("/diesel/{id}", handlers.UpdateDieselReport).Methods("PUT")
	admin.HandleFunc("/diesel/{id}", handlers.DeleteDieselReport).Methods("DELETE")
	api.HandleFunc("/diesel/batch", handlers.BatchDiesels).Methods("POST")

	admin.HandleFunc("/tasks", handlers.GetAllTasks).Methods("GET")
	api.HandleFunc("/tasks", handlers.CreateTask).Methods("POST")
	admin.HandleFunc("/tasks/{id}", handlers.GetTask).Methods("GET")
	admin.HandleFunc("/tasks/{id}", handlers.UpdateTask).Methods("PUT")
	admin.HandleFunc("/tasks/{id}", handlers.DeleteTask).Methods("DELETE")
	api.HandleFunc("/tasks/batch", handlers.BatchTasks).Methods("POST")

	admin.HandleFunc("/vehiclelog", handlers.GetAllVehicleLogs).Methods("GET")
	api.HandleFunc("/vehiclelog", handlers.CreateVehicleLog).Methods("POST")
	admin.HandleFunc("/vehiclelog/{id}", handlers.GetVehicleLog).Methods("GET")
	admin.HandleFunc("/vehiclelog/{id}", handlers.UpdateVehicleLog).Methods("PUT")
	admin.HandleFunc("/vehiclelog/{id}", handlers.DeleteVehicleLog).Methods("DELETE")
	api.HandleFunc("/vehiclelog/batch", handlers.BatchVehicleLogs).Methods("POST")

	api.HandleFunc("/files/upload", handlers.UploadFile).Methods("POST")

	partner := r.PathPrefix("/api/v1/partner").Subrouter()
	partner.Use(middleware.SecurityMiddleware) // API key + IP
	partner.HandleFunc("/dprsite", handlers.GetAllSiteEngineerReports).Methods("GET")
	partner.HandleFunc("/dprsite/{id}", handlers.GetSiteEngineerReport).Methods("GET")
	partner.HandleFunc("/wrapping", handlers.GetAllWrappingReports).Methods("GET")
	partner.HandleFunc("/wrapping/{id}", handlers.GetWrappingReport).Methods("GET")
	partner.HandleFunc("/eway", handlers.GetAllEways).Methods("GET")
	partner.HandleFunc("/eway/{id}", handlers.GetEway).Methods("GET")
	partner.HandleFunc("/water", handlers.GetAllWaterTankerReports).Methods("GET")
	partner.HandleFunc("/water/{id}", handlers.GetWaterTankerReport).Methods("GET")
	partner.HandleFunc("/stock", handlers.GetAllStockReports).Methods("GET")
	partner.HandleFunc("/stock/{id}", handlers.GetStockReport).Methods("GET")
	partner.HandleFunc("/dairysite", handlers.GetAllDairySiteReports).Methods("GET")
	partner.HandleFunc("/dairysite/{id}", handlers.GetDairySiteReport).Methods("GET")
	partner.HandleFunc("/payment", handlers.GetAllPayments).Methods("GET")
	partner.HandleFunc("/payment/{id}", handlers.GetPayment).Methods("GET")
	partner.HandleFunc("/material", handlers.GetAllMaterials).Methods("GET")
	partner.HandleFunc("/material/{id}", handlers.GetMaterial).Methods("GET")
	partner.HandleFunc("/mnr", handlers.GetAllMNRReports).Methods("GET")
	partner.HandleFunc("/mnr/{id}", handlers.GetMNRReport).Methods("GET")
	partner.HandleFunc("/nmr_vehicle", handlers.GetAllNmrVehicle).Methods("GET")
	partner.HandleFunc("/nmr_vehicle/{id}", handlers.GetNmrVehicle).Methods("GET")
	partner.HandleFunc("/contractor", handlers.GetAllContractorReports).Methods("GET")
	partner.HandleFunc("/contractor/{id}", handlers.GetContractorReport).Methods("GET")
	partner.HandleFunc("/painting", handlers.GetAllPaintingReports).Methods("GET")
	partner.HandleFunc("/painting/{id}", handlers.GetPaintingReport).Methods("GET")
	partner.HandleFunc("/diesel", handlers.GetAllDieselReports).Methods("GET")
	partner.HandleFunc("/diesel/{id}", handlers.GetDieselReport).Methods("GET")
	partner.HandleFunc("/tasks", handlers.GetAllTasks).Methods("GET")
	partner.HandleFunc("/tasks/{id}", handlers.GetTask).Methods("GET")
	partner.HandleFunc("/vehiclelog", handlers.GetAllVehicleLogs).Methods("GET")
	partner.HandleFunc("/vehiclelog/{id}", handlers.GetVehicleLog).Methods("GET")

	return r
}
