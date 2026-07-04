package main

import (
	"context"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"time"
)

func main() {
	ctx := context.Background()
	inmemRepo := repository.NewInmemRepository()
	tripService := service.NewService(inmemRepo)

	t, err := tripService.CreateTrip(ctx, &domain.RideFareModel{
		UserID: "42",
	})

	if err != nil {
		log.Printf("err at start %s", err.Error())
	}

	log.Println("worked", t)

	//just for test
	for {
		time.Sleep(time.Second)
	}
}
