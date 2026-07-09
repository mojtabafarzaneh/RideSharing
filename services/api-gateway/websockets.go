package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/util"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleRiderWebSoket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed to upgrade to websocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "userID is required", http.StatusBadRequest)
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			http.Error(w, "failed to read message", http.StatusInternalServerError)
			break
		}

		log.Printf("Recieved message %v", base64.StdEncoding.EncodeToString(message))
	}

}

func handleDriverWebSoket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "failed to upgrade to websocket", http.StatusInternalServerError)
		return
	}

	defer conn.Close()

	packageSlug := r.URL.Query().Get("packageSlug")
	if packageSlug == "" {
		http.Error(w, "packageSlug is required", http.StatusBadRequest)
		return
	}

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "userID is required", http.StatusBadRequest)
		return
	}

	type Driver struct {
		Id             string `json:"id"`
		Name           string `json:"name"`
		ProfilePicture string `json:"profilePicture"`
		PackageSlug    string `json:"packageSlug"`
		CarPlate       string `json:"carPlate"`
	}

	msg := contracts.WSMessage{
		Type: "driver.cmd.register",
		Data: Driver{
			Id:             userID,
			Name:           "Mojtaba",
			ProfilePicture: util.GetRandomAvatar(1),
			PackageSlug:    packageSlug,
			CarPlate:       "ABC-123",
		},
	}
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Error sending message: %v", err)
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			http.Error(w, "failed to read message", http.StatusInternalServerError)
			break
		}

		log.Printf("Recieved message %v", message)
	}

}
