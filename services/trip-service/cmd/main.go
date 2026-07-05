package main

import (
	"encoding/json"
	"log"
	"net/http"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/types"
)

func main() {

	inmemRepo := repository.NewInmemRepository()
	svc := service.NewService(inmemRepo)

	httpHanlder := httpHandler{
		svc: svc,
	}
	mux := http.NewServeMux()

	mux.HandleFunc("POST /preview", httpHanlder.handleTripPreview)

	server := &http.Server{
		Addr:    ":8083",
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Printf("http server err this is a test: %v", err)
	}
}

func (s *httpHandler) handleTripPreview(w http.ResponseWriter, r *http.Request) {

	var reqBody previewTripRequest

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if reqBody.UserID == "" {
		http.Error(w, "userID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	trip, err := s.svc.GetRoute(ctx, &reqBody.Pickup, &reqBody.Destination)
	if err != nil {
		http.Error(w, "failed to create trip", http.StatusInternalServerError)
		return
	}

	writeJson(w, http.StatusOK, trip)

}

func writeJson(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

type previewTripRequest struct {
	UserID      string           `json:"userID"`
	Pickup      types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

type httpHandler struct {
	svc domain.TripService
}
