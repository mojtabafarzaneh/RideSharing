package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_client"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/proto/driver"

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

	driverService, err := grpc_client.NewDriverServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	ctx := r.Context()

	data, err := driverService.Client.RegisterDriver(ctx, &driver.RegisterDriverRequest{
		PackageSlug: packageSlug,
		DriverID:    userID,
	})

	fmt.Println("driver when in websocket we get the data", data)

	if err != nil {
		log.Println(err)
	}

	defer driverService.Close()

	defer func() {
		driverService.Client.UnregisterDriver(ctx, &driver.RegisterDriverRequest{
			DriverID: userID,
		})
		driverService.Close()
		log.Println("Driver Unregistered: ", userID)
	}()
	fmt.Println("driver before the message been sent ", data)
	msg := contracts.WSMessage{
		Type: "driver.cmd.register",
		Data: data.Driver,
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
