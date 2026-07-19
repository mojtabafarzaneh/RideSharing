package repository

import (
	"context"
	"fmt"
	"ride-sharing/services/trip-service/internal/domain"
	pbd "ride-sharing/shared/proto/driver"
	pb "ride-sharing/shared/proto/trip"
)

type inmemRepository struct {
	trips     map[string]*domain.TripModel
	rideFares map[string]*domain.RideFareModel
}

func NewInmemRepository() domain.TripRepository {
	return &inmemRepository{
		trips:     make(map[string]*domain.TripModel),
		rideFares: make(map[string]*domain.RideFareModel),
	}
}

func (i *inmemRepository) SaveTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	i.trips[trip.ID.Hex()] = trip
	return trip, nil
}

func (i *inmemRepository) SaveRideFare(ctx context.Context, fare *domain.RideFareModel) error {

	i.rideFares[fare.ID.Hex()] = fare

	return nil
}

func (i *inmemRepository) GetRideFareByID(ctx context.Context, id string) (*domain.RideFareModel, error) {
	res, exist := i.rideFares[id]
	if !exist {
		return nil, fmt.Errorf("could not find any fare with the provided id")
	}

	return res, nil

}

func (i *inmemRepository) GetTripById(ctx context.Context, id string) (*domain.TripModel, error) {
	trip, ok := i.trips[id]
	if !ok {
		return nil, nil
	}

	return trip, nil
}

func (i *inmemRepository) UpdateTrip(ctx context.Context, tripId string, status string, driver *pbd.Driver) error {
	trip, ok := i.trips[tripId]
	if !ok {
		return fmt.Errorf("trip not fount with Id: %s", tripId)
	}

	trip.Status = status

	if driver != nil {
		trip.Driver = &pb.TripDriver{
			Id:             driver.Id,
			Name:           driver.Name,
			ProfilePicture: driver.ProfilePicture,
			CarPlate:       driver.CarPlate,
		}
	}

	return nil
}
