package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_client"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"
	"ride-sharing/shared/proto/driver"
)

var (
	connManager = messaging.NewConnectionManager()
)

func handleRiderWebSoket(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	conn, err := connManager.Upgrade(w, r)
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
	connManager.Add(userID, conn)
	defer connManager.Remove(userID)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			http.Error(w, "failed to read message", http.StatusInternalServerError)
			break
		}

		log.Printf("Recieved message %v", base64.StdEncoding.EncodeToString(message))
	}

}

func handleDriverWebSoket(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	conn, err := connManager.Upgrade(w, r)
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
	connManager.Add(userID, conn)
	defer connManager.Remove(userID)

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

		connManager.Remove(userID)
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

	queues := []string{
		messaging.DriverCMDTripRequestQueue,
	}

	for _, q := range queues {
		consumer := messaging.NewQueueConsumer(rb, connManager, q)
		if err := consumer.Start(); err != nil {
			log.Printf("failed to start consuemr for queue: %s, err: %v", q, err)
		}
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			http.Error(w, "failed to read message", http.StatusInternalServerError)
			break
		}

		var wsMsg contracts.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Println("Failed to unmarshal incoming message:", err)
			continue

		}

		log.Printf("Recieved message %v", wsMsg)
	}

}
