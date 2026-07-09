package main

import (
	"encoding/json"
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_client"
	"ride-sharing/shared/contracts"
)

func handleTripPreview(w http.ResponseWriter, r *http.Request) {

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

	tripService, err := grpc_client.NewTripServiceClient()
	if err != nil {
		http.Error(w, "failed to create trip service client", http.StatusInternalServerError)
		return
	}
	defer tripService.Close()

	tripReview, err := tripService.Client.PreviewTrip(r.Context(), reqBody.ToProto())
	if err != nil {
		log.Printf("PreviewTrip failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := contracts.APIResponse{Data: tripReview}

	writeJson(w, http.StatusOK, response)
}

func handleTripStart(w http.ResponseWriter, r *http.Request) {

	var reqBody startTripRequest

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "failed to parse JSON data", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	tripService, err := grpc_client.NewTripServiceClient()
	if err != nil {
		http.Error(w, "failed to create trip service client", http.StatusInternalServerError)
		return
	}
	defer tripService.Close()

	tripStart, err := tripService.Client.CreateTrip(r.Context(), reqBody.ToProto())
	if err != nil {
		log.Printf("start trip failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := contracts.APIResponse{Data: tripStart}

	writeJson(w, http.StatusOK, response)

}
